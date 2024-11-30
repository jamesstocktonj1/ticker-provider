package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.wasmcloud.dev/provider"
)

//go:generate wit-bindgen-wrpc go --out-dir bindings --world imports --package github.com/jamesstocktonj1/ticker-provider/bindings wit
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Create new Ticker instance
	t, err := CreateTicker()
	if err != nil {
		return err
	}

	// Create new wasmcloud provider instance
	p, err := provider.New(
		provider.TargetLinkPut(t.handlePutTargetLink),
		provider.TargetLinkDel(t.handleDelTargetLink),
		provider.HealthCheck(t.handleHealthCheck),
		provider.Shutdown(t.handleShutdown),
	)
	if err != nil {
		return err
	}

	// Forward logs to Otel
	if p.HostData().OtelConfig.EnableObservability || p.HostData().OtelConfig.EnableLogs {
		p.Logger = slog.New(slogmulti.Fanout(
			p.Logger.Handler(),
			otelslog.NewLogger(OtelName).Handler(),
		))
	}
	t.provider = p

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle control interface operations
	go func() {
		err := p.Start()
		providerCh <- err
	}()

	// Start ticker scheduler
	t.Start()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	// Run provider until either a shutdown is requested or a SIGINT is received
	select {
	case err = <-providerCh:
		t.Shutdown()
		return err
	case <-signalCh:
		t.Shutdown()
		p.Shutdown()
	}

	return nil
}
