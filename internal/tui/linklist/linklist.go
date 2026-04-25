// Package linklist renders the left-pane list of running blocks. The data
// model is a flat row slice; the parent (tui.App) refreshes it on each tick
// from the supervisor registry.
package linklist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// State is the UI-side projection of block.Status. We keep a TUI-local
// enum so the linklist doesn't depend on the block package directly.
type State string

const (
	StateUp       State = "UP"
	StateDegraded State = "DEGRADED"
	StateDown     State = "DOWN"
	StateStarting State = "STARTING"
	StateStopped  State = "STOPPED"
)

// Row is one line in the link list. KnobsSummary, when non-empty, prints
// on a second indented line below the identity row and turns the row into
// a two-line cell. Source rows leave it empty. Sparkline + Rate are shown
// right-aligned next to the state pill on line 1.
type Row struct {
	ID           string
	Type         string
	State        State
	KnobsSummary string
	Sparkline    string
	Rate         string
}

// Model holds the rendered rows and the cursor.
type Model struct {
	rows     []Row
	selected int
}

// Set replaces the row slice and clamps the selection to a valid index.
// The order of rows is preserved across refreshes; clamping is enough to
// keep the cursor coherent when blocks come and go.
func (m *Model) Set(rows []Row) {
	m.rows = rows
	if m.selected >= len(rows) {
		m.selected = max(0, len(rows)-1)
	}
}

// Rows exposes the underlying rows for outside readers (focused row lookup).
func (m *Model) Rows() []Row { return m.rows }

// Selected returns the index of the focused row, or -1 if the list is empty.
func (m *Model) Selected() int {
	if len(m.rows) == 0 {
		return -1
	}
	return m.selected
}

// MoveUp moves the cursor one row up, clamped to 0.
func (m *Model) MoveUp() {
	if m.selected > 0 {
		m.selected--
	}
}

// MoveDown moves the cursor one row down, clamped to the last row.
func (m *Model) MoveDown() {
	if m.selected < len(m.rows)-1 {
		m.selected++
	}
}

// Render returns the full pane content sized to width × height.
func (m *Model) Render(s Styles, width, height int) string {
	header := s.Header.Render("LINKS")
	if len(m.rows) == 0 {
		return clamp(lipgloss.JoinVertical(lipgloss.Left, header, s.Empty.Render("no blocks · waiting on supervisor")), width, height)
	}

	lines := make([]string, 0, len(m.rows)+1)
	lines = append(lines, header)
	for i, r := range m.rows {
		lines = append(lines, m.renderRow(s, r, i == m.selected, width))
	}
	return clamp(lipgloss.JoinVertical(lipgloss.Left, lines...), width, height)
}

func (m *Model) renderRow(s Styles, r Row, focused bool, width int) string {
	cursor := " "
	if focused {
		cursor = "▸"
	}
	id := r.ID + " "
	typ := s.Type.Render("(" + r.Type + ")")
	state := s.statePill(r.State)
	spark := ""
	if r.Sparkline != "" {
		spark = " " + s.Spark.Render(r.Sparkline)
	}
	rate := ""
	if r.Rate != "" {
		rate = " " + s.Summary.Render(r.Rate)
	}
	left := cursor + " " + id + typ
	right := state + spark + rate
	gap := strings.Repeat(" ", max(width-lipgloss.Width(left)-lipgloss.Width(right), 1))
	identity := left + gap + right
	if focused {
		identity = s.Selected.Render(identity)
	}
	if r.KnobsSummary == "" {
		return identity
	}
	summary := s.Summary.Render("  " + r.KnobsSummary)
	return identity + "\n" + summary
}

func clamp(view string, _, height int) string {
	lines := strings.Split(view, "\n")
	if len(lines) > height && height > 0 {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
