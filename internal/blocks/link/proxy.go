package link

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"

	"github.com/moq77111113/chop/block"
)

const (
	eventLinkUp     = "link.up"
	eventRTPDropped = "rtp.dropped"

	upstreamRetryAttempts = 20
	upstreamRetryDelay    = 200 * time.Millisecond
)

type proxy struct {
	cfg      Config
	ctrls    *ctrlBox
	counters *counters

	upstream *gortsplib.Client
	down     *downstream
	media    *description.Media

	rngMu sync.Mutex
	rng   *rand.Rand
}

func newProxy(cfg Config, ctrls *ctrlBox, ctrs *counters) *proxy {
	return &proxy{
		cfg:      cfg,
		ctrls:    ctrls,
		counters: ctrs,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *proxy) run(ctx context.Context) error {
	desc, err := p.connectUpstream(ctx)
	if err != nil {
		return fmt.Errorf("upstream: %w", err)
	}
	defer p.upstream.Close()

	down, err := startDownstream(p.cfg.ServeAt, desc)
	if err != nil {
		return fmt.Errorf("downstream: %w", err)
	}
	defer down.close()
	p.down = down
	p.media = desc.Medias[0]

	p.upstream.OnPacketRTPAny(func(_ *description.Media, _ format.Format, pkt *rtp.Packet) {
		p.onPacket(ctx, pkt)
	})

	if _, err := p.upstream.Play(nil); err != nil {
		return fmt.Errorf("play: %w", err)
	}

	p.counters.upSinceMs.Store(time.Now().UnixMilli())
	block.Emit(ctx, eventLinkUp, map[string]string{
		"upstream": p.cfg.Upstream,
		"serve_at": p.cfg.ServeAt,
	})

	<-ctx.Done()
	return nil
}

// connectUpstream retries the RTSP handshake until it succeeds, ctx is
// cancelled, or upstreamRetryAttempts are exhausted. M1 has no DAG so the
// link block can boot before the source has finished binding.
func (p *proxy) connectUpstream(ctx context.Context) (*description.Session, error) {
	var lastErr error
	for range upstreamRetryAttempts {
		desc, err := p.dialUpstream()
		if err == nil {
			return desc, nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(upstreamRetryDelay):
		}
	}
	return nil, fmt.Errorf("%w (after %d attempts)", lastErr, upstreamRetryAttempts)
}

func (p *proxy) dialUpstream() (*description.Session, error) {
	u, err := base.ParseURL(p.cfg.Upstream)
	if err != nil {
		return nil, err
	}
	client := &gortsplib.Client{Scheme: u.Scheme, Host: u.Host}
	if err := client.Start(); err != nil {
		return nil, err
	}
	desc, _, err := client.Describe(u)
	if err != nil {
		client.Close()
		return nil, err
	}
	if err := client.SetupAll(desc.BaseURL, desc.Medias); err != nil {
		client.Close()
		return nil, err
	}
	p.upstream = client
	return desc, nil
}

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
