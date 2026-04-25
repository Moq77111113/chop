package supervisor

import "testing"

func TestRegistry_GetReturnsNilForUnknownID(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("missing"); got != nil {
		t.Fatalf("Get(missing) = %v, want nil", got)
	}
}

func TestRegistry_AddThenGetReturnsSameHandle(t *testing.T) {
	r := NewRegistry()
	h := &Handle{ID: "a", Type: "source"}
	r.Add(h)
	if got := r.Get("a"); got != h {
		t.Fatalf("Get(a) = %v, want %v", got, h)
	}
}

func TestRegistry_AddReplacesEntryWithSameID(t *testing.T) {
	r := NewRegistry()
	r.Add(&Handle{ID: "a", Type: "v1"})
	r.Add(&Handle{ID: "a", Type: "v2"})
	if got := r.Get("a"); got.Type != "v2" {
		t.Fatalf("Type = %q after replace, want v2", got.Type)
	}
}

func TestRegistry_ListReturnsAllRegisteredHandles(t *testing.T) {
	r := NewRegistry()
	r.Add(&Handle{ID: "a"})
	r.Add(&Handle{ID: "b"})
	if got := r.List(); len(got) != 2 {
		t.Fatalf("List len = %d, want 2", len(got))
	}
}
