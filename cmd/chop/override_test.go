package main

import (
	"encoding/json"
	"testing"

	"github.com/moq77111113/chop/internal/scenario"
)

func TestParseOverride_LossAsPercent(t *testing.T) {
	ov, err := parseOverride("link-1:loss=12%")
	if err != nil {
		t.Fatal(err)
	}
	if ov.ID != "link-1" {
		t.Fatalf("id = %q, want link-1", ov.ID)
	}
	if got := ov.Controls["loss"]; got != 0.12 {
		t.Fatalf("loss = %v, want 0.12", got)
	}
}

func TestParseOverride_LatencyJitterMs(t *testing.T) {
	ov, err := parseOverride("link-1:latency=200ms,jitter=45ms")
	if err != nil {
		t.Fatal(err)
	}
	if got := ov.Controls["latency_ms"]; got != uint32(200) {
		t.Fatalf("latency_ms = %v, want 200", got)
	}
	if got := ov.Controls["jitter_ms"]; got != uint32(45) {
		t.Fatalf("jitter_ms = %v, want 45", got)
	}
}

func TestParseOverride_BandwidthMbps(t *testing.T) {
	ov, err := parseOverride("link-1:bw=1.5M")
	if err != nil {
		t.Fatal(err)
	}
	if got := ov.Controls["bandwidth_kbps"]; got != uint32(1500) {
		t.Fatalf("bandwidth_kbps = %v, want 1500", got)
	}
}

func TestParseOverride_BandwidthOffMapsToZero(t *testing.T) {
	ov, err := parseOverride("link-1:bw=off")
	if err != nil {
		t.Fatal(err)
	}
	if got := ov.Controls["bandwidth_kbps"]; got != 0 {
		t.Fatalf("bandwidth_kbps = %v, want 0", got)
	}
}

func TestParseOverride_RejectsMissingColon(t *testing.T) {
	if _, err := parseOverride("link-1 loss=12%"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestApplyOverrides_MergesIntoExistingControls(t *testing.T) {
	sc := &scenario.Scenario{
		Blocks: []scenario.Block{
			{ID: "link-1", Type: "link", Controls: json.RawMessage(`{"loss":0.0,"latency_ms":50}`)},
		},
	}
	if err := applyOverrides(sc, []Override{
		{ID: "link-1", Controls: map[string]any{"loss": 0.10, "jitter_ms": uint32(20)}},
	}); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(sc.Blocks[0].Controls, &got); err != nil {
		t.Fatal(err)
	}
	if got["loss"] != 0.10 {
		t.Fatalf("loss = %v, want 0.10 (overridden)", got["loss"])
	}
	if got["latency_ms"] != float64(50) {
		t.Fatalf("latency_ms = %v, want 50 (preserved)", got["latency_ms"])
	}
}

func TestApplyOverrides_UnknownIDIsAnError(t *testing.T) {
	sc := &scenario.Scenario{Blocks: []scenario.Block{{ID: "link-1"}}}
	err := applyOverrides(sc, []Override{{ID: "nope", Controls: map[string]any{"loss": 0.5}}})
	if err == nil {
		t.Fatal("expected error for unknown id")
	}
}
