package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moq77111113/chop/internal/supervisor"
)

func TestListBlocks_EmptyRegistryReturnsEmptyArray(t *testing.T) {
	rec := serve(t, http.MethodGet, "/api/blocks", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("body = %v, want empty array", got)
	}
}

func TestGetBlock_UnknownIDReturns404(t *testing.T) {
	rec := serve(t, http.MethodGet, "/api/blocks/missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestApplyControls_UnknownIDReturns404(t *testing.T) {
	rec := serve(t, http.MethodPatch, "/api/blocks/missing/controls", strings.NewReader(`{}`))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func serve(t *testing.T, method, path string, body *strings.Reader) *httptest.ResponseRecorder {
	t.Helper()
	sup, err := supervisor.New()
	if err != nil {
		t.Fatalf("supervisor.New: %v", err)
	}
	mux := http.NewServeMux()
	New(sup).Mount(mux)

	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, body)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}
