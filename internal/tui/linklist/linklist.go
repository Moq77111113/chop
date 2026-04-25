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

// Row is one line in the link list. Knob summary and sparkline arrive in
// later steps; for now a row is identity + state.
type Row struct {
	ID    string
	Type  string
	State State
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
		body := s.Empty.Render("no blocks · waiting on supervisor")
		return clamp(lipgloss.JoinVertical(lipgloss.Left, header, body), width, height)
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
	left := cursor + " " + id + typ
	gap := strings.Repeat(" ", max(width-lipgloss.Width(left)-lipgloss.Width(state), 1))
	row := left + gap + state
	if focused {
		return s.Selected.Render(row)
	}
	return row
}

func clamp(view string, _, height int) string {
	lines := strings.Split(view, "\n")
	if len(lines) > height && height > 0 {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
