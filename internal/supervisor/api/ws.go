package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	snapshotTickInterval = time.Second
	eventTypeSnapshot    = "snapshot"
)

// CheckOrigin allows any origin: M1 ships chop as a localhost dev tool,
// dashboard and API on the same port. Tighten this if the API ever serves
// outside dev.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type wsHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newWSHub() *wsHub {
	return &wsHub{clients: map[*websocket.Conn]struct{}{}}
}

func (h *wsHub) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.add(c)
	defer h.remove(c)
	for {
		if _, _, err := c.NextReader(); err != nil {
			return
		}
	}
}

func (h *wsHub) add(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *wsHub) remove(c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

func (h *wsHub) broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		_ = c.WriteMessage(websocket.TextMessage, msg)
	}
}

// Run drives the 1Hz snapshot fan-out: every tick, snapshot every registered
// block and broadcast the result to all connected WebSocket clients. Blocks
// until ctx is cancelled.
func (a *API) Run(ctx context.Context) error {
	t := time.NewTicker(snapshotTickInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			a.broadcastSnapshots(ctx)
		}
	}
}

func (a *API) broadcastSnapshots(ctx context.Context) {
	for _, h := range a.sup.Registry.List() {
		snap, err := h.Snapshot(ctx)
		if err != nil {
			continue
		}
		msg, err := json.Marshal(map[string]any{
			"type":     eventTypeSnapshot,
			"id":       h.ID,
			"snapshot": snap,
		})
		if err != nil {
			continue
		}
		a.hub.broadcast(msg)
	}
}
