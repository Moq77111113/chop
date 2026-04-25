package supervisor

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/moq77111113/chop/block"
)

// Handle is the supervisor-side reference to a running block. It exposes
// the public RPC surface (Snapshot, Apply) without leaking the underlying
// child process to API consumers.
type Handle struct {
	ID    string
	Type  string
	child *child
}

// Snapshot calls the block's snapshot method over JSON-RPC and decodes the
// reply.
func (h *Handle) Snapshot(ctx context.Context) (block.Snapshot, error) {
	raw, err := h.child.rpc.Call(ctx, block.MethodSnapshot, nil)
	if err != nil {
		return block.Snapshot{}, err
	}
	var s block.Snapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		return block.Snapshot{}, err
	}
	return s, nil
}

// Apply forwards a controls JSON payload to the block's apply method.
func (h *Handle) Apply(ctx context.Context, controls json.RawMessage) error {
	_, err := h.child.rpc.Call(ctx, block.MethodApply, controls)
	return err
}

// Registry tracks the running blocks of a Supervisor by id. Safe for
// concurrent use.
type Registry struct {
	mu     sync.RWMutex
	blocks map[string]*Handle
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{blocks: map[string]*Handle{}}
}

// Add inserts h into the registry, replacing any existing entry with the
// same ID.
func (r *Registry) Add(h *Handle) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blocks[h.ID] = h
}

// Get returns the Handle for id, or nil if no such block is registered.
func (r *Registry) Get(id string) *Handle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.blocks[id]
}

// List returns a snapshot of all currently registered Handles in
// unspecified order.
func (r *Registry) List() []*Handle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Handle, 0, len(r.blocks))
	for _, h := range r.blocks {
		out = append(out, h)
	}
	return out
}
