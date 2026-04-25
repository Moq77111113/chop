package source

import (
	"sync"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"

	"github.com/moq77111113/chop/internal/rtsp"
)

const (
	rtpPayloadType        = 96
	h264PacketizationMode = 1
)

type server struct {
	rtsp   *gortsplib.Server
	stream *gortsplib.ServerStream
	ready  sync.RWMutex
}

func startServer(listen string, sps, pps []byte) (*server, error) {
	rtpAddr, rtcpAddr, err := rtsp.DeriveUDPPorts(listen)
	if err != nil {
		return nil, err
	}

	s := &server{}
	s.ready.Lock()

	s.rtsp = &gortsplib.Server{
		Handler:        s,
		RTSPAddress:    listen,
		UDPRTPAddress:  rtpAddr,
		UDPRTCPAddress: rtcpAddr,
	}
	if err := s.rtsp.Start(); err != nil {
		return nil, err
	}

	s.stream = &gortsplib.ServerStream{Server: s.rtsp, Desc: newSessionDesc(sps, pps)}
	if err := s.stream.Initialize(); err != nil {
		s.rtsp.Close()
		return nil, err
	}

	s.ready.Unlock()
	return s, nil
}

func newSessionDesc(sps, pps []byte) *description.Session {
	return &description.Session{Medias: []*description.Media{{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{&format.H264{
			PayloadTyp:        rtpPayloadType,
			PacketizationMode: h264PacketizationMode,
			SPS:               sps,
			PPS:               pps,
		}},
	}}}
}

func (s *server) close() {
	s.stream.Close()
	s.rtsp.Close()
}

func (s *server) OnDescribe(*gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	s.ready.RLock()
	defer s.ready.RUnlock()
	return &base.Response{StatusCode: base.StatusOK}, s.stream, nil
}

func (s *server) OnSetup(*gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	s.ready.RLock()
	defer s.ready.RUnlock()
	return &base.Response{StatusCode: base.StatusOK}, s.stream, nil
}

func (s *server) OnPlay(*gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	return &base.Response{StatusCode: base.StatusOK}, nil
}
