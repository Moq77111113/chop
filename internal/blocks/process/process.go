// Package process is a generic block wrapping any external CLI as a
// chop block. It captures the wrapped command's stderr into a bounded
// ring exposed via Snapshot, and propagates exit codes so the TUI can
// distinguish clean exits from crashes.
package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/moq77111113/chop/block"
)

const (
	stderrRingSize = 100
	stderrTailSize = 20

	eventStarted = "process.started"
	eventExited  = "process.exited"
)

// ErrMissingCmd is the sentinel returned when a process block declaration
// omits its required `cmd:` field.
var ErrMissingCmd = errors.New("process: cmd is required")

// Config is the static configuration of a process block, parsed from the
// scenario YAML `config:` section. cmd is required; everything else
// defaults to an empty value (no args, inherited env, current cwd).
type Config struct {
	Cmd  string            `json:"cmd"`
	Args []string          `json:"args,omitempty"`
	Env  map[string]string `json:"env,omitempty"`
	Cwd  string            `json:"cwd,omitempty"`
}

type stats struct {
	PID        int      `json:"pid"`
	ExitCode   *int     `json:"exit_code"`
	StderrTail []string `json:"stderr_tail"`
}

// ProcessBlock wraps an external CLI as a supervised child. The wrapped
// command's stderr is drained into a fixed-size ring and surfaced on
// every Snapshot.
type ProcessBlock struct {
	cfg      block.Config
	proc     Config
	parseErr error

	stderr *ring

	pid atomic.Int32

	mu       sync.Mutex
	exitCode *int
	status   block.Status
}

// New is the process block factory. Configuration errors are deferred
// to Run so the runtime can construct the block unconditionally.
func New(c block.Config) block.Block {
	cfg, err := parseConfig(c.Static)
	return &ProcessBlock{
		cfg:      c,
		proc:     cfg,
		parseErr: err,
		stderr:   newRing(stderrRingSize),
		status:   block.StatusStopped,
	}
}

func (b *ProcessBlock) Info() block.Info {
	return block.Info{ID: b.cfg.ID, Type: b.cfg.Type, Config: b.cfg.Static}
}

func (b *ProcessBlock) Snapshot() block.Snapshot {
	b.mu.Lock()
	st, code := b.status, b.exitCode
	b.mu.Unlock()
	payload, _ := json.Marshal(stats{
		PID:        int(b.pid.Load()),
		ExitCode:   code,
		StderrTail: b.stderr.tail(stderrTailSize),
	})
	return block.Snapshot{Status: st, Stats: payload, TsMs: time.Now().UnixMilli()}
}

func (b *ProcessBlock) Apply(json.RawMessage) error          { return nil }
func (b *ProcessBlock) Action(string, json.RawMessage) error { return nil }

func (b *ProcessBlock) Run(ctx context.Context) error {
	if b.parseErr != nil {
		return b.parseErr
	}
	return errors.New("process: Run not yet implemented")
}

func parseConfig(raw json.RawMessage) (Config, error) {
	var c Config
	if len(raw) == 0 {
		return Config{}, ErrMissingCmd
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("process: parse config: %w", err)
	}
	if c.Cmd == "" {
		return Config{}, ErrMissingCmd
	}
	return c, nil
}
