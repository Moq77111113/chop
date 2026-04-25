package link

import (
	"context"
	"time"

	"github.com/pion/rtp"

	"github.com/moq77111113/chop/block"
)

func (p *proxy) onPacket(ctx context.Context, pkt *rtp.Packet) {
	p.counters.pktsIn.Add(1)
	d := p.decide()

	if d.Drop {
		p.counters.pktsDropped.Add(1)
		block.Emit(ctx, eventRTPDropped, map[string]any{
			"reason": d.DropReason,
			"seq":    pkt.SequenceNumber,
		})
		return
	}
	if d.Delay > 0 {
		time.AfterFunc(d.Delay, func() { p.forward(pkt) })
		return
	}
	p.forward(pkt)
}

func (p *proxy) decide() Decision {
	p.rngMu.Lock()
	defer p.rngMu.Unlock()
	return Decide(p.ctrls.Load(), p.rng)
}

func (p *proxy) forward(pkt *rtp.Packet) {
	if err := p.down.stream.WritePacketRTP(p.media, pkt); err == nil {
		p.counters.pktsOut.Add(1)
	}
}
