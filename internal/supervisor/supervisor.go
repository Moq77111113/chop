// Package supervisor orchestrates the lifecycle of chop blocks: spawns each
// as a child OS process, wires stdio JSON-RPC, exposes a registry, drives
// shutdown, and supports per-block Kill / Restart at runtime.
package supervisor

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/moq77111113/chop/internal/scenario"
)

// Supervisor orchestrates a scenario: spawns each block as a child process,
// keeps a registry, propagates shutdown when its context is cancelled.
type Supervisor struct {
	Registry *Registry
	selfExe  string
	events   chan Event

	mu     sync.Mutex
	runCtx context.Context
}

// New constructs a Supervisor bound to the current executable. The same
// binary is re-used to spawn each block child via `chop block <type>`.
func New() (*Supervisor, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return &Supervisor{
		Registry: NewRegistry(),
		selfExe:  exe,
		events:   make(chan Event, eventsBufferSize),
	}, nil
}

// Run spawns each block declared in the scenario in declaration order
// (M1: no DAG yet), then waits for ctx to be cancelled.
func (s *Supervisor) Run(ctx context.Context, sc *scenario.Scenario) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.mu.Lock()
	s.runCtx = runCtx
	s.mu.Unlock()

	for _, b := range sc.Blocks {
		if err := s.spawn(runCtx, b); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return nil
}

// Kill terminates the block under id and leaves the handle in place so the
// TUI keeps showing the slot. Status will flip on the next snapshot — the
// supervisor itself only reports the request via the events bus.
func (s *Supervisor) Kill(_ context.Context, id string) error {
	h := s.Registry.Get(id)
	if h == nil {
		return errBlockNotFound(id)
	}
	s.emit(Event{BlockID: id, Kind: EventBlockKillRequested, TsMs: nowMs()})
	h.stop()
	return nil
}

// Restart kills the block (if alive) then re-spawns it from the cached
// scenario.Block under the same id. Live knob changes are not preserved
// in A1 — the YAML's initial controls are re-injected.
func (s *Supervisor) Restart(_ context.Context, id string) error {
	h := s.Registry.Get(id)
	if h == nil {
		return errBlockNotFound(id)
	}
	decl := h.block
	h.stop()

	s.mu.Lock()
	runCtx := s.runCtx
	s.mu.Unlock()
	if runCtx == nil {
		return fmt.Errorf("supervisor: not running")
	}
	if err := s.spawn(runCtx, decl); err != nil {
		return fmt.Errorf("respawn %s: %w", id, err)
	}
	s.emit(Event{BlockID: id, Kind: EventBlockRestarted, TsMs: nowMs()})
	return nil
}

func (s *Supervisor) spawn(parent context.Context, b scenario.Block) error {
	childCtx, cancel := context.WithCancel(parent)
	ch, err := spawnChild(childCtx, s.selfExe, b.Type, b.ID, b.Config, b.Controls)
	if err != nil {
		cancel()
		return fmt.Errorf("spawn %s: %w", b.ID, err)
	}
	s.forwardEvents(b.ID, ch.rpc)
	s.Registry.Add(&Handle{
		ID:     b.ID,
		Type:   b.Type,
		block:  b,
		child:  ch,
		cancel: cancel,
	})
	go func() { _ = ch.rpc.Serve(childCtx) }()
	return nil
}

func (s *Supervisor) emit(ev Event) {
	select {
	case s.events <- ev:
	default:
	}
}
