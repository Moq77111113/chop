// Package source serves a looped H264 Annex-B file over RTSP.
// M1 scope: file replay only. Pure-Go pattern generation is M2.
package source

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/moq77111113/chop/block"
)

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
	block.Emit(ctx, "source.up", map[string]string{"listen": b.source.Listen})

	return replay(ctx, srv.stream, b.source.File, b.source.FPS, &b.served)
}
