package process

import (
	"reflect"
	"testing"
)

func TestRing_TailReturnsAllWhenUnderCapacity(t *testing.T) {
	r := newRing(3)
	r.append("a")
	r.append("b")
	got := r.tail(20)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tail = %v, want %v", got, want)
	}
}

func TestRing_OverflowDropsOldestKeepsInsertionOrder(t *testing.T) {
	r := newRing(3)
	for _, line := range []string{"a", "b", "c", "d"} {
		r.append(line)
	}
	got := r.tail(20)
	want := []string{"b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tail = %v, want %v", got, want)
	}
}

func TestRing_TailLimitClipsToRequestedSize(t *testing.T) {
	r := newRing(5)
	for _, line := range []string{"a", "b", "c", "d"} {
		r.append(line)
	}
	got := r.tail(2)
	want := []string{"c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tail(2) = %v, want %v", got, want)
	}
}

func TestRing_EmptyTailReturnsEmpty(t *testing.T) {
	r := newRing(3)
	if got := r.tail(20); len(got) != 0 {
		t.Fatalf("tail of empty ring = %v, want empty", got)
	}
}
