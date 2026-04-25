// Package intent renders the "stream looks like" strip — the highest-leverage
// element on the focused-link pane. Body copy is generated from the link's
// state and impairment values; humans don't author per-state strings.
package intent

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Severity drives both the border and label color of the strip.
type Severity int

const (
	Calm Severity = iota
	Warning
	Bad
)

// Snapshot is the slice of link state the intent strip needs. Decoupled from
// any block-package types so the TUI stays a leaf consumer.
type Snapshot struct {
	Loss          float64 // 0..1
	LatencyMs     uint32
	JitterMs      uint32
	BandwidthKbps uint32
	BitrateKbps   uint64 // observed downstream throughput
	LinkUp        bool
	HasController bool // false for source blocks
}

// Styles is the small palette the strip needs from the parent theme.
type Styles struct {
	Calm    lipgloss.Style // border + label
	Warning lipgloss.Style
	Bad     lipgloss.Style
	Body    lipgloss.Style
	Number  lipgloss.Style
	Frame   [3]lipgloss.Style // border styles indexed by Severity
}

// Render returns the bordered strip sized to width.
func Render(s Snapshot, st Styles, width int) string {
	sev, label, body := compose(s)
	innerW := max(width-4, 1)
	wrapped := wrap(st.Body.Render(body), innerW)
	content := st.styleFor(sev).Render(label) + "\n" + wrapped
	frame := st.Frame[sev]
	return frame.Width(width).Render(content)
}

// SeverityOf is exposed so the parent pane can pick a frame color in sync.
func SeverityOf(s Snapshot) Severity {
	sev, _, _ := compose(s)
	return sev
}

func (st Styles) styleFor(sev Severity) lipgloss.Style {
	switch sev {
	case Bad:
		return st.Bad
	case Warning:
		return st.Warning
	}
	return st.Calm
}

const headerLabel = "STREAM LOOKS LIKE"

// compose picks the highest-severity clause from §6 of DESIGN.md and
// optionally appends a secondary one. The return is (severity, header, body).
func compose(s Snapshot) (Severity, string, string) {
	if !s.LinkUp {
		return Bad, headerLabel, "consumer disconnected. perturbations preserved — they reapply on reconnect."
	}
	clauses := nonZeroClauses(s)
	if len(clauses) == 0 {
		return Calm, headerLabel, healthyClause(s)
	}
	sev := worstSeverity(clauses)
	body := joinClauses(clauses)
	return sev, headerLabel, body
}

type clause struct {
	sev  Severity
	text string
}

func nonZeroClauses(s Snapshot) []clause {
	out := make([]clause, 0, 4)
	if s.Loss > 0 {
		out = append(out, lossClause(s.Loss))
	}
	if s.LatencyMs > 0 {
		out = append(out, latencyClause(s.LatencyMs))
	}
	if s.JitterMs > 0 {
		out = append(out, jitterClause(s.JitterMs))
	}
	if s.BandwidthKbps > 0 {
		out = append(out, bandwidthClause(s.BandwidthKbps))
	}
	return out
}

func lossClause(loss float64) clause {
	rate := 0
	if loss > 0 {
		rate = int(1.0/loss + 0.5)
	}
	switch {
	case loss > 0.15:
		return clause{Bad, fmt.Sprintf("losing ~1 in %d packets. the picture will struggle.", rate)}
	case loss >= 0.05:
		return clause{Warning, fmt.Sprintf("losing ~1 in %d packets. expect macroblocking.", rate)}
	default:
		return clause{Warning, fmt.Sprintf("losing ~1 in %d packets. you'll see brief artifacts.", rate)}
	}
}

func latencyClause(ms uint32) clause {
	if ms > 250 {
		return clause{Bad, fmt.Sprintf("delayed %dms. anything realtime is gone.", ms)}
	}
	return clause{Warning, fmt.Sprintf("delayed %dms. feels like a sluggish link.", ms)}
}

func jitterClause(ms uint32) clause {
	sev := Warning
	if ms > 80 {
		sev = Bad
	}
	return clause{sev, fmt.Sprintf("wobbling ±%dms. expect re-buffering.", ms)}
}

func bandwidthClause(kbps uint32) clause {
	mbps := float64(kbps) / 1000.0
	sev := Warning
	if kbps < 1000 {
		sev = Bad
	}
	return clause{sev, fmt.Sprintf("squeezed to %.1f Mb/s. encoder will throttle.", mbps)}
}

func healthyClause(s Snapshot) string {
	if s.BitrateKbps == 0 {
		return "healthy · no perturbation applied."
	}
	return fmt.Sprintf("healthy · %.1f Mb/s · no perturbation applied.", float64(s.BitrateKbps)/1000.0)
}

func worstSeverity(cs []clause) Severity {
	worst := Calm
	for _, c := range cs {
		if c.sev > worst {
			worst = c.sev
		}
	}
	return worst
}

func joinClauses(cs []clause) string {
	if len(cs) == 1 {
		return cs[0].text
	}
	// Two highest-severity clauses, then the compound footer.
	first, second := topTwo(cs)
	return first.text + " " + second.text + " expect freezes + macroblocking."
}

func topTwo(cs []clause) (clause, clause) {
	a, b := cs[0], clause{Calm, ""}
	for _, c := range cs[1:] {
		if c.sev > a.sev {
			b = a
			a = c
			continue
		}
		if c.sev > b.sev {
			b = c
		}
	}
	if b.text == "" {
		b = cs[len(cs)-1]
	}
	return a, b
}

func wrap(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		candidate := cur
		if cur != "" {
			candidate += " "
		}
		candidate += w
		if lipgloss.Width(candidate) > width && cur != "" {
			lines = append(lines, cur)
			cur = w
			continue
		}
		cur = candidate
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}
