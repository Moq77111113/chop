package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/moq77111113/chop/internal/supervisor"
)

const (
	eventsBufferSize = 50
	eventsRender     = 6
)

// uiEvent is the rendered form of a supervisor event — pre-derived level
// and pre-formatted message so the renderer is pure formatting.
type uiEvent struct {
	when    time.Time
	level   string
	source  string
	blockID string
	message string
}

// EventMsg carries a supervisor event into the bubbletea Update loop.
// Exported so cmd/chop can forward events from the supervisor's bus.
type EventMsg struct{ Event supervisor.Event }

func (a *App) appendEvent(ev supervisor.Event) {
	ui := decodeEvent(ev)
	a.data.events = append(a.data.events, ui)
	if len(a.data.events) > eventsBufferSize {
		a.data.events = a.data.events[len(a.data.events)-eventsBufferSize:]
	}
}

func decodeEvent(ev supervisor.Event) uiEvent {
	when := time.UnixMilli(ev.TsMs)
	if ev.TsMs == 0 {
		when = time.Now()
	}
	level, source := levelAndSource(ev.Kind)
	return uiEvent{
		when:    when,
		level:   level,
		source:  source,
		blockID: ev.BlockID,
		message: formatEventMessage(ev.Kind, ev.Payload, ev.BlockID),
	}
}

func levelAndSource(kind string) (level, source string) {
	parts := strings.SplitN(kind, ".", 2)
	src := parts[0]
	switch {
	case strings.Contains(kind, "dropped"), strings.HasSuffix(kind, ".error"):
		return "ERR", src
	case strings.HasSuffix(kind, ".warn"), strings.Contains(kind, "spike"):
		return "WRN", src
	}
	return "INF", src
}

func formatEventMessage(kind string, payload json.RawMessage, blockID string) string {
	switch kind {
	case "rtp.dropped":
		var p struct {
			Reason string `json:"reason"`
			Seq    uint16 `json:"seq"`
		}
		_ = json.Unmarshal(payload, &p)
		return fmt.Sprintf("dropped pkt seq %d (%s) on %s", p.Seq, p.Reason, blockID)
	case "link.up":
		return blockID + " up"
	case "source.up":
		return blockID + " up"
	}
	if len(payload) > 0 && string(payload) != "null" {
		return kind + " " + string(payload)
	}
	return kind
}

type eventStyles struct {
	header lipgloss.Style
	time   lipgloss.Style
	source lipgloss.Style
	body   lipgloss.Style
	inf    lipgloss.Style
	wrn    lipgloss.Style
	err    lipgloss.Style
}

func newEventStyles(t Theme) eventStyles {
	return eventStyles{
		header: lipgloss.NewStyle().Foreground(t.Muted).Bold(true).MarginTop(1),
		time:   lipgloss.NewStyle().Foreground(t.Dim),
		source: lipgloss.NewStyle().Foreground(t.Muted),
		body:   lipgloss.NewStyle().Foreground(t.Fg),
		inf:    lipgloss.NewStyle().Foreground(t.Info).Bold(true),
		wrn:    lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		err:    lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
	}
}

func (s eventStyles) levelTag(level string) string {
	switch level {
	case "WRN":
		return s.wrn.Render(level)
	case "ERR":
		return s.err.Render(level)
	}
	return s.inf.Render(level)
}

// renderEventsTicker draws the bottom-of-left-pane events list. Returns
// empty string when no events have been seen yet.
func renderEventsTicker(events []uiEvent, s eventStyles, _ int) string {
	if len(events) == 0 {
		return ""
	}
	start := max(0, len(events)-eventsRender)
	tail := events[start:]
	lines := []string{s.header.Render("EVENTS · GLOBAL")}
	for _, ev := range tail {
		ts := s.time.Render(ev.when.Format("15:04:05"))
		lvl := s.levelTag(ev.level)
		src := s.source.Render(ev.source)
		body := s.body.Render(ev.message)
		lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", ts, lvl, src, body))
	}
	return strings.Join(lines, "\n")
}
