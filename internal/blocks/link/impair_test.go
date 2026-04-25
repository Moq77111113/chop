package link

import (
	"math/rand"
	"sync"
	"testing"
)

const (
	noLossIterations = 1_000
	lossSampleSize   = 100_000
	lossTargetRate   = 0.10
	lossTolerance    = 0.005
	latencyMs        = 100
	jitterMs         = 20
	swapReaders      = 100
	swapWrites       = 100
)

func TestDecide_NoLossAlwaysForwards(t *testing.T) {
	c := &Controls{Loss: 0}
	rng := rand.New(rand.NewSource(1))
	for i := range noLossIterations {
		if Decide(c, rng).Drop {
			t.Fatalf("packet %d dropped with loss=0", i)
		}
	}
}

func TestDecide_LossMatchesProbability(t *testing.T) {
	c := &Controls{Loss: lossTargetRate}
	rng := rand.New(rand.NewSource(42))
	dropped := 0
	for range lossSampleSize {
		if Decide(c, rng).Drop {
			dropped++
		}
	}
	rate := float64(dropped) / float64(lossSampleSize)
	if rate < lossTargetRate-lossTolerance || rate > lossTargetRate+lossTolerance {
		t.Fatalf("loss rate = %.3f, want ~%.2f", rate, lossTargetRate)
	}
}

func TestDecide_DelayInJitterWindow(t *testing.T) {
	c := &Controls{LatencyMs: latencyMs, JitterMs: jitterMs}
	rng := rand.New(rand.NewSource(7))
	for range noLossIterations {
		d := Decide(c, rng).Delay.Milliseconds()
		if d < latencyMs || d > latencyMs+jitterMs {
			t.Fatalf("delay %d outside [%d,%d]", d, latencyMs, latencyMs+jitterMs)
		}
	}
}

func TestCtrlBox_AtomicSwap(t *testing.T) {
	box := newCtrlBox(Controls{Loss: 0})
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(swapReaders)
	for range swapReaders {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = box.Load().Loss
				}
			}
		}()
	}
	for i := range swapWrites {
		box.Store(&Controls{Loss: float64(i) / float64(swapWrites)})
	}
	close(done)
	wg.Wait()
}
