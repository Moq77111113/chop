package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/moq77111113/chop/internal/scenario"
)

// Override is a CLI-supplied patch applied to a block's controls before
// the scenario starts. The shape mirrors the YAML scenario but only the
// fields a user might tweak from the command line are exposed.
type Override struct {
	ID       string
	Controls map[string]any
}

// parseOverrides parses every --override flag value. Each value is of
// form `<block-id>:<knob>=<val>[,<knob>=<val>...]`.
func parseOverrides(values []string) ([]Override, error) {
	out := make([]Override, 0, len(values))
	for _, v := range values {
		ov, err := parseOverride(v)
		if err != nil {
			return nil, fmt.Errorf("--override %q: %w", v, err)
		}
		out = append(out, ov)
	}
	return out, nil
}

func parseOverride(raw string) (Override, error) {
	id, body, ok := strings.Cut(raw, ":")
	if !ok || id == "" {
		return Override{}, fmt.Errorf("expected <id>:<knob>=<val>")
	}
	ctrls := map[string]any{}
	for _, pair := range strings.Split(body, ",") {
		key, val, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok {
			return Override{}, fmt.Errorf("expected <knob>=<val>, got %q", pair)
		}
		jsonKey, jsonVal, err := decodeKnob(key, val)
		if err != nil {
			return Override{}, err
		}
		ctrls[jsonKey] = jsonVal
	}
	return Override{ID: id, Controls: ctrls}, nil
}

// decodeKnob translates a human-friendly knob/value pair (e.g. "loss=12%")
// into the JSON key + numeric value the link block consumes ("loss", 0.12).
func decodeKnob(key, val string) (string, any, error) {
	const (
		percentScale = 100.0
		mbpsToKbps   = 1000.0
	)
	switch key {
	case "loss":
		f, err := parsePercent(val)
		if err != nil {
			return "", nil, fmt.Errorf("loss: %w", err)
		}
		return "loss", f / percentScale, nil
	case "latency", "lat":
		ms, err := parseDurationMs(val)
		if err != nil {
			return "", nil, fmt.Errorf("latency: %w", err)
		}
		return "latency_ms", ms, nil
	case "jitter", "jit":
		ms, err := parseDurationMs(val)
		if err != nil {
			return "", nil, fmt.Errorf("jitter: %w", err)
		}
		return "jitter_ms", ms, nil
	case "bandwidth", "bw":
		if val == "off" || val == "∞" {
			return "bandwidth_kbps", 0, nil
		}
		mbps, err := parseSuffixedFloat(val, "M")
		if err != nil {
			return "", nil, fmt.Errorf("bandwidth: %w", err)
		}
		return "bandwidth_kbps", uint32(mbps * mbpsToKbps), nil
	}
	return "", nil, fmt.Errorf("unknown knob %q (want loss|latency|jitter|bandwidth)", key)
}

func parsePercent(s string) (float64, error) {
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseFloat(s, 64)
}

func parseDurationMs(s string) (uint32, error) {
	s = strings.TrimSuffix(s, "ms")
	s = strings.TrimSuffix(s, "±") // e.g. ±50ms
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func parseSuffixedFloat(s, suffix string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSuffix(s, suffix), 64)
}

// applyOverrides merges each override into the matching block's Controls.
// Unknown ids are reported as errors so a typo in the flag isn't silently
// swallowed.
func applyOverrides(sc *scenario.Scenario, overrides []Override) error {
	for _, ov := range overrides {
		idx := indexOfBlock(sc.Blocks, ov.ID)
		if idx < 0 {
			return fmt.Errorf("--override: unknown block id %q", ov.ID)
		}
		current := decodeExistingControls(sc.Blocks[idx].Controls)
		for k, v := range ov.Controls {
			current[k] = v
		}
		merged, err := json.Marshal(current)
		if err != nil {
			return fmt.Errorf("--override %s: marshal: %w", ov.ID, err)
		}
		sc.Blocks[idx].Controls = merged
	}
	return nil
}

func indexOfBlock(blocks []scenario.Block, id string) int {
	for i, b := range blocks {
		if b.ID == id {
			return i
		}
	}
	return -1
}

func decodeExistingControls(raw json.RawMessage) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}
