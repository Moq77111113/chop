package link

import (
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/moq77111113/chop/block"
)

type counters struct {
	pktsIn      atomic.Uint64
	pktsOut     atomic.Uint64
	pktsDropped atomic.Uint64
	upSinceMs   atomic.Int64
}

type stats struct {
	PacketsIn      uint64   `json:"packets_in"`
	PacketsOut     uint64   `json:"packets_out"`
	PacketsDropped uint64   `json:"packets_dropped"`
	UpSinceMs      int64    `json:"up_since_ms"`
	Controls       Controls `json:"controls"`
}

func snapshotOf(c *counters, ctrls *ctrlBox) block.Snapshot {
	payload, _ := json.Marshal(stats{
		PacketsIn:      c.pktsIn.Load(),
		PacketsOut:     c.pktsOut.Load(),
		PacketsDropped: c.pktsDropped.Load(),
		UpSinceMs:      c.upSinceMs.Load(),
		Controls:       *ctrls.Load(),
	})
	return block.Snapshot{
		Status: block.StatusRunning,
		Stats:  payload,
		TsMs:   time.Now().UnixMilli(),
	}
}
