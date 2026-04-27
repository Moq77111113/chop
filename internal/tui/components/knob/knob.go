// Package knob renders one slider row: header (label / hint / value),
// track, scale.
package knob

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/knobs"
	"github.com/moq77111113/chop/internal/tui/ui"
)

const (
	hintText    = "←→ adjust"
	minRowWidth = 20
)

// Styles bundles the knob palette plus the underlying track palette.
type Styles struct {
	Label   lipgloss.Style
	LabelOn lipgloss.Style
	Hint    lipgloss.Style
	Value   [3]lipgloss.Style // indexed by knobs.Severity
	Scale   lipgloss.Style
	Track   ui.TrackStyles
}

// Render draws a 3-line knob (label/value, track, scale) sized to
// width. focused swaps in the focused styles and shows the arrow hint.
func Render(k knobs.Knob, s Styles, width int, focused bool) string {
	if width < minRowWidth {
		width = minRowWidth
	}
	sev := knobs.SevOK
	if k.Severity != nil {
		sev = k.Severity(k.Value)
	}
	header := renderHeader(k, s, width, focused, sev)
	ratio := 0.0
	if k.Max > k.Min {
		ratio = (k.Value - k.Min) / (k.Max - k.Min)
	}
	track := ui.Track(ratio, width, ui.TrackVariant(sev), focused, s.Track)
	scale := renderScale(k, s, width)
	return strings.Join([]string{header, track, scale}, "\n")
}

func renderHeader(k knobs.Knob, s Styles, width int, focused bool, sev knobs.Severity) string {
	labelStyle := s.Label
	if focused {
		labelStyle = s.LabelOn
	}
	label := labelStyle.Render(strings.ToUpper(k.Label))
	val := s.Value[sev].Render(k.Format(k.Value))
	hint := ""
	if focused {
		hint = s.Hint.Render(hintText)
	}
	gapL := max(width-lipgloss.Width(label)-lipgloss.Width(hint)-lipgloss.Width(val), 1)
	if hint == "" {
		return label + strings.Repeat(" ", gapL) + val
	}
	gapR := gapL / 2
	gapM := gapL - gapR
	return label + strings.Repeat(" ", gapM) + hint + strings.Repeat(" ", gapR) + val
}

func renderScale(k knobs.Knob, s Styles, width int) string {
	if k.Max <= k.Min || width <= 0 {
		return ""
	}
	labels := k.ScaleLabels
	if len(labels) == 0 {
		labels = autoLabels(k)
	}
	if !labelsFit(labels, width) {
		return ""
	}
	return s.Scale.Render(" " + placeLabels(labels, width))
}

func autoLabels(k knobs.Knob) []string {
	ticks := []float64{k.Min, k.Min + (k.Max-k.Min)*0.5, k.Max}
	out := make([]string, len(ticks))
	for i, t := range ticks {
		out[i] = k.Format(t)
	}
	return out
}

func labelsFit(labels []string, width int) bool {
	total := 0
	for _, l := range labels {
		total += lipgloss.Width(l)
	}
	return total+max(len(labels)-1, 0) <= width
}

// placeLabels distributes labels across width, anchored at proportional
// positions, then concatenates. Multi-byte runes (±, ∞, /) keep their
// visual width thanks to lipgloss.Width — no byte-by-byte slicing.
func placeLabels(labels []string, width int) string {
	if len(labels) == 1 {
		return labels[0]
	}
	var b strings.Builder
	cursor := 0
	for i, label := range labels {
		target := max(cursor, int(float64(i)/float64(len(labels)-1)*float64(width-lipgloss.Width(label))))
		if pad := target - cursor; pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		b.WriteString(label)
		cursor = target + lipgloss.Width(label)
	}
	return b.String()
}
