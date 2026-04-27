package knobs

import "fmt"

// Knob configuration constants — all in display units (loss in %,
// latency/jitter in ms, bandwidth in Mb/s). Conversion to wire format
// happens in Pane.toSnapshot().
const (
	lossScale     = 100.0  // value 0..30 in pct, snapshot 0..0.30 fraction
	kbpsToMbps    = 1000.0 // bandwidth knob is Mb/s; wire is kbps
	bandwidthMax  = 5.0    // bandwidth knob upper bound in Mb/s; >Max is "uncapped" (use ∞)
	lossWarnAt    = 1.0
	lossDangerAt  = 15.0
	latencyWarnAt = 50.0
	latencyDanger = 300.0
	jitterWarnAt  = 20.0
	jitterDanger  = 80.0
)

func defaultKnobs() [KnobCount]Knob {
	return [KnobCount]Knob{
		KnobLoss:      lossKnob(),
		KnobLatency:   latencyKnob(),
		KnobJitter:    jitterKnob(),
		KnobBandwidth: bandwidthKnob(),
	}
}

func lossKnob() Knob {
	return Knob{
		Label:       "loss",
		Min:         0,
		Max:         30,
		Step:        0.5,
		BigStep:     5,
		Format:      func(v float64) string { return fmt.Sprintf("%.1f%%", v) },
		Severity:    severityBands(lossWarnAt, lossDangerAt),
		ScaleLabels: []string{"0%", "5", "10", "20", "30%"},
	}
}

func latencyKnob() Knob {
	return Knob{
		Label:       "latency",
		Min:         0,
		Max:         500,
		Step:        10,
		BigStep:     100,
		Format:      func(v float64) string { return fmt.Sprintf("%.0f ms", v) },
		Severity:    severityBands(latencyWarnAt, latencyDanger),
		ScaleLabels: []string{"0", "100", "250", "500ms"},
	}
}

func jitterKnob() Knob {
	return Knob{
		Label:       "jitter",
		Min:         0,
		Max:         150,
		Step:        5,
		BigStep:     50,
		Format:      func(v float64) string { return fmt.Sprintf("±%.0f ms", v) },
		Severity:    severityBands(jitterWarnAt, jitterDanger),
		ScaleLabels: []string{"0", "50", "100", "150ms"},
	}
}

// bandwidthKnob axis: left=off (0 Mb/s), right=∞ (uncapped). Initial
// value and reset target are both Max (∞), since "no impairment" is
// the canonical neutral state for every chop knob.
func bandwidthKnob() Knob {
	return Knob{
		Label:   "bandwidth",
		Value:   bandwidthMax,
		Min:     0,
		Max:     bandwidthMax,
		Step:    0.1,
		BigStep: 0.5,
		ResetTo: bandwidthMax,
		Format: func(v float64) string {
			if v >= bandwidthMax {
				return "∞"
			}
			if v <= 0 {
				return "off"
			}
			return fmt.Sprintf("%.1f Mb/s", v)
		},
		Severity:    bandwidthSeverity,
		ScaleLabels: []string{"off", "1M", "2.5M", "∞"},
	}
}

func severityBands(warn, danger float64) func(float64) Severity {
	return func(v float64) Severity {
		switch {
		case v >= danger:
			return SevDanger
		case v >= warn:
			return SevWarn
		}
		return SevOK
	}
}

// bandwidthSeverity reads the slider as "amount of bandwidth allowed":
// at or near Max → ample (OK), middling → warn, near zero → danger.
func bandwidthSeverity(v float64) Severity {
	switch {
	case v >= bandwidthMax/2:
		return SevOK
	case v >= 1:
		return SevWarn
	}
	return SevDanger
}
