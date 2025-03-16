package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("=== UFO: Universal Fast Orderer ===")
	fmt.Println("A lightweight alternative to CometBFT for Cosmos applications")
	fmt.Println()
	
	fmt.Println("UFO is ready to be used as a library.")
	fmt.Println("To integrate UFO with Osmosis, use the build-osmosis-ufo script.")
	fmt.Println()
	
	// Wait for interrupt signal if no arguments provided
	if len(os.Args) <= 1 {
		fmt.Println("Press Ctrl+C to exit.")
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("Shutting down...")
	}
} 