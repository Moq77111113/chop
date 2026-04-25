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

// Knob describes one slider's data and behavior. Inverted reverses the
// fill direction so a "0 = off / unlimited" knob (e.g. bandwidth) reads
// full at zero and empties as the value climbs.
type Knob struct {
	Label    string
	Value    float64
	Min      float64
	Max      float64
	Step     float64
	BigStep  float64
	Inverted bool
	Format   func(float64) string
	Severity func(float64) Severity
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

// Zero returns a copy with Value reset to Min.
func (k Knob) Zero() Knob {
	k.Value = k.Min
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
		ratio := (k.Value - k.Min) / (k.Max - k.Min)
		if k.Inverted {
			ratio = 1 - ratio
		}
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
	if k.Max <= k.Min {
		return ""
	}
	ticks := []float64{k.Min, k.Min + (k.Max-k.Min)*0.25, k.Min + (k.Max-k.Min)*0.5, k.Min + (k.Max-k.Min)*0.75, k.Max}
	labels := make([]string, len(ticks))
	for i, t := range ticks {
		labels[i] = k.Format(t)
	}
	cells := make([]string, width)
	for i := range cells {
		cells[i] = " "
	}
	for i, label := range labels {
		pos := int(float64(i) / float64(len(ticks)-1) * float64(width-len(label)))
		for j := 0; j < len(label) && pos+j < width; j++ {
			cells[pos+j] = string(label[j])
		}
	}
	return st.Scale.Render(" " + strings.Join(cells, ""))
}
