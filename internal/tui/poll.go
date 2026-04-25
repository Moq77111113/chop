package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
			row := linklist.Row{ID: h.ID, Type: h.Type, State: state}
			if ls, ok := links[h.ID]; ok {
				row.KnobsSummary = formatKnobsSummary(ls)
			}
			rows = append(rows, row)
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

// enrichRows decorates each row with its formatted Mb/s rate and the
// sparkline of recent samples. Rows without a known rate yet (first tick,
// non-link blocks) keep the fields empty and render single-line.
func (a *App) enrichRows(rows []linklist.Row, rates map[string]float64) []linklist.Row {
	for i, r := range rows {
		mbps, ok := rates[r.ID]
		if ok {
			rows[i].Rate = fmt.Sprintf("%.1f Mb/s", mbps)
		}
		if hist, hasHist := a.data.history[r.ID]; hasHist {
			rows[i].Sparkline = renderSparkline(hist)
		}
	}
	return rows
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

// formatKnobsSummary builds the second-line dim summary for a link row,
// e.g. `loss 12% lat 200ms jit ±45ms bw 1.5M`. Bandwidth 0 prints as `∞`
// to match the bandwidth knob's "off" semantics.
func formatKnobsSummary(s linkSnapshot) string {
	const (
		percentScale = 100
		kbpsToMbps   = 1000
	)
	bw := "∞"
	if s.Controls.BandwidthKbps > 0 {
		bw = fmt.Sprintf("%.1fM", float64(s.Controls.BandwidthKbps)/kbpsToMbps)
	}
	return fmt.Sprintf("loss %d%% lat %dms jit ±%dms bw %s",
		int(s.Controls.Loss*percentScale+0.5),
		s.Controls.LatencyMs,
		s.Controls.JitterMs,
		bw,
	)
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
