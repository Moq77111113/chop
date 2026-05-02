// Package process is a generic block wrapping any external CLI as a
// chop block. It captures the wrapped command's stderr into a bounded
// ring exposed via Snapshot, and propagates exit codes so the TUI can
// distinguish clean exits from crashes.
package process

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/moq77111113/chop/block"
)

const (
	stderrRingSize = 100
	stderrTailSize = 20

	eventStarted = "process.started"
	eventExited  = "process.exited"

	shutdownGrace  = 5 * time.Second
	stderrScanMax  = 1024 * 1024
	stderrScanInit = 64 * 1024
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

	cmd := exec.CommandContext(ctx, b.proc.Cmd, b.proc.Args...)
	cmd.Env = mergedEnv(b.proc.Env)
	cmd.Dir = b.proc.Cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = shutdownGrace

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("process: stderr pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("process: stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("process: start %s: %w", b.proc.Cmd, err)
	}
	b.pid.Store(int32(cmd.Process.Pid))
	b.setStatus(block.StatusRunning, nil)
	block.Emit(ctx, eventStarted, map[string]int{"pid": cmd.Process.Pid})

	var drainWG sync.WaitGroup
	drainWG.Add(2)
	go func() { defer drainWG.Done(); drainStderr(stderrPipe, b.stderr) }()
	go func() { defer drainWG.Done(); drainStderr(stdoutPipe, b.stderr) }()

	waitErr := cmd.Wait()
	drainWG.Wait()

	code, status := classifyExit(waitErr)
	b.setStatus(status, &code)
	block.Emit(ctx, eventExited, map[string]int{"code": code})

	// Stay alive after the wrapped CLI exits so Snapshot/Apply keep
	// answering — the supervisor relies on this to surface the final
	// status and to honour Restart. Block until the supervisor cancels
	// the Run context.
	<-ctx.Done()
	return nil
}

func mergedEnv(extra map[string]string) []string {
	if len(extra) == 0 {
		return nil
	}
	base := append([]string{}, os.Environ()...)
	for k, v := range extra {
		base = append(base, k+"="+v)
	}
	return base
}

func drainStderr(r io.Reader, dst *ring) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, stderrScanInit), stderrScanMax)
	for sc.Scan() {
		dst.append(sc.Text())
	}
}

func classifyExit(err error) (code int, status block.Status) {
	if err == nil {
		return 0, block.StatusStopped
	}
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode(), block.StatusDown
	}
	return -1, block.StatusDown
}

func (b *ProcessBlock) setStatus(s block.Status, code *int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = s
	b.exitCode = code
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
