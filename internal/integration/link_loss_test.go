package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/blocks/link"
	"github.com/moq77111113/chop/internal/blocks/source"
)

const (
	srcListen      = "127.0.0.1:15101"
	linkServeAt    = "127.0.0.1:18501"
	srcStreamPath  = "/stream"
	upstreamURL    = "rtsp://" + srcListen + srcStreamPath
	patternFixture = "../../testdata/pattern.h264"
	srcFPS         = 25

	srcWarmup      = 300 * time.Millisecond
	linkWarmup     = 1500 * time.Millisecond
	measureWindow  = 4 * time.Second
	overallTimeout = 10 * time.Second

	targetLoss        = 0.20
	lossLowerBound    = 0.13
	lossUpperBound    = 0.27
	minPacketsForTest = 50
)

func TestLink_LossRateMatchesControl(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	src := source.New(block.Config{
		ID: "src", Type: "source",
		Static: mustJSON(t, source.Config{
			File: patternFixture, Listen: srcListen, FPS: srcFPS,
		}),
	})

	lnk := link.New(block.Config{
		ID: "lnk", Type: "link",
		Static: mustJSON(t, link.Config{Upstream: upstreamURL, ServeAt: linkServeAt}),
		Live:   mustJSON(t, link.Controls{Loss: targetLoss}),
	})

	ctx, cancel := context.WithTimeout(context.Background(), overallTimeout)
	defer cancel()

	srcErr := runAsync(ctx, src)
	time.Sleep(srcWarmup)
	lnkErr := runAsync(ctx, lnk)
	time.Sleep(linkWarmup + measureWindow)

	stats := readLinkStats(t, lnk)
	cancel()

	if err := <-srcErr; err != nil {
		t.Errorf("source: %v", err)
	}
	if err := <-lnkErr; err != nil {
		t.Errorf("link: %v", err)
	}

	if stats.PacketsIn < minPacketsForTest {
		t.Fatalf("link saw only %d packets (want >=%d) — upstream not flowing?",
			stats.PacketsIn, minPacketsForTest)
	}
	rate := float64(stats.PacketsDropped) / float64(stats.PacketsIn)
	if rate < lossLowerBound || rate > lossUpperBound {
		t.Fatalf("loss rate %.3f outside [%.2f,%.2f] (in=%d dropped=%d out=%d)",
			rate, lossLowerBound, lossUpperBound,
			stats.PacketsIn, stats.PacketsDropped, stats.PacketsOut)
	}
}

type linkStats struct {
	PacketsIn      uint64 `json:"packets_in"`
	PacketsOut     uint64 `json:"packets_out"`
	PacketsDropped uint64 `json:"packets_dropped"`
}

func runAsync(ctx context.Context, b block.Block) <-chan error {
	ch := make(chan error, 1)
	go func() { ch <- b.Run(ctx) }()
	return ch
}

func readLinkStats(t *testing.T, b block.Block) linkStats {
	t.Helper()
	var s linkStats
	if err := json.Unmarshal(b.Snapshot().Stats, &s); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	return s
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}
