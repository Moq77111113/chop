// Package linkcard renders one row of the link list: title, knobs
// sub-row, state pill, sparkline, rate. The row is a flat two-line
// stack — no border, no padding — matching design/screens/03-perturbed.
// Selection shows as a primary-coloured title prefixed with "▸".
package linkcard

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/data"
	"github.com/moq77111113/chop/internal/tui/ui"
)

// Props wraps the cross-layer Row plus the focus flag for this card.
type Props struct {
	Row     data.Row
	Focused bool
}

// Styles bundles the palette the row renderer needs.
type Styles struct {
	Row         lipgloss.Style
	RowSelected lipgloss.Style
	Title       lipgloss.Style
	Cursor      lipgloss.Style
	Rate        lipgloss.Style
	KnobsLabel  lipgloss.Style
	KnobsOK     lipgloss.Style
	KnobsDim    lipgloss.Style
	KnobsWarn   lipgloss.Style
	KnobsDanger lipgloss.Style
	Spark       ui.SparkStyles
	Pill        ui.PillStyles
}

const (
	rowOverhead = 4
	minTitleW   = 8

	statusSlotW = 9  // longest pill ("DEGRADED" = 8) + 1 trailing
	barsSlotW   = 12 // sparkline width
	rateSlotW   = 9  // rate ("2.4 Mb/s" = 8) + 1 leading
	colGap      = 2
)

const rightTotalW = statusSlotW + colGap + barsSlotW + colGap + rateSlotW

// Render draws the row at exactly width cells.
func Render(p Props, s Styles, width int) string {
	title := titleOrID(p.Row)
	cursor := renderCursor(p.Focused, s)
	cursorW := lipgloss.Width(cursor)
	inner := max(width-rowOverhead, 10)

	rightCluster := buildRightCluster(p.Row, s)
	titleLine := composeTitleLine(cursor, cursorW, title, rightCluster, s.Title, inner)
	body := composeBody(p.Row, s, p.Focused, titleLine, cursorW, inner)

	frame := s.Row
	if p.Focused {
		frame = s.RowSelected
	}
	return frame.Render(body)
}

func titleOrID(r data.Row) string {
	if r.Title != "" {
		return r.Title
	}
	return r.ID
}

func renderCursor(focused bool, s Styles) string {
	if focused {
		return s.Cursor.Render("▸") + " "
	}
	return "  "
}

func buildRightCluster(row data.Row, s Styles) string {
	statusSlot := padRight(ui.Pill(row.State, s.Pill), statusSlotW)
	barsSlot := strings.Repeat(" ", barsSlotW)
	if len(row.Spark) > 0 {
		barsSlot = padRight(ui.Spark(row.Spark, s.Spark), barsSlotW)
	}
	rateSlot := strings.Repeat(" ", rateSlotW)
	if row.Rate != "" {
		rateSlot = padLeft(s.Rate.Render(row.Rate), rateSlotW)
	}
	gapStr := strings.Repeat(" ", colGap)
	return statusSlot + gapStr + barsSlot + gapStr + rateSlot
}

func composeTitleLine(cursor string, cursorW int, title, rightCluster string, titleStyle lipgloss.Style, inner int) string {
	titleBudget := max(inner-cursorW-rightTotalW-colGap, minTitleW)
	titleText := titleStyle.Render(truncate(title, titleBudget))
	spacer := strings.Repeat(" ", max(inner-cursorW-lipgloss.Width(titleText)-rightTotalW, colGap))
	return padRight(cursor+titleText+spacer+rightCluster, inner)
}

func composeBody(row data.Row, s Styles, focused bool, titleLine string, cursorW, inner int) string {
	if row.Type != data.BlockTypeLink {
		return titleLine
	}
	knobIndent := strings.Repeat(" ", cursorW)
	knobLine := padRight(knobIndent+renderKnobs(row.State, row.Knobs, s), inner)
	if focused {
		return titleLine + "\n" + strings.Repeat(" ", inner) + "\n" + knobLine
	}
	return titleLine + "\n" + knobLine
}

func padRight(line string, width int) string {
	if w := lipgloss.Width(line); w < width {
		return line + strings.Repeat(" ", width-w)
	}
	return line
}

func padLeft(line string, width int) string {
	if w := lipgloss.Width(line); w < width {
		return strings.Repeat(" ", width-w) + line
	}
	return line
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

// renderKnobs paints the per-row knob summary using the severity tags
// from data.Values. Down state forces every value to danger; the
// bandwidth "off" slot (∞ or —) reads dim regardless of state, unless
// the link itself is down.
func renderKnobs(linkState data.LinkState, v data.Values, s Styles) string {
	down := linkState == data.StateDown || linkState == data.StateStopped
	val := func(token string, sev data.Severity, dim bool) string {
		switch {
		case down:
			return s.KnobsDanger.Render(token)
		case dim:
			return s.KnobsDim.Render(token)
		}
		return styleForSeverity(sev, s).Render(token)
	}
	parts := []string{
		s.KnobsLabel.Render("loss") + " " + val(v.Loss, v.LossSev, isZeroToken(v.Loss)),
		s.KnobsLabel.Render("lat") + " " + val(v.Latency, v.LatencySev, isZeroToken(v.Latency)),
		s.KnobsLabel.Render("jit") + " " + val(v.Jitter, v.JitterSev, isZeroToken(v.Jitter)),
		s.KnobsLabel.Render("bw") + " " + val(v.Bandwidth, v.BandwidthSev, v.BandwidthOff && !down),
	}
	return strings.Join(parts, "  ")
}

func styleForSeverity(sev data.Severity, s Styles) lipgloss.Style {
	switch sev {
	case data.SevWarn:
		return s.KnobsWarn
	case data.SevDanger:
		return s.KnobsDanger
	}
	return s.KnobsOK
}

func isZeroToken(token string) bool {
	switch token {
	case "0%", "0ms", "±0ms":
		return true
	}
	return false
}
