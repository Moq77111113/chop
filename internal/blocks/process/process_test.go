package process_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

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

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}
