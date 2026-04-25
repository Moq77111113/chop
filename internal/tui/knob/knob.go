// Package knob renders a single horizontal slider with label, value, track,
// and tick scale. Each knob is data-only — input routing lives in the parent
// pane that owns the focus ring.
package knob

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Severity is the color band a value falls into for a given knob.
type Severity int

const (
	SevOK Severity = iota
	SevWarn
	SevDanger
)

// Knob describes one slider's data and behavior. ResetTo, when non-zero,
// is the value Zero() reverts to (defaults to Min — fine for impairment
// knobs where 0 is "no impairment", but bandwidth uses Max as its
// neutral). ScaleLabels overrides the auto-generated tick labels.
type Knob struct {
	Label       string
	Value       float64
	Min         float64
	Max         float64
	Step        float64
	BigStep     float64
	ResetTo     float64
	Format      func(float64) string
	Severity    func(float64) Severity
	ScaleLabels []string
}

// Adjust returns a copy with Value clamped after delta is applied.
func (k Knob) Adjust(delta float64) Knob {
	v := k.Value + delta
	if v < k.Min {
		v = k.Min
	}
	if v > k.Max {
		v = k.Max
	}
	k.Value = v
	return k
}

// Zero returns a copy with Value reset to ResetTo (or Min if ResetTo is
// the zero value of float64, which is the conventional "neutral" for
// impairment knobs anyway).
func (k Knob) Zero() Knob {
	k.Value = k.ResetTo
	if k.ResetTo == 0 {
		k.Value = k.Min
	}
	return k
}

// Styles is the small palette the knob needs from the parent theme.
type Styles struct {
	Label    lipgloss.Style
	LabelOn  lipgloss.Style
	Hint     lipgloss.Style
	Track    lipgloss.Style
	Fill     [3]lipgloss.Style // indexed by Severity
	Value    [3]lipgloss.Style
	Scale    lipgloss.Style
	FocusBar lipgloss.Style
}

const (
	trackFillChar  = "█"
	trackEmptyChar = "░"
	hintText       = "←→ adjust"
)

// Render draws a 3-line knob (label/value, track, scale) sized to width.
// `focused` swaps in the focused styles and shows the arrow hint.
func Render(k Knob, st Styles, width int, focused bool) string {
	if width < 20 {
		width = 20
	}
	sev := SevOK
	if k.Severity != nil {
		sev = k.Severity(k.Value)
	}
	header := renderHeader(k, st, width, focused, sev)
	track := renderTrack(k, st, width, sev, focused)
	scale := renderScale(k, st, width)
	return strings.Join([]string{header, track, scale}, "\n")
}

func renderHeader(k Knob, st Styles, width int, focused bool, sev Severity) string {
	labelStyle := st.Label
	if focused {
		labelStyle = st.LabelOn
	}
	label := labelStyle.Render(strings.ToUpper(k.Label))
	val := st.Value[sev].Render(k.Format(k.Value))
	hint := ""
	if focused {
		hint = st.Hint.Render(hintText)
	}
	gapL := max(width-lipgloss.Width(label)-lipgloss.Width(hint)-lipgloss.Width(val), 1)
	if hint == "" {
		return label + strings.Repeat(" ", gapL) + val
	}
	gapR := gapL / 2
	gapM := gapL - gapR
	return label + strings.Repeat(" ", gapM) + hint + strings.Repeat(" ", gapR) + val
}

func renderTrack(k Knob, st Styles, width int, sev Severity, focused bool) string {
	fillCount := 0
	if k.Max > k.Min {
		ratio := clamp01((k.Value - k.Min) / (k.Max - k.Min))
		fillCount = max(0, min(width, int(ratio*float64(width)+0.5)))
	}
	fill := st.Fill[sev].Render(strings.Repeat(trackFillChar, fillCount))
	empty := st.Track.Render(strings.Repeat(trackEmptyChar, width-fillCount))
	bar := fill + empty
	if focused {
		return st.FocusBar.Render("▌") + bar
	}
	return " " + bar
}

func renderScale(k Knob, st Styles, width int) string {
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
	return st.Scale.Render(" " + placeLabels(labels, width))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func autoLabels(k Knob) []string {
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
	// Need at least one space between adjacent labels.
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
