package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad_AcceptsValidMinimalScenario : un YAML correct et minimal passe.
func TestLoad_AcceptsValidMinimalScenario(t *testing.T) {
	p := writeTmp(t, `
name: test
blocks:
  - id: cam-1
    type: source
  - id: link-1
    type: link
`)
	s, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(s.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(s.Blocks))
	}
}

// TestLoad_RejectsDuplicateID : pas deux blocs avec le même id.
func TestLoad_RejectsDuplicateID(t *testing.T) {
	p := writeTmp(t, `
name: test
blocks:
  - id: dup
    type: source
  - id: dup
    type: link
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

// TestLoad_RejectsEmptyBlocks : un scénario vide est invalide.
func TestLoad_RejectsEmptyBlocks(t *testing.T) {
	p := writeTmp(t, `name: test`)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for empty blocks")
	}
}

// TestLoad_PreservesConfigAndControlsAsJSON: les sous-arbres YAML
// `config:` et `controls:` sortent en json.RawMessage utilisable par les
// blocks child via le wire protocol JSON-RPC.
func TestLoad_PreservesConfigAndControlsAsJSON(t *testing.T) {
	p := writeTmp(t, `
name: test
blocks:
  - id: lnk
    type: link
    config:
      upstream: rtsp://127.0.0.1:5101/cam1
      serve_at: 127.0.0.1:8501
    controls:
      loss: 0.1
`)
	s, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var cfg struct {
		Upstream string `json:"upstream"`
		ServeAt  string `json:"serve_at"`
	}
	if err := json.Unmarshal(s.Blocks[0].Config, &cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if cfg.Upstream != "rtsp://127.0.0.1:5101/cam1" {
		t.Fatalf("upstream = %q, want rtsp://127.0.0.1:5101/cam1", cfg.Upstream)
	}
	var ctrl struct {
		Loss float64 `json:"loss"`
	}
	if err := json.Unmarshal(s.Blocks[0].Controls, &ctrl); err != nil {
		t.Fatalf("decode controls: %v", err)
	}
	if ctrl.Loss != 0.1 {
		t.Fatalf("loss = %v, want 0.1", ctrl.Loss)
	}
}

// TestLoad_RejectsBlockMissingID : un bloc sans id est rejeté tôt.
func TestLoad_RejectsBlockMissingID(t *testing.T) {
	p := writeTmp(t, `
name: test
blocks:
  - type: source
`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "id required") {
		t.Fatalf("expected id required error, got %v", err)
	}
}

func writeTmp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "scenario.yml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}
