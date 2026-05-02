package supervisor

import (
	"encoding/json"
	"time"

	"github.com/moq77111113/chop/internal/transport"
)

const (
	eventsBufferSize = 256

	// Supervisor-level lifecycle event kinds emitted on the events bus
	// when Kill / Restart fire. Block-emitted events keep their own
	// kinds (e.g. process.started, process.exited).
	EventBlockKillRequested = "block.kill_requested"
	EventBlockRestarted     = "block.restarted"
)

// Event is a block-emitted event tagged with its origin block id. The
// supervisor multiplexes events from every spawned block into one bus so
// downstream consumers (TUI, future event log writer) read from a single
// channel.
type Event struct {
	BlockID string
	Kind    string
	TsMs    int64
	Payload json.RawMessage
}

// Events returns the read end of the event bus. The channel is buffered;
// slow consumers drop events rather than block the supervisor.
func (s *Supervisor) Events() <-chan Event { return s.events }

func (s *Supervisor) forwardEvents(blockID string, ep *transport.Endpoint) {
	ep.OnEvent(func(ev transport.Event) {
		s.emit(Event{BlockID: blockID, Kind: ev.Kind, TsMs: ev.TsMs, Payload: ev.Payload})
	})
}

func nowMs() int64 { return time.Now().UnixMilli() }
