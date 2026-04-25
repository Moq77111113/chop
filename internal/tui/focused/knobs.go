package focused

import (
	"fmt"

	"github.com/moq77111113/chop/internal/tui/knob"
)

// Knob configuration constants — all in display units (loss in %,
// latency/jitter in ms, bandwidth in Mb/s). Conversion to wire format
// happens in toSnapshot().
const (
	lossScale     = 100.0  // value 0..30 in pct, snapshot 0..0.30 fraction
	kbpsToMbps    = 1000.0 // bandwidth knob is Mb/s; wire is kbps
	lossWarnAt    = 1.0
	lossDangerAt  = 15.0
	latencyWarnAt = 50.0
	latencyDanger = 300.0
	jitterWarnAt  = 20.0
	jitterDanger  = 80.0
)

func defaultKnobs() [knobCount]knob.Knob {
	return [knobCount]knob.Knob{
		KnobLoss:      lossKnob(),
		KnobLatency:   latencyKnob(),
		KnobJitter:    jitterKnob(),
		KnobBandwidth: bandwidthKnob(),
	}
}

func lossKnob() knob.Knob {
	return knob.Knob{
		Label:    "loss",
		Min:      0,
		Max:      30,
		Step:     0.5,
		BigStep:  5,
		Format:   func(v float64) string { return fmt.Sprintf("%.1f%%", v) },
		Severity: severityBands(lossWarnAt, lossDangerAt),
	}
}

func latencyKnob() knob.Knob {
	return knob.Knob{
		Label:    "latency",
		Min:      0,
		Max:      500,
		Step:     10,
		BigStep:  100,
		Format:   func(v float64) string { return fmt.Sprintf("%.0f ms", v) },
		Severity: severityBands(latencyWarnAt, latencyDanger),
	}
}

func jitterKnob() knob.Knob {
	return knob.Knob{
		Label:    "jitter",
		Min:      0,
		Max:      150,
		Step:     5,
		BigStep:  50,
		Format:   func(v float64) string { return fmt.Sprintf("±%.0f ms", v) },
		Severity: severityBands(jitterWarnAt, jitterDanger),
	}
}

func bandwidthKnob() knob.Knob {
	return knob.Knob{
		Label:    "bandwidth",
		Min:      0,
		Max:      50,
		Step:     0.1,
		BigStep:  1,
		Inverted: true,
		Format: func(v float64) string {
			if v == 0 {
				return "∞"
			}
			return fmt.Sprintf("%.1f Mb/s", v)
		},
		Severity: bandwidthSeverity,
	}
}

func severityBands(warn, danger float64) func(float64) knob.Severity {
	return func(v float64) knob.Severity {
		switch {
		case v >= danger:
			return knob.SevDanger
		case v >= warn:
			return knob.SevWarn
		}
		return knob.SevOK
	}
}

func bandwidthSeverity(v float64) knob.Severity {
	switch {
	case v == 0 || v > 5:
		return knob.SevOK
	case v >= 1:
		return knob.SevWarn
	}
	return knob.SevDanger
}
