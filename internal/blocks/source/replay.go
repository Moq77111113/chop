package source

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
)

const rtpClockRate = 90000

func replay(ctx context.Context, stream *gortsplib.ServerStream, file string, fps int, served *atomic.Int64) error {
	aus, err := loadAccessUnits(file)
	if err != nil {
		return err
	}
	if len(aus) == 0 {
		return nil
	}

	media := stream.Desc.Medias[0]
	enc, err := media.Formats[0].(*format.H264).CreateEncoder()
	if err != nil {
		return err
	}

	return runLoop(ctx, stream, media, enc, aus, fps, served)
}

func runLoop(
	ctx context.Context,
	stream *gortsplib.ServerStream,
	media *description.Media,
	enc *rtph264.Encoder,
	aus [][][]byte,
	fps int,
	served *atomic.Int64,
) error {
	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	tsStep := uint32(rtpClockRate / fps)
	var ts uint32
	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			pkts, err := enc.Encode(aus[i%len(aus)])
			i++
			if err != nil {
				continue
			}
			for _, p := range pkts {
				p.Timestamp = ts
				_ = stream.WritePacketRTP(media, p)
			}
			served.Add(int64(len(pkts)))
			ts += tsStep
		}
	}
}

func loadAccessUnits(file string) ([][][]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return groupAccessUnits(splitAnnexB(data)), nil
}
