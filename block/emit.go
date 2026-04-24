package block

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type contextKey string

const emitterKey contextKey = "chop.emitter"

type emitter interface {
	Emit(kind string, payload any)
}

// Emit publishes a structured event to the supervisor.
// Safe to call from any goroutine inside Block.Run.
func Emit(ctx context.Context, kind string, payload any) {
	e, ok := ctx.Value(emitterKey).(emitter)
	if !ok {
		// No emitter (test or standalone) — fall back to stderr.
		b, _ := json.Marshal(payload)
		fmt.Fprintf(os.Stderr, "[event] %s %s\n", kind, b)
		return
	}
	e.Emit(kind, payload)
}
