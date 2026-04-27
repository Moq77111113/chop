// Package statusbar renders the bottom chrome row — keyboard hints as
// bordered key chips matching design/screens.html.
package statusbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/ui"
)

const separator = "  "

// Hint is one (chip, label) pair.
type Hint struct {
	Key, Label string
}

// Props is the input bag. Body, when non-empty, replaces the hint
// strip entirely — used for transient prompts like "reset N block(s)?
// [y/n]".
type Props struct {
	Body      string
	Hints     []Hint
	RightSlot string // free-form right-aligned text, e.g. "9.1 MB · 0 alloc/s"
}

// Styles bundles the kbd chip palette and the surrounding bar style.
type Styles struct {
	Bar   lipgloss.Style
	Body  lipgloss.Style
	Right lipgloss.Style
	Kbd   ui.KbdStyles
}

// Render assembles the bottom row at exactly width cells.
func Render(p Props, s Styles, width int) string {
	left := p.Body
	if left == "" {
		left = renderHints(p.Hints, s)
	} else {
		left = s.Body.Render(left)
	}
	right := ""
	if p.RightSlot != "" {
		right = s.Right.Render(p.RightSlot)
	}
	if right == "" {
		return s.Bar.Width(width).Padding(0, 1).Render(left)
	}
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right)-2, 1)
	row := left + strings.Repeat(" ", gap) + right
	return s.Bar.Width(width).Padding(0, 1).Render(row)
}

func renderHints(hints []Hint, s Styles) string {
	tokens := make([]string, 0, len(hints))
	for _, h := range hints {
		token := ui.Kbd(h.Key, s.Kbd) + " " + s.Body.Render(h.Label)
		tokens = append(tokens, token)
	}
	return strings.Join(tokens, separator)
}
