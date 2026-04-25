// Package focused renders the right-hand pane: link header, intent strip,
// four knobs (loss / latency / jitter / bandwidth), and a reset bar.
//
// The pane owns the knob focus index but does not handle input itself —
// the parent (tui.App) routes key events to its setters. Wiring an actual
// PATCH on a knob change is also the parent's job.
package focused

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/intent"
	"github.com/moq77111113/chop/internal/tui/knob"
)

// KnobIndex identifies one of the four impairment knobs by position.
type KnobIndex int

const (
	KnobLoss KnobIndex = iota
	KnobLatency
	KnobJitter
	KnobBandwidth
	knobCount = 4
)

// Model holds the focused link's identity and the four knob values.
type Model struct {
	LinkID   string
	LinkType string
	knobs    [knobCount]knob.Knob
	focused  KnobIndex
}

// Snapshot is the slim view of a link the focused pane renders. Decoupled
// from any block-package types so the TUI stays a leaf consumer.
type Snapshot struct {
	Loss          float64
	LatencyMs     uint32
	JitterMs      uint32
	BandwidthKbps uint32
	BitrateKbps   uint64
	LinkUp        bool
}

// New builds a Model with the canonical knob configuration spec'd in
// DESIGN.md §4.4.
func New() *Model {
	return &Model{knobs: defaultKnobs()}
}

// SetLink updates the focused link identity. Called whenever the parent
// changes selection in the link list.
func (m *Model) SetLink(id, typ string) {
	m.LinkID, m.LinkType = id, typ
}

// SetSnapshot pushes fresh values into the knobs. Called on every refresh
// tick so the knobs reflect the live state seen by the supervisor.
func (m *Model) SetSnapshot(s Snapshot) {
	m.knobs[KnobLoss].Value = s.Loss * lossScale
	m.knobs[KnobLatency].Value = float64(s.LatencyMs)
	m.knobs[KnobJitter].Value = float64(s.JitterMs)
	m.knobs[KnobBandwidth].Value = float64(s.BandwidthKbps) / kbpsToMbps
}

// Knob returns a copy of the knob at idx for inspection.
func (m *Model) Knob(idx KnobIndex) knob.Knob { return m.knobs[idx] }

// Focused returns the currently focused knob index.
func (m *Model) Focused() KnobIndex { return m.focused }

// NextKnob cycles forward through the four knobs.
func (m *Model) NextKnob() { m.focused = (m.focused + 1) % knobCount }

// PrevKnob cycles backward through the four knobs.
func (m *Model) PrevKnob() { m.focused = (m.focused + knobCount - 1) % knobCount }

// Adjust applies a small step in the requested direction to the focused
// knob and returns the resulting Snapshot the parent can ship as Controls.
func (m *Model) Adjust(direction int) Snapshot {
	k := m.knobs[m.focused]
	step := k.Step
	if direction < 0 {
		step = -step
	}
	m.knobs[m.focused] = k.Adjust(step)
	return m.toSnapshot()
}

// Zero resets the focused knob to its minimum and returns the resulting
// Snapshot.
func (m *Model) Zero() Snapshot {
	m.knobs[m.focused] = m.knobs[m.focused].Zero()
	return m.toSnapshot()
}

// ResetAll zeroes every knob.
func (m *Model) ResetAll() Snapshot {
	for i := range m.knobs {
		m.knobs[i] = m.knobs[i].Zero()
	}
	return m.toSnapshot()
}

func (m *Model) toSnapshot() Snapshot {
	return Snapshot{
		Loss:          m.knobs[KnobLoss].Value / lossScale,
		LatencyMs:     uint32(m.knobs[KnobLatency].Value),
		JitterMs:      uint32(m.knobs[KnobJitter].Value),
		BandwidthKbps: uint32(m.knobs[KnobBandwidth].Value * kbpsToMbps),
	}
}

// Styles bundles the lipgloss styles the focused pane and its children need.
type Styles struct {
	Header     lipgloss.Style
	Subtle     lipgloss.Style
	Knob       knob.Styles
	Intent     intent.Styles
	Divider    lipgloss.Style
	ResetFrame lipgloss.Style
	ResetLabel lipgloss.Style
	ResetKey   lipgloss.Style
}

// Render returns the full pane content sized to width × height.
func (m *Model) Render(s Styles, width, height int, snap intent.Snapshot) string {
	if m.LinkID == "" {
		return s.Subtle.Render("no link selected")
	}
	header := s.Header.Render(strings.ToUpper(m.LinkID))
	if m.LinkType != "" {
		header += " " + s.Subtle.Render("("+m.LinkType+")")
	}
	parts := []string{
		header,
		"",
		intent.Render(snap, s.Intent, width),
		"",
	}
	for i, k := range m.knobs {
		parts = append(parts, knob.Render(k, s.Knob, width, KnobIndex(i) == m.focused))
		parts = append(parts, "")
	}
	parts = append(parts, renderResetBar(s, width))
	out := strings.Join(parts, "\n")
	return clamp(out, height)
}

func renderResetBar(s Styles, width int) string {
	innerW := max(width-4, 1)
	label := s.ResetLabel.Render("undo all perturbations on this link")
	keys := s.ResetKey.Render("reset r") + " " + s.ResetKey.Render("R all")
	gap := max(innerW-lipgloss.Width(label)-lipgloss.Width(keys), 1)
	body := label + strings.Repeat(" ", gap) + keys
	return s.ResetFrame.Width(innerW).Render(body)
}

func clamp(view string, height int) string {
	if height <= 0 {
		return view
	}
	lines := strings.Split(view, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
