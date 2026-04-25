// Package supervisor orchestrates the lifecycle of chop blocks: spawns each
// as a child OS process, wires stdio JSON-RPC, exposes a registry, drives
// shutdown.
package supervisor

import (
	"context"
	"fmt"
	"os"

	"github.com/moq77111113/chop/internal/scenario"
)

// Supervisor orchestrates a scenario: spawns each block as a child process,
// keeps a registry, propagates shutdown when its context is cancelled.
type Supervisor struct {
	Registry *Registry
	selfExe  string
}

// New constructs a Supervisor bound to the current executable. The same
// binary is re-used to spawn each block child via `chop block <type>`.
func New() (*Supervisor, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return &Supervisor{Registry: NewRegistry(), selfExe: exe}, nil
}

// Run spawns each block declared in the scenario in declaration order
// (M1: no DAG yet), then waits for ctx to be cancelled. On any spawn error
// already-started children are torn down before returning.
func (s *Supervisor) Run(ctx context.Context, sc *scenario.Scenario) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, b := range sc.Blocks {
		if err := s.spawn(runCtx, b); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return nil
}

func (s *Supervisor) spawn(ctx context.Context, b scenario.Block) error {
	ch, err := spawnChild(ctx, s.selfExe, b.Type, b.ID, b.Config, b.Controls)
	if err != nil {
		return fmt.Errorf("spawn %s: %w", b.ID, err)
	}
	s.Registry.Add(&Handle{ID: b.ID, Type: b.Type, child: ch})
	go func() { _ = ch.rpc.Serve(ctx) }()
	return nil
}
