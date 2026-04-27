package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/intent"
	"github.com/moq77111113/chop/internal/tui/knobs"
)

const (
	tickInterval   = time.Second / 15
	snapshotBudget = 200 * time.Millisecond
	downRateLabel  = "— Mb/s"
)

type tickMsg struct{}

type rowsMsg struct {
	rows    []data.Row
	links   map[string]data.LinkSnapshot
	sources map[string]data.SourceSnapshot
}

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// fetchCmd snapshots every block in the registry and emits a rowsMsg
// that the bubbletea loop applies in applyRowsMsg.
func (a *App) fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), snapshotBudget)
		defer cancel()
		handles := a.sup.Registry.List()
		rows := make([]data.Row, 0, len(handles))
		links := map[string]data.LinkSnapshot{}
		sources := map[string]data.SourceSnapshot{}
		for _, h := range handles {
			snap, err := h.Snapshot(ctx)
			st := data.StateStarting
			if err == nil {
				st = data.MapState(snap.Status)
				decodeHandleSnapshot(h, snap, links, sources)
			}
			rows = append(rows, data.Row{ID: h.ID, Type: h.Type, State: st})
		}
		return rowsMsg{rows: rows, links: links, sources: sources}
	}
}

func decodeHandleSnapshot(h *supervisor.Handle, snap block.Snapshot, links map[string]data.LinkSnapshot, sources map[string]data.SourceSnapshot) {
	switch h.Type {
	case data.BlockTypeLink:
		if ls, ok := data.DecodeLink(snap); ok {
			links[h.ID] = ls
		}
	case data.BlockTypeSource:
		if ss, ok := data.DecodeSource(snap); ok {
			sources[h.ID] = ss
		}
	}
}

// applyRowsMsg lands a fresh registry snapshot: rates, history,
// cursor clamp, focused-pane sync.
func (a *App) applyRowsMsg(msg rowsMsg) {
	a.ui.FirstReady = true
	now := time.Now()
	rates := data.ComputeRates(a.links, msg.links, now.Sub(a.linksAt))
	a.links = msg.links
	a.sources = msg.sources
	a.linksAt = now
	for id, mbps := range rates {
		a.history.Push(id, mbps)
	}
	a.rows = a.enrichRows(msg.rows, rates)
	a.cursor.Clamp(len(a.rows))
	a.syncFocusedFromSelection()
}

// enrichRows decorates each row with title, rate label, sparkline
// samples, and knob value tokens.
func (a *App) enrichRows(rows []data.Row, rates map[string]float64) []data.Row {
	titles := data.LinkTitles(rows, a.configs)
	for i, r := range rows {
		if t, ok := titles[r.ID]; ok {
			rows[i].Title = t
		}
		switch r.State {
		case data.StateDown, data.StateStopped:
			rows[i].Rate = downRateLabel
		default:
			if mbps, ok := rates[r.ID]; ok {
				rows[i].Rate = fmt.Sprintf("%.1f Mb/s", mbps)
			}
		}
		if hist := a.history.For(r.ID); len(hist) > 0 {
			rows[i].Spark = hist
		}
		if ls, ok := a.links[r.ID]; ok {
			rows[i].Knobs = data.KnobValues(ls, r.State)
		}
	}
	return rows
}

// resetAllCmd PATCHes every link block with zeroed controls. The local
// knob pane is zeroed too so the next render doesn't show stale values
// until the snapshot tick catches up.
func (a *App) resetAllCmd() tea.Cmd {
	a.pane.ResetAll()
	ids := make([]string, 0, len(a.links))
	for id := range a.links {
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

// applyCmd ships the current knob state to the focused link's Apply
// RPC in a goroutine. The TUI doesn't wait for the reply — the next
// tick will read the resulting state.
func (a *App) applyCmd(s knobs.Snapshot) tea.Cmd {
	id := a.focusedLinkID()
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

// focusedLinkID returns the id of the link block under the cursor, or
// "" when the cursor selects a non-link block (source) or no row.
func (a *App) focusedLinkID() string {
	r, ok := a.selectedRow()
	if !ok || r.Type != data.BlockTypeLink {
		return ""
	}
	return r.ID
}

// syncFocusedFromSelection pushes the selected link's snapshot into
// the knob pane so the renderer reflects what the supervisor reports.
// For non-link selections the pane is left as-is (its values will be
// overwritten next time a link is focused).
func (a *App) syncFocusedFromSelection() {
	r, ok := a.selectedRow()
	if !ok || r.Type != data.BlockTypeLink {
		return
	}
	a.pane.SetSnapshot(focusedSnapshotFor(r, a.links))
}

func focusedSnapshotFor(r data.Row, links map[string]data.LinkSnapshot) knobs.Snapshot {
	up := r.State == data.StateUp || r.State == data.StateDegraded
	snap, ok := links[r.ID]
	if !ok {
		return knobs.Snapshot{LinkUp: up}
	}
	return knobs.Snapshot{
		Loss:          snap.Controls.Loss,
		LatencyMs:     snap.Controls.LatencyMs,
		JitterMs:      snap.Controls.JitterMs,
		BandwidthKbps: snap.Controls.BandwidthKbps,
		LinkUp:        up,
	}
}

func (a *App) intentSnapshot() intent.Snapshot {
	id := a.focusedLinkID()
	if id == "" {
		return intent.Snapshot{}
	}
	snap, ok := a.links[id]
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
