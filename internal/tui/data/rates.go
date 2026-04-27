package data

import "time"

// Constants for the packets→Mb/s conversion used by ComputeRates.
const (
	avgRTPPacketBytes = 1500.0
	bitsPerByte       = 8
	bitsPerMb         = 1_000_000.0
	maxRateInterval   = 5 * time.Second
)

// ComputeRates derives a per-link Mb/s readout from two consecutive
// snapshot samples. dt outside (0, maxRateInterval] returns no rates so
// a stale previous sample doesn't produce a misleading number.
func ComputeRates(prev, cur map[string]LinkSnapshot, dt time.Duration) map[string]float64 {
	if dt <= 0 || dt > maxRateInterval {
		return nil
	}
	rates := make(map[string]float64, len(cur))
	for id, c := range cur {
		p, ok := prev[id]
		if !ok || c.PacketsOut < p.PacketsOut {
			continue
		}
		pps := float64(c.PacketsOut-p.PacketsOut) / dt.Seconds()
		rates[id] = pps * avgRTPPacketBytes * bitsPerByte / bitsPerMb
	}
	return rates
}
