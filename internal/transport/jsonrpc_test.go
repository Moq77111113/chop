package transport

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

// TestEndpoint_RoundTrip: une requête envoyée a sa réponse délivrée à l'appelant.
// Si ce test casse, Call() ou la dispatch ne fonctionne plus.
func TestEndpoint_RoundTrip(t *testing.T) {
	aIn, bOut := io.Pipe()
	bIn, aOut := io.Pipe()

	a := NewEndpoint(aIn, aOut)
	b := NewEndpoint(bIn, bOut)

	b.Handle("echo", func(p json.RawMessage) (json.RawMessage, error) {
		return p, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _ = a.Serve(ctx) }()
	go func() { _ = b.Serve(ctx) }()

	res, err := a.Call(ctx, "echo", map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	var got map[string]string
	_ = json.Unmarshal(res, &got)
	if got["hello"] != "world" {
		t.Fatalf("expected echo, got %v", got)
	}
}

// TestEndpoint_UnknownMethod: appel à une méthode non enregistrée renvoie erreur structurée.
func TestEndpoint_UnknownMethod(t *testing.T) {
	aIn, bOut := io.Pipe()
	bIn, aOut := io.Pipe()
	a := NewEndpoint(aIn, aOut)
	b := NewEndpoint(bIn, bOut)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go func() { _ = a.Serve(ctx) }()
	go func() { _ = b.Serve(ctx) }()

	_, err := a.Call(ctx, "notdefined", nil)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
}
