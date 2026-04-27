// Package intent classifies a link's impairment state into a severity
// + headline + body copy. It owns the human-readable language; it does
// not own the rendering. components/intentstrip frames the verdict.
package intent

import "fmt"

// Severity drives both the border and label color of the strip.
type Severity int

// Severity bands — calm baseline, warning when something's tugging,
// bad when the stream's gone or the link's down.
const (
	Calm Severity = iota
	Warning
	Bad
)

// Snapshot is the slice of link state the intent classifier needs.
// Decoupled from any block-package types so the TUI stays a leaf
// consumer.
type Snapshot struct {
	Loss          float64 // 0..1
	LatencyMs     uint32
	JitterMs      uint32
	BandwidthKbps uint32
	BitrateKbps   uint64 // observed downstream throughput
	LinkUp        bool
	HasController bool // false for source blocks
}

// Verdict is the structured output of Classify: a severity band, the
// short headline label ("STREAM LOOKS LIKE" / "LINK IS DOWN"), and
// the descriptive body composed from the relevant clauses.
type Verdict struct {
	Severity Severity
	Label    string
	Body     string
}

// Classify produces the verdict for a snapshot. Pure function — no
// rendering, no styling.
func Classify(s Snapshot) Verdict {
	sev, label, body := compose(s)
	return Verdict{Severity: sev, Label: label, Body: body}
}

// SeverityOf returns just the severity for callers that only need the
// frame colour decision.
func SeverityOf(s Snapshot) Severity { return Classify(s).Severity }

const (
	headerLabel     = "STREAM LOOKS LIKE"
	headerDownLabel = "LINK IS DOWN"
)

// compose picks the highest-severity clause from §6 of DESIGN.md and
// optionally appends a secondary one. The return is (severity, header,
// body). A down link gets its own header so the strip reads as a
// status, not a guess.
func compose(s Snapshot) (Severity, string, string) {
	if !s.LinkUp {
		return Bad, headerDownLabel, "consumer disconnected. perturbations preserved — they reapply on reconnect."
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
