package source_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"

	"github.com/moq77111113/chop/block"
	"github.com/moq77111113/chop/internal/blocks/source"
)

const (
	testFixture    = "../../../testdata/pattern.mp4"
	testListenAddr = "127.0.0.1:5101"
	testStreamPath = "/stream"
	serverWarmup   = 300 * time.Millisecond
	testTimeout    = 5 * time.Second
	testFPS        = 25
)

func TestSource_ClientCanDescribeVideo(t *testing.T) {
	if _, err := os.Stat(testFixture); err != nil {
		t.Skipf("fixture %s not found — run `make fixture` to generate", testFixture)
	}
	b := source.New(block.Config{
		ID:   "src-test",
		Type: "source",
		Static: marshal(t, source.Config{
			File:   testFixture,
			Listen: testListenAddr,
			FPS:    testFPS,
		}),
	})

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	runErr := make(chan error, 1)
	go func() { runErr <- b.Run(ctx) }()

	time.Sleep(serverWarmup)

	u, err := base.ParseURL("rtsp://" + testListenAddr + testStreamPath)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	c := &gortsplib.Client{Scheme: u.Scheme, Host: u.Host}
	if err := c.Start(); err != nil {
		t.Fatalf("client start: %v", err)
	}
	defer c.Close()

	desc, _, err := c.Describe(u)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if len(desc.Medias) == 0 || desc.Medias[0].Type != description.MediaTypeVideo {
		t.Fatalf("expected a video media in SDP, got %+v", desc.Medias)
	}
}

func marshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}
