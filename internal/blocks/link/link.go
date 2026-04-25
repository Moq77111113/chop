// Package link implements the chop "link" block: an RTSP pull-proxy that
// applies impairments (loss, latency, jitter) per RTP packet.
package link

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/moq77111113/chop/block"
)

// Config is the static configuration of a link block, parsed from the
// scenario YAML `config:` section.
type Config struct {
	Upstream string `json:"upstream"`
	ServeAt  string `json:"serve_at"`
}

var (
	errMissingUpstream = errors.New("link: upstream is required")
	errMissingServeAt  = errors.New("link: serve_at is required")
)

// LinkBlock is the link block implementation: an RTSP pull-proxy that
// applies live impairments (loss, latency, jitter) to forwarded RTP packets.
type LinkBlock struct {
	cfg      block.Config
	link     Config
	parseErr error
	ctrls    *ctrlBox
	counters *counters
	proxy    *proxy
}

// New is the link block factory. Configuration errors are deferred to Run
// so the block can be constructed unconditionally by the runtime.
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

func parseConfig(raw json.RawMessage) (Config, error) {
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("link: parse config: %w", err)
	}
	if c.Upstream == "" {
		return Config{}, errMissingUpstream
	}
	if c.ServeAt == "" {
		return Config{}, errMissingServeAt
	}
	return c, nil
}

func parseInitialControls(raw json.RawMessage) Controls {
	var c Controls
	if len(raw) == 0 {
		return c
	}
	_ = json.Unmarshal(raw, &c)
	return c
}
