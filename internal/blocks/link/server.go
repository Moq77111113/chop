package link

import (
	"sync"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"

	"github.com/moq77111113/chop/internal/rtsp"
)

type downstream struct {
	rtsp   *gortsplib.Server
	stream *gortsplib.ServerStream
	ready  sync.RWMutex
}

func startDownstream(listen string, desc *description.Session) (*downstream, error) {
	rtpAddr, rtcpAddr, err := rtsp.DeriveUDPPorts(listen)
	if err != nil {
		return nil, err
	}

	d := &downstream{}
	d.ready.Lock()

	d.rtsp = &gortsplib.Server{
		Handler:        d,
		RTSPAddress:    listen,
		UDPRTPAddress:  rtpAddr,
		UDPRTCPAddress: rtcpAddr,
	}
	if err := d.rtsp.Start(); err != nil {
		return nil, err
	}

	d.stream = &gortsplib.ServerStream{Server: d.rtsp, Desc: desc}
	if err := d.stream.Initialize(); err != nil {
		d.rtsp.Close()
		return nil, err
	}

	d.ready.Unlock()
	return d, nil
}

func (d *downstream) close() {
	d.stream.Close()
	d.rtsp.Close()
}

func (d *downstream) OnDescribe(*gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	d.ready.RLock()
	defer d.ready.RUnlock()
	return &base.Response{StatusCode: base.StatusOK}, d.stream, nil
}

func (d *downstream) OnSetup(*gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	d.ready.RLock()
	defer d.ready.RUnlock()
	return &base.Response{StatusCode: base.StatusOK}, d.stream, nil
}

func (d *downstream) OnPlay(*gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	return &base.Response{StatusCode: base.StatusOK}, nil
}
