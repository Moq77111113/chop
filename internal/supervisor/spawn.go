package supervisor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/moq77111113/chop/internal/transport"
)

const (
	blockSubcommand  = "block"
	flagBlockID      = "-id"
	flagStaticConfig = "-config"
	flagInitControls = "-controls"
	shutdownGrace    = 5 * time.Second

	stderrScanInit = 64 * 1024
	stderrScanMax  = 1024 * 1024
)

type child struct {
	cmd *exec.Cmd
	rpc *transport.Endpoint
}

func spawnChild(ctx context.Context, exe, blockType, id string, staticCfg, controls json.RawMessage) (*child, error) {
	cmd := exec.CommandContext(ctx, exe,
		blockSubcommand, blockType,
		flagBlockID, id,
		flagStaticConfig, string(staticCfg),
		flagInitControls, string(controls),
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = shutdownGrace

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start child %s: %w", id, err)
	}

	// Drain to discard. The pipe must be read or the child blocks once it
	// fills; routing to os.Stderr corrupts the bubbletea render. Headless
	// visibility of child stderr is deferred to a future log sink.
	go drainAndDiscard(stderr)
	return &child{cmd: cmd, rpc: transport.NewEndpoint(stdout, stdin)}, nil
}

func (c *child) wait() error { return c.cmd.Wait() }

func drainAndDiscard(r io.Reader) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, stderrScanInit), stderrScanMax)
	for sc.Scan() {
		_ = sc.Text()
	}
	_, _ = io.Copy(io.Discard, r)
}
