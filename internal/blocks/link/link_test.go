package link

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/moq77111113/chop/block"
)

func TestNew_MissingUpstreamSurfacesOnRun(t *testing.T) {
	b := New(block.Config{
		ID: "lnk", Type: "link",
		Static: mustJSON(t, Config{ServeAt: "127.0.0.1:8551"}),
	})
	err := b.Run(t.Context())
	if !errors.Is(err, errMissingUpstream) {
		t.Fatalf("want errMissingUpstream, got %v", err)
	}
}

func TestNew_MissingServeAtSurfacesOnRun(t *testing.T) {
	b := New(block.Config{
		ID: "lnk", Type: "link",
		Static: mustJSON(t, Config{Upstream: "rtsp://127.0.0.1:5101/stream"}),
	})
	err := b.Run(t.Context())
	if !errors.Is(err, errMissingServeAt) {
		t.Fatalf("want errMissingServeAt, got %v", err)
	}
}

func TestApply_ReplacesControlsAtomically(t *testing.T) {
	b := New(block.Config{
		ID: "lnk", Type: "link",
		Static: mustJSON(t, Config{Upstream: "rtsp://x", ServeAt: "x"}),
		Live:   mustJSON(t, Controls{Loss: 0}),
	}).(*LinkBlock)

	if err := b.Apply(mustJSON(t, Controls{Loss: 0.5, LatencyMs: 50})); err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := b.ctrls.Load()
	if got.Loss != 0.5 || got.LatencyMs != 50 {
		t.Fatalf("controls = %+v, want Loss=0.5 LatencyMs=50", got)
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}
