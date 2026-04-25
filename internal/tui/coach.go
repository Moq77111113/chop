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

// coachHint is one balloon's contents: a one-line headline and a
// secondary cue that names the relevant key. Matches the three-balloon
// pattern in design/screens/05-coach.png.
type coachHint struct {
	title, cue string
}

// coachHints are the three first-run tips, ordered by the exploration
// path: pick a link → adjust a knob → escape the wall of perturbations.
var coachHints = []coachHint{
	{"this is the link you're poking", "↕ to pick another"},
	{"the focused knob — ↔ adjusts, tab cycles", ""},
	{"visible undo. always here.", "press any key to dismiss"},
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

// renderCoach paints three stacked green balloons centered over the body.
// Anchored callouts (per design/screens/05-coach.png) would need geometry
// tracking on every framed element — deferred. The stacked variant carries
// the same copy and dismissal semantics with one tenth of the layout work.
func renderCoach(width, height int, theme Theme) string {
	balloonW := min(coachWidthMax, max(width-4, 12))
	balloons := make([]string, 0, len(coachHints))
	for i, h := range coachHints {
		balloons = append(balloons, renderCoachBalloon(i+1, h, theme, balloonW))
	}
	stack := strings.Join(balloons, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, stack)
}

// renderCoachBalloon draws one balloon: numbered chip on the left, the
// hint title and optional cue stacked on the right. Background fills so
// the green call-out reads as a single solid shape.
func renderCoachBalloon(n int, h coachHint, theme Theme, width int) string {
	chip := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Background(theme.Bg).
		Bold(true).
		Padding(0, 1).
		Render(numberedPrefix(n))
	titleStyle := lipgloss.NewStyle().Foreground(theme.Bg).Background(theme.Primary).Bold(true)
	cueStyle := lipgloss.NewStyle().Foreground(theme.Bg).Background(theme.Primary).Italic(true)
	body := titleStyle.Render(h.title)
	if h.cue != "" {
		body += "\n" + cueStyle.Render(h.cue)
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, chip, " "+body)
	return lipgloss.NewStyle().
		Background(theme.Primary).
		Padding(0, 1).
		Width(width).
		Render(row)
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
