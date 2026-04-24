package block

import (
	"encoding/json"

	"github.com/moq77111113/chop/internal/transport"
)

func registerBlockHandlers(ep *transport.Endpoint, b Block) {
	ep.Handle(MethodInfo, infoHandler(b))
	ep.Handle(MethodSnapshot, snapshotHandler(b))
	ep.Handle(MethodApply, applyHandler(b))
	ep.Handle(MethodAction, actionHandler(b))
}

func infoHandler(b Block) transport.Handler {
	return func(_ json.RawMessage) (json.RawMessage, error) {
		return json.Marshal(b.Info())
	}
}

func snapshotHandler(b Block) transport.Handler {
	return func(_ json.RawMessage) (json.RawMessage, error) {
		return json.Marshal(b.Snapshot())
	}
}

func applyHandler(b Block) transport.Handler {
	return func(p json.RawMessage) (json.RawMessage, error) {
		if err := b.Apply(p); err != nil {
			return nil, err
		}
		return emptyAck, nil
	}
}

func actionHandler(b Block) transport.Handler {
	return func(p json.RawMessage) (json.RawMessage, error) {
		var a struct {
			Name string          `json:"name"`
			Args json.RawMessage `json:"args"`
		}
		if err := json.Unmarshal(p, &a); err != nil {
			return nil, err
		}
		if err := b.Action(a.Name, a.Args); err != nil {
			return nil, err
		}
		return emptyAck, nil
	}
}
