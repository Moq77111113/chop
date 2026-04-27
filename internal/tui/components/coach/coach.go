// Package coach renders the first-run guidance overlay: three
// numbered green balloons stacked at the centre of the body. The
// bubbletea adapter handles the body dim.
package coach

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const widthMax = 50

// hint is one balloon's contents: a one-line headline and a secondary
// cue that names the relevant key.
type hint struct {
	title, cue string
}

// hints are the three first-run tips, ordered by the exploration path:
// pick a link → adjust a knob → escape the wall of perturbations.
var hints = []hint{
	{"this is the link you're poking", "↕ to pick another"},
	{"the focused knob — ↔ adjusts, tab cycles", ""},
	{"visible undo. always here.", "press any key to dismiss"},
}

// Props is empty — the copy is fixed.
type Props struct{}

// Styles bundles the balloon palette: chip background/foreground,
// title, cue, and the surrounding frame.
type Styles struct {
	Chip  lipgloss.Style
	Title lipgloss.Style
	Cue   lipgloss.Style
	Frame lipgloss.Style
}

// Render produces the centered balloon stack at width × height.
func Render(_ Props, s Styles, width, height int) string {
	balloonW := min(widthMax, max(width-4, 12))
	balloons := make([]string, 0, len(hints))
	for i, h := range hints {
		balloons = append(balloons, renderBalloon(i+1, h, s, balloonW))
	}
	stack := strings.Join(balloons, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}

// renderBalloon draws one balloon: numbered chip on the left, the hint
// title and optional cue stacked on the right. Background fills so the
// green call-out reads as a single solid shape.
func renderBalloon(n int, h hint, s Styles, width int) string {
	chip := s.Chip.Render(numberedPrefix(n))
	body := s.Title.Render(h.title)
	if h.cue != "" {
		body += "\n" + s.Cue.Render(h.cue)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, chip, " "+body)
	return s.Frame.Width(width).Render(row)
}

func numberedPrefix(n int) string {
	switch n {
	case 1:
		return "①"
	case 2:
		return "②"
	case 3:
		return "③"
	}
	return ""
}
