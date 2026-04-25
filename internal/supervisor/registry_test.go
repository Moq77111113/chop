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

func TestRegistry_ListPreservesInsertionOrder(t *testing.T) {
	r := NewRegistry()
	ids := []string{"cam-1", "link-1", "process-x", "link-2"}
	for _, id := range ids {
		r.Add(&Handle{ID: id})
	}
	r.Add(&Handle{ID: "link-1", Type: "replaced"}) // replacement keeps slot

	got := r.List()
	if len(got) != len(ids) {
		t.Fatalf("len = %d, want %d", len(got), len(ids))
	}
	for i, want := range ids {
		if got[i].ID != want {
			t.Fatalf("position %d: got %q, want %q", i, got[i].ID, want)
		}
	}
}
