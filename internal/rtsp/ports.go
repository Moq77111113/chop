// Package rtsp provides low-level RTSP/RTP helpers shared between block types.
package rtsp

import (
	"net"
	"strconv"
)

// DeriveUDPPorts returns the UDP RTP/RTCP address pair derived from an RTSP
// listen address. gortsplib requires the RTP port to be even and RTCP to be
// RTP+1 (RFC 3550 §11). The first even port at or above rtspPort+1 is chosen.
func DeriveUDPPorts(rtspAddr string) (rtp, rtcp string, err error) {
	host, portStr, err := net.SplitHostPort(rtspAddr)
	if err != nil {
		return "", "", err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", "", err
	}
	rtpPort := nextEven(port + 1)
	return net.JoinHostPort(host, strconv.Itoa(rtpPort)),
		net.JoinHostPort(host, strconv.Itoa(rtpPort+1)),
		nil
}

func nextEven(n int) int {
	if n%2 == 0 {
		return n
	}
	return n + 1
}
