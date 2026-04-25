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

// Row is one cell in the link list. Link rows render two lines: bold
// title (ID or "src → dst" if Title is set), then an indented metrics
// line with the state pill, sparkline and rate right-aligned. Source
// rows render as a single title line.
type Row struct {
	ID           string
	Type         string
	State        State
	Title        string
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
		return clamp(lipgloss.JoinVertical(lipgloss.Left, header, renderEmpty(s, width, height-emptyHeaderRows)), width, height)
	}

	lines := make([]string, 0, len(m.rows)+1)
	lines = append(lines, header)
	for i, r := range m.rows {
		lines = append(lines, m.renderRow(s, r, i == m.selected, width))
	}
	return clamp(strings.Join(lines, "\n"), width, height)
}

// emptyHeaderRows is the vertical budget the LINKS header eats above the
// empty-state card so the centering math sees the actual usable height.
const emptyHeaderRows = 2

// emptyArt is the stylised "no input" silhouette shown when no blocks are
// running. Plain runes — no ANSI — so EmptyArt can colour it as a whole.
var emptyArt = []string{
	"    .---.",
	"    |   |",
	".---'---'---.",
	"|  no input  |",
	"'------------'",
}

// emptyExample is a single suggested command rendered as a small framed
// card. The split (cmd / arg) lets the renderer paint the user-supplied
// portion in the primary accent so the call-to-action reads at a glance.
type emptyExample struct{ cmd, arg string }

var emptyExamples = []emptyExample{
	{"chop run", "cam.mp4"},
	{"chop run", "rtsp://192.168.1.10:554/live"},
}

const emptyHintCopy = "no source attached. point chop at an rtsp url or a file:"

// renderEmpty paints the centered illustration + hint + example cards
// shown when the registry has no blocks yet. Width/height are the pane's
// inner dimensions; the result is vertically centered within them.
func renderEmpty(s Styles, width, height int) string {
	art := s.EmptyArt.Render(strings.Join(emptyArt, "\n"))
	hint := s.EmptyText.Render(emptyHintCopy)
	cards := make([]string, 0, len(emptyExamples))
	for _, ex := range emptyExamples {
		body := s.ExamplePrompt.Render("$") + " " +
			s.ExampleCommand.Render(ex.cmd) + " " +
			s.ExampleArg.Render(ex.arg)
		cards = append(cards, s.ExampleFrame.Render(body))
	}
	stack := lipgloss.JoinVertical(lipgloss.Center, art, "", hint, "", cards[0], cards[1])
	if height <= 0 {
		return lipgloss.PlaceHorizontal(width, lipgloss.Center, stack)
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}

// renderRow draws one row of the link list following the design layout:
//
//	line 1: cursor + title (left)            state · spark · rate (right)
//	line 2: metrics summary (only for link rows with KnobsSummary)
//
// The focused row is wrapped in a rounded primary-colored border; other
// rows render as plain indented text with no border.
func (m *Model) renderRow(s Styles, r Row, focused bool, width int) string {
	title := r.Title
	if title == "" {
		title = r.ID
	}
	titleStyle := s.Title
	cursor := "  "
	if focused {
		titleStyle = s.Selected
		cursor = s.Selected.Render("▸") + " "
	}
	titleRendered := cursor + titleStyle.Render(title)
	rightCluster := s.statePill(r.State)
	if r.Sparkline != "" {
		rightCluster += " " + s.Spark.Render(r.Sparkline)
	}
	if r.Rate != "" {
		rightCluster += " " + s.Summary.Render(r.Rate)
	}

	inner := max(width-rowFrameOverhead, 10)
	gap := strings.Repeat(" ", max(inner-lipgloss.Width(titleRendered)-lipgloss.Width(rightCluster), 1))
	line1 := titleRendered + gap + rightCluster
	body := line1
	if r.KnobsSummary != "" {
		// KnobsSummary arrives pre-styled (labels muted, values fg/danger);
		// wrapping it in s.Summary would flatten the per-token emphasis.
		body += "\n  " + r.KnobsSummary
	}
	if focused {
		return s.RowFrameSelected.Render(body)
	}
	return s.RowFrame.Render(body)
}

// rowFrameOverhead is the visual budget the row's border + padding eats
// out of the pane width: 2 border chars + 2 padding(0,1) chars. Drives
// the gap-fill calc in renderRow; the frame itself sizes to content so
// selected and unselected rows render at the same outer width (no jitter
// when toggling selection).
const rowFrameOverhead = 4

func clamp(view string, _, height int) string {
	lines := strings.Split(view, "\n")
	if len(lines) > height && height > 0 {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
