package tui

import "strings"

// sparkChars maps a normalized 0..1 sample to one of the eight partial-block
// glyphs. Index 0 is space (no signal); 1..8 climb in equal steps.
var sparkChars = []rune(" ▁▂▃▄▅▆▇█")

const sparkLength = 12

// renderSparkline draws a fixed-width row of partial-block characters
// scaled by the maximum sample in the window. An empty/short window pads
// with spaces so widths align across rows.
func renderSparkline(samples []float64) string {
	if len(samples) == 0 {
		return strings.Repeat(" ", sparkLength)
	}
	maxSample := samples[0]
	for _, s := range samples[1:] {
		maxSample = max(maxSample, s)
	}
	if maxSample <= 0 {
		return strings.Repeat(" ", sparkLength)
	}
	var b strings.Builder
	if pad := max(sparkLength-len(samples), 0); pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}
	steps := len(sparkChars) - 1
	for _, s := range samples {
		idx := int((s/maxSample)*float64(steps) + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx > steps {
			idx = steps
		}
		b.WriteRune(sparkChars[idx])
	}
	return b.String()
}

// pushSample appends v to history bounded by sparkLength.
func pushSample(history []float64, v float64) []float64 {
	history = append(history, v)
	if len(history) > sparkLength {
		history = history[len(history)-sparkLength:]
	}
	return history
}
