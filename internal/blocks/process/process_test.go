package process_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/blocks/process"
)

func TestNew_MissingCmdSurfacesOnRun(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{}),
	})
	err := b.Run(context.Background())
	if err == nil || err.Error() != "process: cmd is required" {
		t.Fatalf("Run err = %v, want cmd-required", err)
	}
}

func TestSnapshot_BeforeRunReportsStoppedZeroPid(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{Cmd: "true"}),
	})
	snap := b.Snapshot()
	if snap.Status != block.StatusStopped {
		t.Fatalf("status = %q, want stopped", snap.Status)
	}
	var s struct {
		PID        int      `json:"pid"`
		ExitCode   *int     `json:"exit_code"`
		StderrTail []string `json:"stderr_tail"`
	}
	if err := json.Unmarshal(snap.Stats, &s); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if s.PID != 0 || s.ExitCode != nil || len(s.StderrTail) != 0 {
		t.Fatalf("pre-run stats = %+v, want zero", s)
	}
}

func TestApply_AlwaysAcksWithoutMutating(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{Cmd: "true"}),
	})
	if err := b.Apply(json.RawMessage(`{"anything":1}`)); err != nil {
		t.Fatalf("Apply err = %v, want nil", err)
	}
}

func TestRun_RejectsParseErrorWithSentinel(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: json.RawMessage(`{"cmd":""}`),
	})
	err := b.Run(context.Background())
	if err == nil || !errors.Is(err, process.ErrMissingCmd) {
		t.Fatalf("err = %v, want ErrMissingCmd", err)
	}
}

func TestRun_TrueExitsCleanlyAndReportsStopped(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{Cmd: "true"}),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Run(ctx); err != nil {
		t.Fatalf("Run err = %v, want nil", err)
	}
	snap := b.Snapshot()
	if snap.Status != block.StatusStopped {
		t.Fatalf("status = %q, want stopped", snap.Status)
	}
	var s struct {
		ExitCode *int `json:"exit_code"`
	}
	_ = json.Unmarshal(snap.Stats, &s)
	if s.ExitCode == nil || *s.ExitCode != 0 {
		t.Fatalf("exit_code = %v, want 0", s.ExitCode)
	}
}

func TestRun_FalseExitsNonZeroAndReportsDown(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{Cmd: "false"}),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = b.Run(ctx)
	snap := b.Snapshot()
	if snap.Status != block.StatusDown {
		t.Fatalf("status = %q, want down", snap.Status)
	}
	var s struct {
		ExitCode *int `json:"exit_code"`
	}
	_ = json.Unmarshal(snap.Stats, &s)
	if s.ExitCode == nil || *s.ExitCode == 0 {
		t.Fatalf("exit_code = %v, want non-zero", s.ExitCode)
	}
}

func TestRun_CapturesStderrIntoRing(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{
			Cmd:  "sh",
			Args: []string{"-c", "echo line-one >&2; echo line-two >&2"},
		}),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = b.Run(ctx)
	snap := b.Snapshot()
	var s struct {
		StderrTail []string `json:"stderr_tail"`
	}
	_ = json.Unmarshal(snap.Stats, &s)
	if len(s.StderrTail) < 2 || s.StderrTail[len(s.StderrTail)-2] != "line-one" || s.StderrTail[len(s.StderrTail)-1] != "line-two" {
		t.Fatalf("stderr tail = %v, want [..., line-one, line-two]", s.StderrTail)
	}
}

func TestRun_CtxCancelStopsRunningChild(t *testing.T) {
	b := process.New(block.Config{
		ID: "p", Type: "process",
		Static: mustJSON(t, process.Config{
			Cmd:  "sh",
			Args: []string{"-c", "sleep 30"},
		}),
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.Run(ctx) }()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(7 * time.Second):
		t.Fatal("Run did not return within 7s of ctx cancel")
	}
	if status := b.Snapshot().Status; status == block.StatusRunning {
		t.Fatalf("status still running after cancel")
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
