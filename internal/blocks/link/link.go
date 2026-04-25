// Package link implements the chop "link" block: an RTSP pull-proxy that
// applies impairments (loss, latency, jitter) per RTP packet.
package link

import (
	"context"
	"encoding/json"

	"github.com/moq77111113/chop/block"
)

type LinkBlock struct {
	cfg      block.Config
	link     Config
	parseErr error
	ctrls    *ctrlBox
	counters *counters
	proxy    *proxy
}

func New(c block.Config) block.Block {
	cfg, err := parseConfig(c.Static)
	ctrs := &counters{}
	box := newCtrlBox(parseInitialControls(c.Live))
	return &LinkBlock{
		cfg:      c,
		link:     cfg,
		parseErr: err,
		ctrls:    box,
		counters: ctrs,
		proxy:    newProxy(cfg, box, ctrs),
	}
}

func (b *LinkBlock) Info() block.Info {
	return block.Info{ID: b.cfg.ID, Type: b.cfg.Type, Config: b.cfg.Static}
}

func (b *LinkBlock) Snapshot() block.Snapshot { return snapshotOf(b.counters) }

func (b *LinkBlock) Apply(p json.RawMessage) error {
	var c Controls
	if err := json.Unmarshal(p, &c); err != nil {
		return err
	}
	b.ctrls.Store(&c)
	return nil
}

func (b *LinkBlock) Action(string, json.RawMessage) error { return nil }

func (b *LinkBlock) Run(ctx context.Context) error {
	if b.parseErr != nil {
		return b.parseErr
	}
	return b.proxy.run(ctx)
}
