package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sparkChars = []rune(" ▁▂▃▄▅▆▇█")

// SparkLength is the number of sparkline columns. Matches
// state.SparkLength so the rolling history pushes the right window
// size; kept on this side too so renderers don't need to import state.
const SparkLength = 12

// SparkStyles holds the bar foreground style.
type SparkStyles struct {
	Bar lipgloss.Style
}

// Spark renders a fixed-width sparkline scaled by the maximum sample
// in the window. An empty / all-zero window pads with spaces so widths
// align across rows.
func Spark(samples []float64, s SparkStyles) string {
	return s.Bar.Render(sparkBody(samples))
}

func sparkBody(samples []float64) string {
	if len(samples) == 0 {
		return strings.Repeat(" ", SparkLength)
	}
	maxSample := samples[0]
	for _, v := range samples[1:] {
		if v > maxSample {
			maxSample = v
		}
	}
	if maxSample <= 0 {
		return strings.Repeat(" ", SparkLength)
	}
	var b strings.Builder
	if pad := max(SparkLength-len(samples), 0); pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}
	steps := len(sparkChars) - 1
	for _, v := range samples {
		idx := min(max(int((v/maxSample)*float64(steps)+0.5), 0), steps)
		b.WriteRune(sparkChars[idx])
	}
	return b.String()
}
