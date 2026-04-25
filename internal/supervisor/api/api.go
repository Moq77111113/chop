// Package api exposes the supervisor's HTTP and WebSocket surface to the
// dashboard.
package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/moq77111113/chop/internal/supervisor"
)

const (
	pathListBlocks    = "GET /api/blocks"
	pathGetBlock      = "GET /api/blocks/{id}"
	pathApplyControls = "PATCH /api/blocks/{id}/controls"
	pathEvents        = "/api/events"

	contentTypeJSON = "application/json"
)

// API is the HTTP/WebSocket façade in front of a Supervisor. The same
// instance serves the REST routes and the live event WebSocket.
type API struct {
	sup *supervisor.Supervisor
	hub *wsHub
}

// New constructs an API bound to the given Supervisor.
func New(sup *supervisor.Supervisor) *API {
	return &API{sup: sup, hub: newWSHub()}
}

// Mount registers all HTTP routes and the WebSocket endpoint on mux.
// It does not start any background work — see Run for the snapshot fan-out.
func (a *API) Mount(mux *http.ServeMux) {
	mux.HandleFunc(pathListBlocks, a.listBlocks)
	mux.HandleFunc(pathGetBlock, a.getBlock)
	mux.HandleFunc(pathApplyControls, a.applyControls)
	mux.HandleFunc(pathEvents, a.hub.handleUpgrade)
}

func (a *API) listBlocks(w http.ResponseWriter, _ *http.Request) {
	type entry struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	handles := a.sup.Registry.List()
	out := make([]entry, 0, len(handles))
	for _, h := range handles {
		out = append(out, entry{ID: h.ID, Type: h.Type})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) getBlock(w http.ResponseWriter, r *http.Request) {
	h := a.sup.Registry.Get(r.PathValue("id"))
	if h == nil {
		http.NotFound(w, r)
		return
	}
	snap, err := h.Snapshot(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       h.ID,
		"type":     h.Type,
		"snapshot": snap,
	})
}

func (a *API) applyControls(w http.ResponseWriter, r *http.Request) {
	h := a.sup.Registry.Get(r.PathValue("id"))
	if h == nil {
		http.NotFound(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Apply(r.Context(), body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
