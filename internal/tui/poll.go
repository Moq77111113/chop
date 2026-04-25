package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/focused"
	"github.com/moq77111113/chop/internal/tui/intent"
	"github.com/moq77111113/chop/internal/tui/linklist"
)

const (
	tickInterval   = time.Second / 15
	snapshotBudget = 200 * time.Millisecond

	blockTypeLink   = "link"
	blockTypeSource = "source"

	avgRTPPacketBytes = 1500.0
	bitsPerByte       = 8
	bitsPerMb         = 1_000_000.0
	maxRateInterval   = 5 * time.Second
)

type tickMsg struct{}

type rowsMsg struct {
	rows    []linklist.Row
	links   map[string]linkSnapshot
	sources map[string]sourceSnapshot
}

// linkSnapshot is the TUI-local projection of a link block's snapshot
// stats. Unmarshaled from the wire JSON so the TUI doesn't import the
// block packages directly.
type linkSnapshot struct {
	Status     block.Status
	PacketsIn  uint64 `json:"packets_in"`
	PacketsOut uint64 `json:"packets_out"`
	UpSinceMs  int64  `json:"up_since_ms"`
	Controls   struct {
		Loss          float64 `json:"loss"`
		LatencyMs     uint32  `json:"latency_ms"`
		JitterMs      uint32  `json:"jitter_ms"`
		BandwidthKbps uint32  `json:"bandwidth_kbps"`
	} `json:"controls"`
}

// sourceSnapshot is the TUI-local projection of a source block's snapshot.
type sourceSnapshot struct {
	RTPServed int64 `json:"rtp_served"`
	UpSinceMs int64 `json:"up_since_ms"`
}

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// fetchCmd snapshots every block in the registry and emits a rowsMsg.
func (a *App) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), snapshotBudget)
		defer cancel()
		handles := a.sup.Registry.List()
		rows := make([]linklist.Row, 0, len(handles))
		links := make(map[string]linkSnapshot)
		sources := make(map[string]sourceSnapshot)
		for _, h := range handles {
			snap, err := h.Snapshot(ctx)
			state := linklist.StateStarting
			if err == nil {
				state = mapState(snap.Status)
				decodeSnapshot(h, snap, links, sources)
			}
			rows = append(rows, linklist.Row{ID: h.ID, Type: h.Type, State: state})
		}
		return rowsMsg{rows: rows, links: links, sources: sources}
	}
}

func (a *App) applyRowsMsg(msg rowsMsg) {
	a.ui.firstReady = true
	now := time.Now()
	rates := computeRates(a.data.links, msg.links, now.Sub(a.data.linksAt))
	a.data.links = msg.links
	a.data.sources = msg.sources
	a.data.linksAt = now
	for id, mbps := range rates {
		a.data.history[id] = pushSample(a.data.history[id], mbps)
	}
	a.list.Set(a.enrichRows(msg.rows, rates))
	a.syncFocusedFromList()
}

// enrichRows decorates each row with its formatted Mb/s rate, the
// sparkline of recent samples, and the topology title (`src → link`).
// Rows without a known rate yet (first tick, non-link blocks) keep the
// fields empty and render single-line. Down/stopped links render their
// rate as `— Mb/s` so the column reads as "no signal" rather than blank.
// The knobs summary is pre-styled here (labels muted, values fg/danger)
// so the linklist renderer can drop it in raw without re-applying a dim
// blanket that would erase the per-token emphasis.
func (a *App) enrichRows(rows []linklist.Row, rates map[string]float64) []linklist.Row {
	titles := a.linkTitles(rows)
	for i, r := range rows {
		if t, ok := titles[r.ID]; ok {
			rows[i].Title = t
		}
		switch r.State {
		case linklist.StateDown, linklist.StateStopped:
			rows[i].Rate = downRateLabel
		default:
			if mbps, ok := rates[r.ID]; ok {
				rows[i].Rate = fmt.Sprintf("%.1f Mb/s", mbps)
			}
		}
		if hist, hasHist := a.data.history[r.ID]; hasHist {
			rows[i].Sparkline = renderSparkline(hist)
		}
		if ls, ok := a.data.links[r.ID]; ok {
			rows[i].KnobsSummary = a.styleKnobsSummary(ls, r.State)
		}
	}
	return rows
}

const downRateLabel = "— Mb/s"

