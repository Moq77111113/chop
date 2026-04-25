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
	desc, err := p.connectUpstream()
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

func (p *proxy) connectUpstream() (*description.Session, error) {
	u, err := base.ParseURL(p.cfg.Upstream)
	if err != nil {
		return nil, err
	}
	p.upstream = &gortsplib.Client{Scheme: u.Scheme, Host: u.Host}
	if err := p.upstream.Start(); err != nil {
		return nil, err
	}
	desc, _, err := p.upstream.Describe(u)
	if err != nil {
		return nil, err
	}
	if err := p.upstream.SetupAll(desc.BaseURL, desc.Medias); err != nil {
		return nil, err
	}
	return desc, nil
}
