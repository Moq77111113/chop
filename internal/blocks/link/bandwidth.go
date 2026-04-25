package link

import (
	"sync"
	"time"
)

// bandwidthBurstSeconds is how long an idle stream can over-shoot the
// configured rate before pacing re-engages. Half a second matches what most
// real-world links absorb without dropping (TCP-friendly).
const bandwidthBurstSeconds = 0.5

// bucket paces packet emission to a target bit-rate using a leaky-bucket
// scheduler. It tracks the earliest instant the next packet may leave;
// callers translate the returned wait duration into a deferred forward.
type bucket struct {
	mu     sync.Mutex
	nextOK time.Time
}

// take returns the wait duration before `bits` bits may be sent at rateBps
// bits/second. burstSec controls how far in the past nextOK can fall when
// the link is idle (== max burst window in seconds). rateBps == 0 means
// "unlimited" — no wait.
func (b *bucket) take(bits int, rateBps float64, burstSec float64) time.Duration {
	if rateBps <= 0 {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	burstFloor := now.Add(-time.Duration(burstSec * float64(time.Second)))
	if b.nextOK.Before(burstFloor) {
		b.nextOK = burstFloor
	}
	departure := b.nextOK
	if departure.Before(now) {
		departure = now
	}
	cost := time.Duration(float64(bits) / rateBps * float64(time.Second))
	b.nextOK = b.nextOK.Add(cost)
	return departure.Sub(now)
}
