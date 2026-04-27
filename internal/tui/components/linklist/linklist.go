// Package linklist renders the left-pane list: header on top followed
// by one card per row. Empty-state and the cursor live elsewhere
// (components/empty and state.Cursor on the bubbletea adapter).
package linklist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/components/linkcard"
	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/ui"
)

// Props carries the rows and the selected index from state.Cursor.
// Selected == -1 means no selection (empty list); the caller is
// expected to delegate to components/empty before reaching here.
type Props struct {
	Rows     []data.Row
	Selected int
}

// Styles bundles the header style and the link-card palette.
type Styles struct {
	Header lipgloss.Style
	Card   linkcard.Styles
}

// Render produces the pane at width × height.
func Render(p Props, s Styles, width, height int) string {
	header := ui.Header("LINKS", s.Header)
	lines := []string{header}
	for i, r := range p.Rows {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, linkcard.Render(linkcard.Props{
			Row:     r,
			Focused: i == p.Selected,
		}, s.Card, width))
	}
	return clamp(strings.Join(lines, "\n"), height)
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
