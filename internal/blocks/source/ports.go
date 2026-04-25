package source

import (
	"net"
	"strconv"
)

func deriveUDPPorts(rtspAddr string) (rtp, rtcp string, err error) {
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
