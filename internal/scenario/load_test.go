package scenario

import (
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
