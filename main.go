package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	t.provider = p

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle control interface operations
	go func() {
		providerCh <- p.Start()
	}()

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
