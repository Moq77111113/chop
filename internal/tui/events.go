package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/tui/components/events"
)

const eventsBufferSize = 50

// uiEvent is the rendered form of a supervisor event — pre-derived
// level and pre-formatted message so the renderer is pure formatting.
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
	a.events = append(a.events, ui)
	if len(a.events) > eventsBufferSize {
		a.events = a.events[len(a.events)-eventsBufferSize:]
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
	case kind == "process.exited":
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
	case "process.started":
		var p struct {
			PID int `json:"pid"`
		}
		_ = json.Unmarshal(payload, &p)
		return fmt.Sprintf("%s started (pid %d)", blockID, p.PID)
	case "process.exited":
		var p struct {
			Code int `json:"code"`
		}
		_ = json.Unmarshal(payload, &p)
		return fmt.Sprintf("%s exited (code %d)", blockID, p.Code)
	case "block.kill_requested":
		return blockID + " kill requested"
	case "block.restarted":
		return blockID + " restarted"
	}
	if len(payload) > 0 && string(payload) != "null" {
		return kind + " " + string(payload)
	}
	return kind
}

// toEventList projects the App's stored uiEvent slice into the
// component's input shape. Cheap copy per render — the buffer is
// capped at eventsBufferSize.
func toEventList(src []uiEvent) []events.Event {
	out := make([]events.Event, len(src))
	for i, e := range src {
		out[i] = events.Event{
			When:    e.when,
			Level:   e.level,
			Source:  e.source,
			Message: e.message,
		}
	}
	return out
}
