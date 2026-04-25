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

// Registry tracks the running blocks of a Supervisor by id. List preserves
// insertion order — the order matches the scenario's declaration, which is
// the natural reading order for any consumer (TUI, API, dashboard).
type Registry struct {
	mu     sync.RWMutex
	blocks map[string]*Handle
	order  []string
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{blocks: map[string]*Handle{}}
}

// Add inserts h into the registry. A new ID is appended to the order; an
// existing ID is replaced in place without touching the order.
func (r *Registry) Add(h *Handle) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.blocks[h.ID]; !exists {
		r.order = append(r.order, h.ID)
	}
	r.blocks[h.ID] = h
}

// Get returns the Handle for id, or nil if no such block is registered.
func (r *Registry) Get(id string) *Handle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.blocks[id]
}

// List returns a snapshot of all currently registered Handles in insertion
// order.
func (r *Registry) List() []*Handle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Handle, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.blocks[id])
	}
	return out
}