// linkTitles maps each link block id to a "<source-id> → <link-id>"
// label by matching its upstream URL host:port against the listen
// host:port of source blocks. Links with no matching source keep their
// id as the title (handled by the renderer's fallback).
func (a *App) linkTitles(rows []linklist.Row) map[string]string {
	sourceByListen := map[string]string{}
	for _, r := range rows {
		if r.Type != blockTypeSource {
			continue
		}
		if listen := parseSourceListen(a.configs[r.ID]); listen != "" {
			sourceByListen[listen] = r.ID
		}
	}
	titles := map[string]string{}
	for _, r := range rows {
		if r.Type != blockTypeLink {
			continue
		}
		host := parseLinkUpstreamHost(a.configs[r.ID])
		if src, ok := sourceByListen[host]; ok {
			titles[r.ID] = src + " → " + r.ID
		}
	}
	return titles
}

func parseSourceListen(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var sc struct {
		Listen string `json:"listen"`
	}
	if json.Unmarshal(raw, &sc) != nil {
		return ""
	}
	return sc.Listen
}

func parseLinkUpstreamHost(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var lc struct {
		Upstream string `json:"upstream"`
	}
	if json.Unmarshal(raw, &lc) != nil || lc.Upstream == "" {
		return ""
	}
	u, err := url.Parse(lc.Upstream)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}

func decodeSnapshot(h *supervisor.Handle, snap block.Snapshot, links map[string]linkSnapshot, sources map[string]sourceSnapshot) {
	switch h.Type {
	case blockTypeLink:
		var ls linkSnapshot
		if json.Unmarshal(snap.Stats, &ls) == nil {
			ls.Status = snap.Status
			links[h.ID] = ls
		}
	case blockTypeSource:
		var ss sourceSnapshot
		if json.Unmarshal(snap.Stats, &ss) == nil {
			sources[h.ID] = ss
		}
	}
}

func mapState(s block.Status) linklist.State {
	switch s {
	case block.StatusRunning:
		return linklist.StateUp
	case block.StatusDegraded:
		return linklist.StateDegraded
	case block.StatusStopped:
		return linklist.StateStopped
	}
	return linklist.StateStarting
}

// computeRates derives a per-link Mb/s readout from two consecutive
// snapshot samples. dt outside (0, maxRateInterval] returns no rates so a
// stale previous sample doesn't produce a misleading number.
func computeRates(prev, cur map[string]linkSnapshot, dt time.Duration) map[string]float64 {
	if dt <= 0 || dt > maxRateInterval {
		return nil
	}
	rates := make(map[string]float64, len(cur))
	for id, c := range cur {
		p, ok := prev[id]
		if !ok || c.PacketsOut < p.PacketsOut {
			continue
		}
		pps := float64(c.PacketsOut-p.PacketsOut) / dt.Seconds()
		rates[id] = pps * avgRTPPacketBytes * bitsPerByte / bitsPerMb
	}
	return rates
}

// knobValues holds the four per-knob value tokens as they should appear
// in the row sub-line (e.g. `12%`, `200ms`, `±45ms`, `1.5M`). Splitting
// the format from the styling lets us drive both a plain-text test path
// and the lipgloss-styled render path off the same source of truth.
type knobValues struct {
	loss, latency, jitter, bandwidth string
	bandwidthOff                     bool // bw == ∞ or — → render dim
}

func computeKnobValues(s linkSnapshot, st linklist.State) knobValues {
	const (
		percentScale = 100
		kbpsToMbps   = 1000
	)
	v := knobValues{
		loss:    fmt.Sprintf("%d%%", int(s.Controls.Loss*percentScale+0.5)),
		latency: fmt.Sprintf("%dms", s.Controls.LatencyMs),
		jitter:  fmt.Sprintf("±%dms", s.Controls.JitterMs),
	}
	switch {
	case st == linklist.StateDown || st == linklist.StateStopped:
		v.bandwidth = "—"
		v.bandwidthOff = true
	case s.Controls.BandwidthKbps == 0:
		v.bandwidth = "∞"
		v.bandwidthOff = true
	default:
		v.bandwidth = fmt.Sprintf("%.1fM", float64(s.Controls.BandwidthKbps)/kbpsToMbps)
	}
	return v
}

