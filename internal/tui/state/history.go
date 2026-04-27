package state

// SparkLength is the number of recent samples retained per link. It
// matches the sparkline column count rendered by ui.Spark; keeping the
// constant on the logic side means the renderer doesn't have to round
// or pad on every tick.
const SparkLength = 12

// History is a per-link rolling window of recent Mb/s samples used to
// drive sparklines. Buffer length matches the sparkline column count.
type History struct {
	samples map[string][]float64
}

// NewHistory returns an empty buffer.
func NewHistory() *History { return &History{samples: map[string][]float64{}} }

// Push appends v for id, dropping the oldest sample once the buffer is
// full.
func (h *History) Push(id string, v float64) {
	s := append(h.samples[id], v)
	if len(s) > SparkLength {
		s = s[len(s)-SparkLength:]
	}
	h.samples[id] = s
}

// For returns the rolling window for id (or nil when no samples yet).
func (h *History) For(id string) []float64 { return h.samples[id] }
