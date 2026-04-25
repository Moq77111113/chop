package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	seenFileName  = "seen"
	chopConfigDir = "chop"
	coachWidthMax = 50
)

// coachHints are the three numbered tips first-run users see. Short,
// imperative, ordered by the natural exploration path.
var coachHints = []string{
	"press ↵ or tab to drill into the knobs",
	"press ←→ to push a knob, watch the picture react",
	"press y to copy the current perturbation as a --override flag",
}

// shouldShowCoach returns true when no seen-marker exists in the user's
// config dir. Errors are swallowed: if we can't read the marker, we err on
// the side of showing the coach once.
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

// renderCoach draws the centered coach card over the body area.
func renderCoach(width, height int, theme Theme) string {
	cardW := min(coachWidthMax, max(width-4, 10))
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 2).
		Width(cardW)

	title := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true).Render("welcome to chop")
	subtitle := lipgloss.NewStyle().Foreground(theme.Dim).Italic(true).Render("any key dismisses this card")
	step := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	body := lipgloss.NewStyle().Foreground(theme.Fg)

	lines := []string{title, ""}
	for i, hint := range coachHints {
		lines = append(lines, step.Render(numberedPrefix(i+1))+" "+body.Render(hint))
	}
	lines = append(lines, "", subtitle)

	card := style.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}

func numberedPrefix(n int) string {
	switch n {
	case 1:
		return "①"
	case 2:
		return "②"
	case 3:
		return "③"
	}
	return ""
}
