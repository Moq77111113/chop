// Package events renders the bottom-of-left-pane events ticker: most
// recent N entries with timestamp, level pill, source, and message.
package events

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/tui/ui"
)

const eventsRender = 6

// Event is the rendered form of a supervisor event — pre-formatted
// message, derived level. The bubbletea adapter converts wire events
// into this shape before passing them to Render.
type Event struct {
	When    time.Time
	Level   string // "INF" | "WRN" | "ERR"
	Source  string
	Message string
}

// Styles bundles the per-column palette.
type Styles struct {
	Header lipgloss.Style
	Time   lipgloss.Style
	Source lipgloss.Style
	Body   lipgloss.Style
	Inf    lipgloss.Style
	Wrn    lipgloss.Style
	Err    lipgloss.Style
}

// Render draws the ticker. Returns "" when no events have been seen
// yet so the caller can omit the section entirely.
func Render(events []Event, s Styles, _ int) string {
	if len(events) == 0 {
		return ""
	}
	start := max(0, len(events)-eventsRender)
	tail := events[start:]
	lines := []string{ui.Header("EVENTS · GLOBAL", s.Header)}
	for _, ev := range tail {
		ts := s.Time.Render(ev.When.Format("15:04:05"))
		lvl := levelTag(ev.Level, s)
		src := s.Source.Render(ev.Source)
		body := s.Body.Render(ev.Message)
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", ts, lvl, src, body))
	}
	return strings.Join(lines, "\n")
}

func levelTag(level string, s Styles) string {
	switch level {
	case "WRN":
		return s.Wrn.Render(level)
	case "ERR":
		return s.Err.Render(level)
	}
	return s.Inf.Render(level)
}
