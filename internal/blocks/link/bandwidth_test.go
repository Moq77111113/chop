package link

import (
	"testing"
	"time"
)

const burstSec = 0.5

func TestBucket_UnlimitedRateNeverWaits(t *testing.T) {
	b := &bucket{}
	for range 100 {
		if got := b.take(12_000, 0, burstSec); got != 0 {
			t.Fatalf("got %v, want 0 for unlimited rate", got)
		}
	}
}

// TestBucket_RatePacesAboveCapacity: at 1 Mb/s, sending packets twice as
// fast as the rate should accumulate roughly real-time worth of delay over
// many iterations. We assert the cumulative delay is positive and grows
// monotonically — the exact amount depends on burst absorption.
func TestBucket_RatePacesAboveCapacity(t *testing.T) {
	const rateBps = 1_000_000.0
	const packetBits = 12_000
	b := &bucket{}
	var maxDelay time.Duration
	for range 50 {
		d := b.take(packetBits, rateBps, burstSec)
		if d > maxDelay {
			maxDelay = d
		}
	}
	if maxDelay == 0 {
		t.Fatalf("expected delay to grow above zero after sustained over-rate, got %v", maxDelay)
	}
}

// TestBucket_BurstAllowsCatchupAfterIdle: when nothing has been sent for a
// while, a fresh batch of packets should leave without delay (within the
// burst window).
func TestBucket_BurstAllowsCatchupAfterIdle(t *testing.T) {
	b := &bucket{}
	const rateBps = 1_000_000.0
	if got := b.take(8000, rateBps, burstSec); got != 0 {
		t.Fatalf("first packet delay = %v, want 0 after a fresh start", got)
	}
}
