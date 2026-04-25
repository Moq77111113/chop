package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/moq77111113/chop/internal/scenario"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/supervisor/api"
	"github.com/moq77111113/chop/internal/tui"
)

const (
	defaultBindAddr = "127.0.0.1:6700"
	shutdownGrace   = 5 * time.Second
)

var (
	runBindAddr  string
	runOverrides []string
)

var runCmd = &cobra.Command{
	Use:   "run <scenario.yml>",
	Short: "Start the supervisor for a scenario",
	Args:  cobra.ExactArgs(1),
	RunE:  runScenario,
}

func init() {
	runCmd.Flags().StringVar(&runBindAddr, "bind", defaultBindAddr, "control API bind address")
	runCmd.Flags().StringArrayVar(&runOverrides, "override", nil,
		"per-block control overrides, e.g. link-1:loss=12%,latency=200ms (repeatable)")
}

func runScenario(_ *cobra.Command, args []string) error {
	sc, err := scenario.Load(args[0])
	if err != nil {
		return err
	}
	overrides, err := parseOverrides(runOverrides)
	if err != nil {
		return err
	}
	if err := applyOverrides(sc, overrides); err != nil {
		return err
	}
	sup, err := supervisor.New()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a := api.New(sup)
	mux := http.NewServeMux()
	a.Mount(mux)
	srv := &http.Server{Addr: runBindAddr, Handler: mux, ReadHeaderTimeout: shutdownGrace}

	go runComponent(cancel, "supervisor", func() error { return sup.Run(ctx, sc) })
	go runComponent(cancel, "api", func() error { return a.Run(ctx) })
	go runComponent(cancel, "http", func() error { return serveHTTP(srv, runBindAddr) })

	runErr := runForeground(ctx, sup, sc)
	cancel()
	if shutdownErr := gracefulShutdown(srv); shutdownErr != nil && runErr == nil {
		return shutdownErr
	}
	return runErr
}

// runForeground opens the TUI in interactive terminals; otherwise it stays
// headless and blocks until ctx is cancelled.
func runForeground(ctx context.Context, sup *supervisor.Supervisor, sc *scenario.Scenario) error {
	if !hasControllingTerminal() {
		<-ctx.Done()
		return nil
	}
	// Skip termenv's capability probe — slow on tmux/screen/some SSH paths.
	lipgloss.SetColorProfile(termenv.TrueColor)
	prog := tea.NewProgram(tui.New(sup, sc), tea.WithAltScreen(), tea.WithContext(ctx))
	go pumpEvents(ctx, sup, prog)
	_, err := prog.Run()
	return err
}

// pumpEvents forwards supervisor events into the TUI program as tea
// messages. It exits when ctx is cancelled.
func pumpEvents(ctx context.Context, sup *supervisor.Supervisor, prog *tea.Program) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-sup.Events():
			if !ok {
				return
			}
			prog.Send(tui.EventMsg{Event: ev})
		}
	}
}

// hasControllingTerminal mirrors what bubbletea actually needs at runtime
// (an openable /dev/tty for input). Stdin being a character device is a
// weaker guarantee — under nohup, &, or non-interactive shells, stdin can
// still look terminal-like while /dev/tty is unreachable.
func hasControllingTerminal() bool {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

func runComponent(cancel context.CancelFunc, name string, fn func() error) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		cancel()
	}
}

func serveHTTP(srv *http.Server, addr string) error {
	fmt.Printf("control API: http://%s\n", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func gracefulShutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	return srv.Shutdown(ctx)
}
