package knobs

// KnobIndex identifies one of the four impairment knobs by position.
type KnobIndex int

// KnobIndex constants and KnobCount — the four-knob inventory.
const (
	KnobLoss KnobIndex = iota
	KnobLatency
	KnobJitter
	KnobBandwidth
	KnobCount = 4
)

// Snapshot is the slim view of a link's controls. Decoupled from any
// block-package types so the TUI stays a leaf consumer.
type Snapshot struct {
	Loss          float64
	LatencyMs     uint32
	JitterMs      uint32
	BandwidthKbps uint32
	BitrateKbps   uint64
	LinkUp        bool
}

// Pane holds the four knob slots and the focused index.
type Pane struct {
	knobs   [KnobCount]Knob
	focused KnobIndex
}

// NewPane returns the default chop knob pane spec'd in DESIGN.md §4.4.
func NewPane() *Pane { return &Pane{knobs: defaultKnobs()} }

// Focused returns the currently focused knob index.
func (p *Pane) Focused() KnobIndex { return p.focused }

// Knob returns a copy of the knob at idx.
func (p *Pane) Knob(idx KnobIndex) Knob { return p.knobs[idx] }

// NextKnob cycles forward through the four knobs.
func (p *Pane) NextKnob() { p.focused = (p.focused + 1) % KnobCount }

// PrevKnob cycles backward through the four knobs.
func (p *Pane) PrevKnob() { p.focused = (p.focused + KnobCount - 1) % KnobCount }

// Adjust applies one Step in the requested direction to the focused
// knob and returns the resulting Snapshot the parent can ship as
// Controls.
func (p *Pane) Adjust(direction int) Snapshot {
	k := p.knobs[p.focused]
	step := k.Step
	if direction < 0 {
		step = -step
	}
	p.knobs[p.focused] = k.Adjust(step)
	return p.toSnapshot()
}

// Zero resets the focused knob to its neutral and returns the
// resulting Snapshot.
func (p *Pane) Zero() Snapshot {
	p.knobs[p.focused] = p.knobs[p.focused].Zero()
	return p.toSnapshot()
}

// ResetAll zeroes every knob.
func (p *Pane) ResetAll() Snapshot {
	for i := range p.knobs {
		p.knobs[i] = p.knobs[i].Zero()
	}
	return p.toSnapshot()
}

// SetSnapshot pushes fresh values into the knobs from a wire-side
// snapshot. BandwidthKbps == 0 is the wire's "no shaping" sentinel; we
// display it at the ∞ end (Max).
func (p *Pane) SetSnapshot(s Snapshot) {
	p.knobs[KnobLoss].Value = s.Loss * lossScale
	p.knobs[KnobLatency].Value = float64(s.LatencyMs)
	p.knobs[KnobJitter].Value = float64(s.JitterMs)
	if s.BandwidthKbps == 0 {
		p.knobs[KnobBandwidth].Value = p.knobs[KnobBandwidth].Max
	} else {
		p.knobs[KnobBandwidth].Value = float64(s.BandwidthKbps) / kbpsToMbps
	}
}

func (p *Pane) toSnapshot() Snapshot {
	bw := p.knobs[KnobBandwidth]
	bwKbps := uint32(bw.Value * kbpsToMbps)
	if bw.Value >= bw.Max {
		bwKbps = 0
	}
	return Snapshot{
		Loss:          p.knobs[KnobLoss].Value / lossScale,
		LatencyMs:     uint32(p.knobs[KnobLatency].Value),
		JitterMs:      uint32(p.knobs[KnobJitter].Value),
		BandwidthKbps: bwKbps,
	}
}
