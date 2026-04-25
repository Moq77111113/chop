// Package link implements the chop "link" block: an RTSP proxy that pulls
// from an upstream and serves to a client, applying impairments per RTP packet.
package link

import "sync/atomic"

type Controls struct {
	Loss          float64 `json:"loss"`
	LatencyMs     uint32  `json:"latency_ms"`
	JitterMs      uint32  `json:"jitter_ms"`
	BandwidthKbps uint32  `json:"bandwidth_kbps"`
}

type ctrlBox struct{ v atomic.Pointer[Controls] }

func newCtrlBox(initial Controls) *ctrlBox {
	box := &ctrlBox{}
	box.v.Store(&initial)
	return box
}

func (c *ctrlBox) Load() *Controls      { return c.v.Load() }
func (c *ctrlBox) Store(next *Controls) { c.v.Store(next) }
