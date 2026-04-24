package block

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moq77111113/chop/internal/transport"
)

type counterBlock struct {
	n atomic.Int64
}

func (c *counterBlock) Info() Info {
	return Info{ID: "test", Type: "counter"}
}
func (c *counterBlock) Snapshot() Snapshot {
	return Snapshot{Status: StatusRunning, TsMs: time.Now().UnixMilli()}
}
func (c *counterBlock) Apply(p json.RawMessage) error {
	var n int64
	if err := json.Unmarshal(p, &n); err != nil {
		return err
	}
	c.n.Store(n)
	return nil
}
func (c *counterBlock) Action(_ string, _ json.RawMessage) error { return nil }
func (c *counterBlock) Run(ctx context.Context) error            { <-ctx.Done(); return nil }

// TestRunBlock_AppliesControls : on simule un supervisor (Endpoint a) parlant
// via deux pipes à un fake block (Endpoint b) qui possède le counterBlock.
// L'envoi d'apply via Call() doit muter l'état du bloc. Si ce test casse,
// le routage stdio↔Block est cassé.
func TestRunBlock_AppliesControls(t *testing.T) {
	inR, inW := io.Pipe()   // supervisor → block
	outR, outW := io.Pipe() // block → supervisor

	counter := &counterBlock{}
	sup := transport.NewEndpoint(outR, inW) // supervisor side reads outR, writes inW

	var wg sync.WaitGroup
	wg.Go(func() {
		bep := transport.NewEndpoint(inR, outW)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		bep.Handle(MethodApply, func(p json.RawMessage) (json.RawMessage, error) {
			if err := counter.Apply(p); err != nil {
				return nil, err
			}
			return emptyAck, nil
		})
		_ = bep.Serve(ctx)
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _ = sup.Serve(ctx) }()

	_, err := sup.Call(ctx, "apply", int64(42))
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if got := counter.n.Load(); got != 42 {
		t.Fatalf("counter = %d, want 42", got)
	}

	cancel()
	inW.Close()
	outW.Close()
	wg.Wait()
}
