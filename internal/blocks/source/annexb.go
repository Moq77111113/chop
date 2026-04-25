package source

import "bytes"

var (
	startCode4 = []byte{0x00, 0x00, 0x00, 0x01}
	startCode3 = []byte{0x00, 0x00, 0x01}
)

const (
	naluTypeMask   byte = 0x1F
	naluTypeNonIDR byte = 1
	naluTypeIDR    byte = 5
	naluTypeSPS    byte = 7
	naluTypePPS    byte = 8
)

func splitAnnexB(data []byte) [][]byte {
	var nalus [][]byte
	for {
		skip := matchStartCode(data)
		if skip == 0 {
			return nalus
		}
		data = data[skip:]
		end := findNextStartCode(data)
		if end < 0 {
			if len(data) > 0 {
				nalus = append(nalus, data)
			}
			return nalus
		}
		nalus = append(nalus, data[:end])
		data = data[end:]
	}
}

func matchStartCode(data []byte) int {
	switch {
	case bytes.HasPrefix(data, startCode4):
		return 4
	case bytes.HasPrefix(data, startCode3):
		return 3
	}
	return 0
}

func findNextStartCode(data []byte) int {
	i4 := bytes.Index(data, startCode4)
	i3 := bytes.Index(data, startCode3)
	if i4 < 0 {
		return i3
	}
	if i3 < 0 || i4 <= i3 {
		return i4
	}
	return i3
}

func groupAccessUnits(nalus [][]byte) [][][]byte {
	var aus [][][]byte
	var current [][]byte
	for _, n := range nalus {
		if len(n) == 0 {
			continue
		}
		current = append(current, n)
		if isVCL(n[0]) {
			aus = append(aus, current)
			current = nil
		}
	}
	if len(current) > 0 {
		aus = append(aus, current)
	}
	return aus
}

func isVCL(header byte) bool {
	t := header & naluTypeMask
	return t == naluTypeNonIDR || t == naluTypeIDR
}

// firstParameterSets returns the first SPS and PPS NALUs found in nalus, or
// nil for any not present. Required to populate the H264 SDP description so
// downstream decoders can configure themselves before the RTP stream arrives.
func firstParameterSets(nalus [][]byte) (sps, pps []byte) {
	for _, n := range nalus {
		if len(n) == 0 {
			continue
		}
		switch n[0] & naluTypeMask {
		case naluTypeSPS:
			if sps == nil {
				sps = n
			}
		case naluTypePPS:
			if pps == nil {
				pps = n
			}
		}
		if sps != nil && pps != nil {
			return
		}
	}
	return
}
