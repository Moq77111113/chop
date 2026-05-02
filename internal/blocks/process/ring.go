package process

import "sync"

// ring is a fixed-capacity FIFO of stderr lines. append is O(1) amortised;
// tail is O(min(n, cap)) and copies so callers can't mutate the buffer.
type ring struct {
	mu    sync.Mutex
	buf   []string
	cap   int
	start int
	size  int
}

func newRing(capacity int) *ring {
	return &ring{buf: make([]string, capacity), cap: capacity}
}

func (r *ring) append(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.size < r.cap {
		r.buf[(r.start+r.size)%r.cap] = line
		r.size++
		return
	}
	r.buf[r.start] = line
	r.start = (r.start + 1) % r.cap
}

func (r *ring) tail(n int) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n > r.size {
		n = r.size
	}
	out := make([]string, n)
	off := r.size - n
	for i := 0; i < n; i++ {
		out[i] = r.buf[(r.start+off+i)%r.cap]
	}
	return out
}
