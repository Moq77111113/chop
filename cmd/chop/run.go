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

	"github.com/spf13/cobra"

	"github.com/moq77111113/chop/internal/dashboard"
	"github.com/moq77111113/chop/internal/scenario"
	"github.com/moq77111113/chop/internal/supervisor"
	"github.com/moq77111113/chop/internal/supervisor/api"
)

const (
	defaultBindAddr = "127.0.0.1:6700"
	shutdownGrace   = 5 * time.Second
	pathDashboard   = "/"
)

var runBindAddr string

var runCmd = &cobra.Command{
	Use:   "run <scenario.yml>",
	Short: "Start the supervisor and dashboard for a scenario",
	Args:  cobra.ExactArgs(1),
	RunE:  runScenario,
}

func init() {
	runCmd.Flags().StringVar(&runBindAddr, "bind", defaultBindAddr, "dashboard bind address")
}

func runScenario(_ *cobra.Command, args []string) error {
	sc, err := scenario.Load(args[0])
	if err != nil {
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
	mux.Handle(pathDashboard, dashboard.Handler())
	srv := &http.Server{Addr: runBindAddr, Handler: mux, ReadHeaderTimeout: shutdownGrace}

	go runComponent(cancel, "supervisor", func() error { return sup.Run(ctx, sc) })
	go runComponent(cancel, "api", func() error { return a.Run(ctx) })
	go runComponent(cancel, "http", func() error { return serveHTTP(srv, runBindAddr) })

	<-ctx.Done()
	return gracefulShutdown(srv)
}

func runComponent(cancel context.CancelFunc, name string, fn func() error) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		cancel()
	}
}

func serveHTTP(srv *http.Server, addr string) error {
	fmt.Printf("dashboard: http://%s\n", addr)
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
