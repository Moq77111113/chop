package source

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/Eyevinn/mp4ff/avc"
	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
)

const rtpClockRate = 90000

type h264Stream struct {
	sps []byte
	pps []byte
	aus [][][]byte
}

var (
	errNoVideoTrack = errors.New("source: MP4 has no video track")
	errNoAvcConfig  = errors.New("source: MP4 video track is not H264")
	errNoSPSPPS     = errors.New("source: MP4 AVC config missing SPS or PPS")
)

// loadMP4 reads a progressive MP4 file and returns its first H264 video track
// as access units, plus the SPS/PPS extracted from the avcC box. MP4 sample
// boundaries map directly to access units, sidestepping the multi-slice
// ambiguity of raw Annex-B streams.
func loadMP4(file string) (h264Stream, error) {
	f, err := os.Open(file)
	if err != nil {
		return h264Stream{}, err
	}
	defer f.Close()

	parsed, err := mp4.DecodeFile(f)
	if err != nil {
		return h264Stream{}, fmt.Errorf("source: parse MP4: %w", err)
	}
	trak, ok := firstVideoTrack(parsed.Moov)
	if !ok {
		return h264Stream{}, errNoVideoTrack
	}
	sps, pps, err := extractAvcParameterSets(trak)
	if err != nil {
		return h264Stream{}, err
	}
	aus, err := readAccessUnits(parsed.Mdat, trak)
	if err != nil {
		return h264Stream{}, err
	}
	return h264Stream{sps: sps, pps: pps, aus: aus}, nil
}

func firstVideoTrack(moov *mp4.MoovBox) (*mp4.TrakBox, bool) {
	for _, t := range moov.Traks {
		if t.Mdia != nil && t.Mdia.Hdlr != nil && t.Mdia.Hdlr.HandlerType == "vide" {
			return t, true
		}
	}
	return nil, false
}

func extractAvcParameterSets(trak *mp4.TrakBox) (sps, pps []byte, err error) {
	avcX := trak.Mdia.Minf.Stbl.Stsd.AvcX
	if avcX == nil || avcX.AvcC == nil {
		return nil, nil, errNoAvcConfig
	}
	if len(avcX.AvcC.SPSnalus) == 0 || len(avcX.AvcC.PPSnalus) == 0 {
		return nil, nil, errNoSPSPPS
	}
	return avcX.AvcC.SPSnalus[0], avcX.AvcC.PPSnalus[0], nil
}

func readAccessUnits(mdat *mp4.MdatBox, trak *mp4.TrakBox) ([][][]byte, error) {
	stbl := trak.Mdia.Minf.Stbl
	count := int(stbl.Stsz.SampleNumber)
	mdatStart := mdat.PayloadAbsoluteOffset()
	aus := make([][][]byte, 0, count)

	for sampleNr := 1; sampleNr <= count; sampleNr++ {
		offset, err := absoluteSampleOffset(stbl, sampleNr)
		if err != nil {
			return nil, err
		}
		size := stbl.Stsz.GetSampleSize(sampleNr)
		rel := uint64(offset) - mdatStart
		nalus, err := avc.GetNalusFromSample(mdat.Data[rel : rel+uint64(size)])
		if err != nil {
			return nil, fmt.Errorf("source: decode AVCC sample %d: %w", sampleNr, err)
		}
		aus = append(aus, nalus)
	}
	return aus, nil
}

func absoluteSampleOffset(stbl *mp4.StblBox, sampleNr int) (int64, error) {
	chunkNr, firstSampleInChunk, err := stbl.Stsc.ChunkNrFromSampleNr(sampleNr)
	if err != nil {
		return 0, err
	}
	offset, err := chunkOffset(stbl, chunkNr)
	if err != nil {
		return 0, err
	}
	for s := firstSampleInChunk; s < sampleNr; s++ {
		offset += int64(stbl.Stsz.GetSampleSize(s))
	}
	return offset, nil
}

func chunkOffset(stbl *mp4.StblBox, chunkNr int) (int64, error) {
	switch {
	case stbl.Stco != nil:
		return int64(stbl.Stco.ChunkOffset[chunkNr-1]), nil
	case stbl.Co64 != nil:
		return int64(stbl.Co64.ChunkOffset[chunkNr-1]), nil
	}
	return 0, errors.New("source: MP4 missing stco/co64 chunk offsets")
}

func replay(ctx context.Context, stream *gortsplib.ServerStream, aus [][][]byte, fps int, served *atomic.Int64) error {
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
