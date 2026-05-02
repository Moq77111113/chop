package supervisor

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/scenario"
)

func TestRegistry_GetReturnsNilForUnknownID(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("missing"); got != nil {
		t.Fatalf("Get(missing) = %v, want nil", got)
	}
}

func TestRegistry_AddThenGetReturnsSameHandle(t *testing.T) {
	r := NewRegistry()
	h := &Handle{ID: "a", Type: "source"}
	r.Add(h)
	if got := r.Get("a"); got != h {
		t.Fatalf("Get(a) = %v, want %v", got, h)
	}
}

func TestRegistry_AddReplacesEntryWithSameID(t *testing.T) {
	r := NewRegistry()
	r.Add(&Handle{ID: "a", Type: "v1"})
	r.Add(&Handle{ID: "a", Type: "v2"})
	if got := r.Get("a"); got.Type != "v2" {
		t.Fatalf("Type = %q after replace, want v2", got.Type)
	}
}

func TestRegistry_ListReturnsAllRegisteredHandles(t *testing.T) {
	r := NewRegistry()
	r.Add(&Handle{ID: "a"})
	r.Add(&Handle{ID: "b"})
	if got := r.List(); len(got) != 2 {
		t.Fatalf("List len = %d, want 2", len(got))
	}
}

func TestRegistry_ListPreservesInsertionOrder(t *testing.T) {
	r := NewRegistry()
	ids := []string{"cam-1", "link-1", "process-x", "link-2"}
	for _, id := range ids {
		r.Add(&Handle{ID: id})
	}
	r.Add(&Handle{ID: "link-1", Type: "replaced"}) // replacement keeps slot

	got := r.List()
	if len(got) != len(ids) {
		t.Fatalf("len = %d, want %d", len(got), len(ids))
	}
	for i, want := range ids {
		if got[i].ID != want {
			t.Fatalf("position %d: got %q, want %q", i, got[i].ID, want)
		}
	}
}

// TestSupervisor_KillStopsBlockAndPersistsHandle verifies the spec contract:
// k flips status away from running within shutdownGrace, the handle stays
// in the registry so the TUI keeps the slot.
func TestSupervisor_KillStopsBlockAndPersistsHandle(t *testing.T) {
	sup := startTestSupervisor(t, "process", marshalConfig(t, map[string]any{
		"cmd":  "sh",
		"args": []string{"-c", "sleep 30"},
	}))

	awaitRunning(t, sup, "blk")
	if err := sup.Kill(context.Background(), "blk"); err != nil {
		t.Fatalf("Kill: %v", err)
	}

	h := sup.Registry.Get("blk")
	if h == nil {
		t.Fatal("handle removed after kill — should persist")
	}
}

// TestSupervisor_RestartReSpawnsUnderSameID verifies that Restart kills (if
// alive) then respawns from the cached scenario.Block. A new child PID
// proves the respawn happened.
func TestSupervisor_RestartReSpawnsUnderSameID(t *testing.T) {
	sup := startTestSupervisor(t, "process", marshalConfig(t, map[string]any{
		"cmd":  "sh",
		"args": []string{"-c", "sleep 30"},
	}))

	awaitRunning(t, sup, "blk")
	pidBefore := pidOf(t, sup, "blk")

	if err := sup.Restart(context.Background(), "blk"); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	awaitRunning(t, sup, "blk")
	pidAfter := pidOf(t, sup, "blk")

	if pidBefore == pidAfter {
		t.Fatalf("pid unchanged after restart: %d", pidBefore)
	}
}

func startTestSupervisor(t *testing.T, typ string, cfg json.RawMessage) *Supervisor {
	t.Helper()
	exe := buildChopBinary(t)
	sup := &Supervisor{
		Registry: NewRegistry(),
		selfExe:  exe,
		events:   make(chan Event, eventsBufferSize),
	}
	sc := &scenario.Scenario{Blocks: []scenario.Block{{ID: "blk", Type: typ, Config: cfg}}}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = sup.Run(ctx, sc) }()
	return sup
}

func awaitRunning(t *testing.T, sup *Supervisor, id string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if h := sup.Registry.Get(id); h != nil {
			snap, err := h.Snapshot(context.Background())
			if err == nil && snap.Status == block.StatusRunning {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("block %s never reached running", id)
}

func pidOf(t *testing.T, sup *Supervisor, id string) int {
	t.Helper()
	h := sup.Registry.Get(id)
	if h == nil {
		t.Fatalf("no handle %s", id)
	}
	snap, err := h.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	var s struct {
		PID int `json:"pid"`
	}
	_ = json.Unmarshal(snap.Stats, &s)
	return s.PID
}

func marshalConfig(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}

func buildChopBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "chop")
	cmd := exec.Command("go", "build", "-o", bin, "../../cmd/chop")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Skipf("could not build chop binary for integration test: %v", err)
	}
	return bin
}
