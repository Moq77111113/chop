// Package data is the TUI-local projection of supervisor wire types.
// It owns JSON decode, rate math, knob-value formatting, and link-title
// derivation. It imports no lipgloss or bubbletea — components consume
// these types and decide how to display them.
package data

import (
	"encoding/json"

	"github.com/moq77111113/chop/block"
)

// LinkState is the UI-side projection of block.Status used by the link
// list and the components that decorate its rows. Kept as a string so a
// renderer can spell the label directly without a lookup table.
type LinkState string

// LinkState constants: the five values a link row can report.
const (
	StateUp       LinkState = "UP"
	StateDegraded LinkState = "DEGRADED"
	StateDown     LinkState = "DOWN"
	StateStarting LinkState = "STARTING"
	StateStopped  LinkState = "STOPPED"
)

// MapState converts a wire-side block.Status into the LinkState enum.
// Unknown statuses fold to StateStarting so the row reads as "still
// coming up" rather than disappearing.
func MapState(s block.Status) LinkState {
	switch s {
	case block.StatusRunning:
		return StateUp
	case block.StatusDegraded:
		return StateDegraded
	case block.StatusStopped:
		return StateStopped
	case block.StatusDown:
		return StateDown
	}
	return StateStarting
}

// LinkSnapshot is the TUI-local projection of a link block's snapshot
// stats. Unmarshaled from the wire JSON so the TUI doesn't import the
// block packages directly.
type LinkSnapshot struct {
	Status     block.Status
	PacketsIn  uint64 `json:"packets_in"`
	PacketsOut uint64 `json:"packets_out"`
	UpSinceMs  int64  `json:"up_since_ms"`
	Controls   struct {
		Loss          float64 `json:"loss"`
		LatencyMs     uint32  `json:"latency_ms"`
		JitterMs      uint32  `json:"jitter_ms"`
		BandwidthKbps uint32  `json:"bandwidth_kbps"`
	} `json:"controls"`
}

// SourceSnapshot is the TUI-local projection of a source block's snapshot.
type SourceSnapshot struct {
	RTPServed int64 `json:"rtp_served"`
	UpSinceMs int64 `json:"up_since_ms"`
}

// DecodeLink decodes a block.Snapshot into a LinkSnapshot. The wire
// Status is copied across so callers don't have to thread it
// separately. Returns ok=false when the JSON payload is malformed.
func DecodeLink(snap block.Snapshot) (LinkSnapshot, bool) {
	var ls LinkSnapshot
	if err := json.Unmarshal(snap.Stats, &ls); err != nil {
		return LinkSnapshot{}, false
	}
	ls.Status = snap.Status
	return ls, true
}

// DecodeSource decodes a block.Snapshot into a SourceSnapshot. Returns
// ok=false when the JSON payload is malformed.
func DecodeSource(snap block.Snapshot) (SourceSnapshot, bool) {
	var ss SourceSnapshot
	if err := json.Unmarshal(snap.Stats, &ss); err != nil {
		return SourceSnapshot{}, false
	}
	return ss, true
}

// ProcessSnapshot is the TUI-local projection of a process block's snapshot.
type ProcessSnapshot struct {
	Status     block.Status
	PID        int      `json:"pid"`
	ExitCode   *int     `json:"exit_code"`
	StderrTail []string `json:"stderr_tail"`
}

// DecodeProcess decodes a block.Snapshot into a ProcessSnapshot. Returns
// ok=false when the JSON payload is malformed.
func DecodeProcess(snap block.Snapshot) (ProcessSnapshot, bool) {
	var ps ProcessSnapshot
	if err := json.Unmarshal(snap.Stats, &ps); err != nil {
		return ProcessSnapshot{}, false
	}
	ps.Status = snap.Status
	return ps, true
}
