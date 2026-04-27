// Package knobs holds the four-knob impairment model used by the right
// pane. It owns Adjust / Zero / ResetAll math and the Snapshot↔values
// round-trip; it imports no lipgloss or bubbletea.
package knobs

// Severity is the colour band a value falls into for a given knob.
// Components map this to their fill / value styles; the math layer
// only signals which band.
type Severity int

// Severity bands — ordered worst-to-best is reversed: SevOK is index 0
// so an OK band is the default zero value.
const (
	SevOK Severity = iota
	SevWarn
	SevDanger
)

// Knob describes one slider's data and behaviour. ResetTo, when
// non-zero, is the value Zero() reverts to (defaults to Min — fine for
// impairment knobs where 0 is "no impairment", but bandwidth uses Max
// as its neutral). ScaleLabels overrides the auto-generated tick
// labels.
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

// Zero returns a copy with Value reset to ResetTo (or Min if ResetTo
// is the zero value of float64, which is the conventional "neutral"
// for impairment knobs anyway).
func (k Knob) Zero() Knob {
	k.Value = k.ResetTo
	if k.ResetTo == 0 {
		k.Value = k.Min
	}
	return k
}
