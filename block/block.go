// Package block is the public framework surface for chop block implementations.
// External devs (Go, or via the wire protocol from Rust/Python) implement the
// Block interface and call RunBlock to plug into a chop supervisor.
package block

import (
	"context"
	"encoding/json"
)

// Status is the live operating state of a block, reported in every Snapshot.
type Status string

const (
	StatusRunning  Status = "running"
	StatusDegraded Status = "degraded"
	StatusStopped  Status = "stopped"
)

// Info is the static identity of a block, returned to the supervisor on demand.
// The Config field carries the original YAML `config:` section verbatim.
type Info struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

// Snapshot is the live observation of a block at a point in time.
// Stats is a block-specific JSON payload (each block defines its own shape).
type Snapshot struct {
	Status Status          `json:"status"`
	Stats  json.RawMessage `json:"stats"`
	TsMs   int64           `json:"ts"`
}

// Config is what RunBlock parses from argv and hands to the factory.
// Static maps to the YAML `config:` section, Live to `controls:` (initial values).
type Config struct {
	ID     string
	Type   string
	Static json.RawMessage
	Live   json.RawMessage
}

// Block is the contract every chop block implementation must satisfy.
// Implementations are constructed by a Factory, then driven by RunBlock:
// the supervisor calls Info/Snapshot/Apply/Action over JSON-RPC, and Run
// executes the block's effects until the context is cancelled.
type Block interface {
	Info() Info
	Snapshot() Snapshot
	Apply(controls json.RawMessage) error
	Action(name string, args json.RawMessage) error
	Run(ctx context.Context) error
}
