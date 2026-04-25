package supervisor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

	go relayStderr(id, stderr)
	return &child{cmd: cmd, rpc: transport.NewEndpoint(stdout, stdin)}, nil
}

func relayStderr(id string, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", id, sc.Text())
	}
}
