package data

import "fmt"

const (
	percentScale = 100.0
	kbpsToMbps   = 1000.0

	lossWarnPct     = 1.0
	lossDangerPct   = 15.0
	latencyWarnMs   = 50.0
	latencyDangerMs = 300.0
	jitterWarnMs    = 20.0
	jitterDangerMs  = 80.0
	bwOkMbps        = 2.5
	bwWarnMbps      = 1.0
)

// Severity tags a knob value so the renderer can paint OK / Warn /
// Danger without re-deriving the thresholds.
type Severity int

const (
	SevOK Severity = iota
	SevWarn
	SevDanger
)

// Values holds the four per-knob value tokens plus their severity tag
// so the renderer can paint each value without re-parsing the string.
// BandwidthOff stays for the "no shaping" case (∞ or — when down).
type Values struct {
	Loss, Latency, Jitter, Bandwidth             string
	LossSev, LatencySev, JitterSev, BandwidthSev Severity
	BandwidthOff                                 bool
}

// KnobValues formats the four control-knob values from a LinkSnapshot
// and tags each with a severity band. The state argument lets us
// substitute "—" for bandwidth when a link is down.
func KnobValues(s LinkSnapshot, st LinkState) Values {
	lossPct := s.Controls.Loss * percentScale
	v := Values{
		Loss:       fmt.Sprintf("%d%%", int(lossPct+0.5)),
		Latency:    fmt.Sprintf("%dms", s.Controls.LatencyMs),
		Jitter:     fmt.Sprintf("±%dms", s.Controls.JitterMs),
		LossSev:    severityOf(lossPct, lossWarnPct, lossDangerPct),
		LatencySev: severityOf(float64(s.Controls.LatencyMs), latencyWarnMs, latencyDangerMs),
		JitterSev:  severityOf(float64(s.Controls.JitterMs), jitterWarnMs, jitterDangerMs),
	}
	switch {
	case st == StateDown || st == StateStopped:
		v.Bandwidth = "—"
		v.BandwidthOff = true
	case s.Controls.BandwidthKbps == 0:
		v.Bandwidth = "∞"
		v.BandwidthOff = true
	default:
		mbps := float64(s.Controls.BandwidthKbps) / kbpsToMbps
		v.Bandwidth = fmt.Sprintf("%.1fM", mbps)
		v.BandwidthSev = bandwidthSeverity(mbps)
	}
	return v
}

func severityOf(value, warn, danger float64) Severity {
	switch {
	case value >= danger:
		return SevDanger
	case value >= warn:
		return SevWarn
	}
	return SevOK
}

func bandwidthSeverity(mbps float64) Severity {
	switch {
	case mbps >= bwOkMbps:
		return SevOK
	case mbps >= bwWarnMbps:
		return SevWarn
	}
	return SevDanger
}
