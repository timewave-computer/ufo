package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	fauxmosisintegration "github.com/timewave/ufo/src/fauxmosis-integration"
)

func main() {
	fmt.Println("=== Fauxmosis with UFO ===")
	fmt.Println("Starting UFO as a consensus engine for Fauxmosis")

	// Create and start the integration
	integration := fauxmosisintegration.NewFauxmosisUFOIntegration()
	err := integration.Start()
	if err != nil {
		fmt.Printf("Error starting integration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Fauxmosis with UFO integration is running")
	fmt.Println("RPC server is available at :26657")
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
	integration.Stop()
}
