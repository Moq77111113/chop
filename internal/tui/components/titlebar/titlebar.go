// Package titlebar renders the top chrome row: pulsing dot, "chop"
// wordmark, IPA, registry summary on the left; consumable URL,
// wall-clock time, version stamp on the right.
package titlebar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	separator = " · "
	ipa       = "/tʃɒp/"
)

// Props is the input bag.
type Props struct {
	// Summary on the left of the bar — e.g. "3 links · running" or "idle".
	Summary string
	// ConsumableURL is the rtsp endpoint a downstream consumer would hit
	// for the focused link. Empty string omits the URL from the right
	// cluster.
	ConsumableURL string
	// Time is the formatted wall-clock string (e.g. "02:14:38").
	Time string
	// Version is the build stamp (e.g. "v0.4.1-dev" or "dev").
	Version string
	// Narrow suppresses the right cluster entirely so the wordmark gets
	// the full row.
	Narrow bool
}

// Styles bundles the bar palette.
type Styles struct {
	Bar    lipgloss.Style
	Dot    lipgloss.Style
	Name   lipgloss.Style
	IPA    lipgloss.Style
	Subtle lipgloss.Style
}

// Render assembles the bar at exactly width cells.
func Render(p Props, s Styles, width int) string {
	left := renderLeft(p, s)
	right := renderRight(p, s)
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right)-2, 1)
	bar := left + strings.Repeat(" ", gap) + right
	return s.Bar.Width(width).Padding(0, 1).Render(bar)
}

func renderLeft(p Props, s Styles) string {
	parts := []string{
		s.Dot.Render("●") + " " + s.Name.Render("chop"),
		s.IPA.Render(ipa),
	}
	if !p.Narrow && p.Summary != "" {
		parts = append(parts, s.Subtle.Render(p.Summary))
	}
	return strings.Join(parts, separator)
}

func renderRight(p Props, s Styles) string {
	if p.Narrow {
		return ""
	}
	parts := make([]string, 0, 3)
	if p.ConsumableURL != "" {
		parts = append(parts, p.ConsumableURL)
	}
	if p.Time != "" {
		parts = append(parts, p.Time)
	}
	if p.Version != "" {
		parts = append(parts, p.Version)
	}
	if len(parts) == 0 {
		return ""
	}
	return s.Subtle.Render(strings.Join(parts, separator))
}
