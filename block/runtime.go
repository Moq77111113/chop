package block

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moq77111113/chop/internal/transport"
)

const (
	exitBlockRunErr = 1
	exitBadArgv     = 2

	// argvBlockTypeOffset skips the program name and the "block" subcommand
	// in `chop block <type> -id ... -config ... -controls ...`.
	argvBlockTypeOffset = 2
)

var emptyAck = json.RawMessage(`{}`)

// Factory builds a Block from its parsed Config. Block implementations
// register their factory by name and pass it to RunBlock.
type Factory func(Config) Block

// RunBlock is the entry point for a block binary. It parses argv, opens stdio
// JSON-RPC, wires the Block to the supervisor, and blocks until SIGTERM/EOF.
func RunBlock(typeName string, factory Factory) {
	cfg := parseArgv(typeName)
	b := factory(cfg)

	ep := transport.NewEndpoint(os.Stdin, os.Stdout)
	ctx, cancel := signalContext()
	defer cancel()

	ctx = context.WithValue(ctx, emitterKey, ep)
	registerBlockHandlers(ep, b)

	go func() {
		if err := ep.Serve(ctx); err != nil && err != context.Canceled {
			fmt.Fprintf(os.Stderr, "rpc serve: %v\n", err)
			cancel()
		}
	}()

	if err := b.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "block run: %v\n", err)
		os.Exit(exitBlockRunErr)
	}
}

func parseArgv(typeName string) Config {
	fs := flag.NewFlagSet(typeName, flag.ExitOnError)
	id := fs.String("id", "", "block id (required)")
	static := fs.String("config", string(emptyAck), "static config (JSON)")
	live := fs.String("controls", string(emptyAck), "initial controls (JSON)")
	_ = fs.Parse(os.Args[argvBlockTypeOffset:])
	if *id == "" {
		fmt.Fprintln(os.Stderr, "missing -id")
		os.Exit(exitBadArgv)
	}
	return Config{
		ID:     *id,
		Type:   typeName,
		Static: json.RawMessage(*static),
		Live:   json.RawMessage(*live),
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
	return ctx, cancel
}