// styleKnobsSummary builds the per-row "loss 12% lat 200ms jit ±45ms bw 1.5M"
// strip with labels in Muted and each value styled by salience: zero values
// stay Dim, non-zero values pop to Fg, everything goes Danger when the link
// is down. The result is pre-rendered ANSI so the linklist renderer must
// not re-wrap it in a blanket Foreground style.
func (a *App) styleKnobsSummary(s linkSnapshot, st linklist.State) string {
	v := computeKnobValues(s, st)
	label := lipgloss.NewStyle().Foreground(a.theme.Muted)
	dim := lipgloss.NewStyle().Foreground(a.theme.Dim)
	bright := lipgloss.NewStyle().Foreground(a.theme.Fg).Bold(true)
	danger := lipgloss.NewStyle().Foreground(a.theme.Danger).Bold(true)

	down := st == linklist.StateDown || st == linklist.StateStopped
	val := func(token string, zero bool) string {
		switch {
		case down:
			return danger.Render(token)
		case zero:
			return dim.Render(token)
		}
		return bright.Render(token)
	}

	return strings.Join([]string{
		label.Render("loss") + " " + val(v.loss, s.Controls.Loss == 0),
		label.Render("lat") + " " + val(v.latency, s.Controls.LatencyMs == 0),
		label.Render("jit") + " " + val(v.jitter, s.Controls.JitterMs == 0),
		label.Render("bw") + " " + val(v.bandwidth, v.bandwidthOff && !down),
	}, " ")
}

// resetAllCmd PATCHes every link block with zeroed controls. The local
// focused-pane knobs are zeroed too so the next render doesn't show stale
// values until the snapshot tick catches up.
func (a *App) resetAllCmd() tea.Cmd {
	a.focused.ResetAll()
	ids := make([]string, 0, len(a.data.links))
	for id := range a.data.links {
		ids = append(ids, id)
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), snapshotBudget)
		defer cancel()
		body := json.RawMessage(`{"loss":0,"latency_ms":0,"jitter_ms":0,"bandwidth_kbps":0}`)
		for _, id := range ids {
			if h := a.sup.Registry.Get(id); h != nil {
				_ = h.Apply(ctx, body)
			}
		}
		return nil
	}
}

// applyCmd ships the current knob state to the focused link's Apply RPC
// in a goroutine. The TUI doesn't wait for the reply — the next tick will
// read the resulting state.
func (a *App) applyCmd(s focused.Snapshot) tea.Cmd {
	id := a.focused.LinkID
	if id == "" {
		return nil
	}
	return func() tea.Msg {
		h := a.sup.Registry.Get(id)
		if h == nil {
			return nil
		}
		body, err := json.Marshal(struct {
			Loss          float64 `json:"loss"`
			LatencyMs     uint32  `json:"latency_ms"`
			JitterMs      uint32  `json:"jitter_ms"`
			BandwidthKbps uint32  `json:"bandwidth_kbps"`
		}{s.Loss, s.LatencyMs, s.JitterMs, s.BandwidthKbps})
		if err != nil {
			return nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), snapshotBudget)
		defer cancel()
		_ = h.Apply(ctx, body)
		return nil
	}
}

// syncFocusedFromList aligns the focused pane with the currently-selected
// row from the link list. It refreshes the knob values from the most
// recent snapshot so values reflect what the supervisor reports. For
// non-link blocks (sources) the knobs are zeroed.
func (a *App) syncFocusedFromList() {
	rows := a.list.Rows()
	idx := a.list.Selected()
	if idx < 0 || idx >= len(rows) {
		a.focused.SetLink("", "")
		return
	}
	r := rows[idx]
	a.focused.SetLink(r.ID, r.Type)
	a.focused.SetSnapshot(a.focusedSnapshotFor(r))
}

func (a *App) focusedSnapshotFor(r linklist.Row) focused.Snapshot {
	up := r.State == linklist.StateUp || r.State == linklist.StateDegraded
	if r.Type != blockTypeLink {
		return focused.Snapshot{LinkUp: up}
	}
	snap, ok := a.data.links[r.ID]
	if !ok {
		return focused.Snapshot{LinkUp: up}
	}
	return focused.Snapshot{
		Loss:          snap.Controls.Loss,
		LatencyMs:     snap.Controls.LatencyMs,
		JitterMs:      snap.Controls.JitterMs,
		BandwidthKbps: snap.Controls.BandwidthKbps,
		LinkUp:        up,
	}
}

func (a *App) intentSnapshot() intent.Snapshot {
	id := a.focused.LinkID
	if id == "" {
		return intent.Snapshot{}
	}
	snap, ok := a.data.links[id]
	if !ok {
		return intent.Snapshot{LinkUp: true}
	}
	return intent.Snapshot{
		Loss:          snap.Controls.Loss,
		LatencyMs:     snap.Controls.LatencyMs,
		JitterMs:      snap.Controls.JitterMs,
		BandwidthKbps: snap.Controls.BandwidthKbps,
		LinkUp:        snap.Status != block.StatusStopped,
		HasController: true,
	}
}
