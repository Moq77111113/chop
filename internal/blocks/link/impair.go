package link

import (
	"math/rand"
	"time"
)

const dropReasonDice = "dice"

// Decision is the per-packet outcome of impairment evaluation: drop the
// packet (with a reason) or forward it after Delay.
type Decision struct {
	Drop       bool
	DropReason string
	Delay      time.Duration
}

// Decide returns the impairment Decision for a single RTP packet given the
// current Controls. Pure: same inputs → same distribution. The caller owns
// rng and is responsible for serializing access to it.
func Decide(c *Controls, rng *rand.Rand) Decision {
	if lostToDice(c.Loss, rng) {
		return dropped(dropReasonDice)
	}
	return delayed(latencyWithJitter(c, rng))
}

func lostToDice(loss float64, rng *rand.Rand) bool {
	return loss > 0 && rng.Float64() < loss
}

func latencyWithJitter(c *Controls, rng *rand.Rand) time.Duration {
	base := time.Duration(c.LatencyMs) * time.Millisecond
	if c.JitterMs == 0 {
		return base
	}
	return base + time.Duration(rng.Int31n(int32(c.JitterMs+1)))*time.Millisecond
}

func dropped(reason string) Decision   { return Decision{Drop: true, DropReason: reason} }
func delayed(d time.Duration) Decision { return Decision{Delay: d} }
