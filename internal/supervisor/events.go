package supervisor

import (
	"encoding/json"

	"github.com/moq77111113/chop/internal/transport"
)

const eventsBufferSize = 256

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
		select {
		case s.events <- Event{BlockID: blockID, Kind: ev.Kind, TsMs: ev.TsMs, Payload: ev.Payload}:
		default:
			// bus full — drop the event rather than stall the dispatch goroutine
		}
	})
}
