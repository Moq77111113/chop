// Package transport provides ndjson-framed JSON-RPC over io.Reader/io.Writer.
// Used internally by the chop framework to talk to block child processes
// over their stdin/stdout. Hidden from block authors by block.RunBlock.
package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	msgTypeReq   = "req"
	msgTypeResp  = "resp"
	msgTypeEvent = "event"

	scanInitialBuf = 64 * 1024
	scanMaxBuf     = 4 * 1024 * 1024

	errPrefixUnknownMethod = "unknown method: "
)

// Message is one ndjson frame on the wire. Type ("req", "resp", "event")
// selects which fields are populated.
type Message struct {
	Type   string          `json:"type"`
	ID     uint64          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`

	// Set on event messages only.
	Kind    string          `json:"kind,omitempty"`
	TsMs    int64           `json:"ts,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Handler answers a single JSON-RPC request. Returning an error propagates
// it as the response error string.
type Handler func(params json.RawMessage) (json.RawMessage, error)

// Event is the wire-decoded form of an event-type Message, surfaced to a
// registered handler so consumers don't peer at Message internals.
type Event struct {
	Kind    string
	TsMs    int64
	Payload json.RawMessage
}

// EventHandler receives events as they arrive on the endpoint. The handler
// runs on the dispatch goroutine — keep it fast or hand off to a channel.
type EventHandler func(Event)

// Endpoint is a bidirectional JSON-RPC peer over an ndjson byte stream.
// It serves registered handlers, makes outbound calls, and emits events.
type Endpoint struct {
	in       *bufio.Scanner
	out      io.Writer
	outMu    sync.Mutex
	handlers map[string]Handler
	onEvent  EventHandler
	nextID   atomic.Uint64
	pending  sync.Map // id -> chan Message
}

// OnEvent registers a callback invoked for each event-type message read
// from the peer. Only the last registered handler is kept.
func (e *Endpoint) OnEvent(h EventHandler) { e.onEvent = h }

// NewEndpoint wraps an io.Reader/io.Writer pair as a JSON-RPC endpoint.
// The reader is buffered with a 4MiB max line size; lines beyond that fail.
func NewEndpoint(in io.Reader, out io.Writer) *Endpoint {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, scanInitialBuf), scanMaxBuf)
	return &Endpoint{in: sc, out: out, handlers: map[string]Handler{}}
}

// Handle registers a Handler for the given method name. Not safe for concurrent
// use with Serve — register all handlers before calling Serve.
func (e *Endpoint) Handle(method string, h Handler) {
	e.handlers[method] = h
}

// Serve reads messages until EOF or ctx.Done. Blocking.
func (e *Endpoint) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !e.in.Scan() {
			return e.in.Err()
		}
		var m Message
		if err := json.Unmarshal(e.in.Bytes(), &m); err != nil {
			return fmt.Errorf("decode message: %w", err)
		}
		e.dispatch(m)
	}
}

func (e *Endpoint) dispatch(m Message) {
	switch m.Type {
	case msgTypeReq:
		go e.handleRequest(m)
	case msgTypeResp:
		e.deliverResponse(m)
	case msgTypeEvent:
		if e.onEvent != nil {
			e.onEvent(Event{Kind: m.Kind, TsMs: m.TsMs, Payload: m.Payload})
		}
	}
}

func (e *Endpoint) handleRequest(m Message) {
	h, ok := e.handlers[m.Method]
	if !ok {
		e.write(Message{Type: msgTypeResp, ID: m.ID, Error: errPrefixUnknownMethod + m.Method})
		return
	}
	res, err := h(m.Params)
	resp := Message{Type: msgTypeResp, ID: m.ID}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Result = res
	}
	e.write(resp)
}

func (e *Endpoint) deliverResponse(m Message) {
	if ch, ok := e.pending.LoadAndDelete(m.ID); ok {
		ch.(chan Message) <- m
	}
}

// Call sends a request and waits for the response.
func (e *Endpoint) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	id := e.nextID.Add(1)
	ch := make(chan Message, 1)
	e.pending.Store(id, ch)
	e.write(Message{Type: msgTypeReq, ID: id, Method: method, Params: raw})
	select {
	case <-ctx.Done():
		e.pending.Delete(id)
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != "" {
			return nil, fmt.Errorf("%s", resp.Error)
		}
		return resp.Result, nil
	}
}

// Emit sends an event (fire-and-forget, no response expected).
func (e *Endpoint) Emit(kind string, payload any) {
	raw, _ := json.Marshal(payload)
	e.write(Message{Type: msgTypeEvent, Kind: kind, TsMs: nowMs(), Payload: raw})
}

func (e *Endpoint) write(m Message) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	e.outMu.Lock()
	defer e.outMu.Unlock()
	fmt.Fprintln(e.out, string(b))
}

func nowMs() int64 { return time.Now().UnixMilli() }
