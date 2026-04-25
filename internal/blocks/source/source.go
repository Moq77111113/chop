// Package source serves a looped H264 Annex-B file over RTSP.
// M1 scope: file replay only. Pure-Go pattern generation is M2.
package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/moq77111113/chop/block"
)

const (
	defaultFPS    = 25
	eventSourceUp = "source.up"
)

// Config is the static configuration of a source block, parsed from the
// scenario YAML `config:` section.
type Config struct {
	File   string `json:"file"`
	Listen string `json:"listen"`
	FPS    int    `json:"fps"`
}

var (
	errMissingFile   = errors.New("source: file is required")
	errMissingListen = errors.New("source: listen is required")
)

// SourceBlock is the source block implementation: it serves a looped H264
// Annex-B file as an RTSP stream.
type SourceBlock struct {
	cfg       block.Config
	source    Config
	parseErr  error
	served    atomic.Int64
	upSinceMs atomic.Int64
}

type stats struct {
	RTPServed int64 `json:"rtp_served"`
	UpSinceMs int64 `json:"up_since_ms"`
}

// New is the source block factory. Configuration errors are deferred to Run
// so the block can be constructed unconditionally by the runtime.
func New(c block.Config) block.Block {
	src, err := parseConfig(c.Static)
	return &SourceBlock{cfg: c, source: src, parseErr: err}
}

func (b *SourceBlock) Info() block.Info {
	return block.Info{ID: b.cfg.ID, Type: b.cfg.Type, Config: b.cfg.Static}
}

func (b *SourceBlock) Snapshot() block.Snapshot {
	payload, _ := json.Marshal(stats{
		RTPServed: b.served.Load(),
		UpSinceMs: b.upSinceMs.Load(),
	})
	return block.Snapshot{
		Status: block.StatusRunning,
		Stats:  payload,
		TsMs:   time.Now().UnixMilli(),
	}
}

func (b *SourceBlock) Apply(json.RawMessage) error          { return nil }
func (b *SourceBlock) Action(string, json.RawMessage) error { return nil }

func (b *SourceBlock) Run(ctx context.Context) error {
	if b.parseErr != nil {
		return b.parseErr
	}
	srv, err := startServer(b.source.Listen)
	if err != nil {
		return err
	}
	defer srv.close()

	b.upSinceMs.Store(time.Now().UnixMilli())
	block.Emit(ctx, eventSourceUp, map[string]string{"listen": b.source.Listen})

	return replay(ctx, srv.stream, b.source.File, b.source.FPS, &b.served)
}

func parseConfig(raw json.RawMessage) (Config, error) {
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("source: parse config: %w", err)
	}
	if c.File == "" {
		return Config{}, errMissingFile
	}
	if c.Listen == "" {
		return Config{}, errMissingListen
	}
	if c.FPS == 0 {
		c.FPS = defaultFPS
	}
	return c, nil
}
