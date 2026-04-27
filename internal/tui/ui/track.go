package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	trackFillChar  = "█"
	trackEmptyChar = "░"
)

// TrackVariant selects the fill tint by severity.
type TrackVariant int

// Track variants — ample / warn / danger.
const (
	TrackOK TrackVariant = iota
	TrackWarn
	TrackDanger
)

// TrackStyles bundles the palette: per-variant fill, empty section,
// focus rail glyph.
type TrackStyles struct {
	Fill     [3]lipgloss.Style
	Empty    lipgloss.Style
	FocusBar lipgloss.Style
}

// Track renders a slider track of total `width` cells. The leading
// focus-rail glyph (or space placeholder) is part of the budget — the
// bar itself uses width-1 cells so the whole row fits exactly.
func Track(ratio float64, width int, variant TrackVariant, focused bool, s TrackStyles) string {
	barW := max(width-1, 1)
	ratio = clampUnit(ratio)
	fillCount := max(0, min(barW, int(ratio*float64(barW)+0.5)))
	fill := s.Fill[variant].Render(strings.Repeat(trackFillChar, fillCount))
	empty := s.Empty.Render(strings.Repeat(trackEmptyChar, barW-fillCount))
	bar := fill + empty
	if focused {
		return s.FocusBar.Render("▌") + bar
	}
	return " " + bar
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
