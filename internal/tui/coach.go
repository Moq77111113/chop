package tui

import (
	"os"
	"path/filepath"
)

const (
	seenFileName  = "seen"
	chopConfigDir = "chop"
)

// shouldShowCoach returns true when no seen-marker exists in the user's
// config dir. Errors are swallowed: if we can't read the marker, we err
// on the side of NOT showing the coach (a missing config dir means we'd
// fail to write the marker too — better to skip than to coach forever).
func shouldShowCoach() bool {
	path, err := seenFilePath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return false
	}
	return true
}

// markSeen writes the seen-marker so the coach won't return on the next
// run. Best-effort: a write failure (read-only home, etc.) is tolerated —
// the user just sees the coach a second time.
func markSeen() {
	path, err := seenFilePath()
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte("1"), 0o644)
}

func seenFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, chopConfigDir, seenFileName), nil
}
