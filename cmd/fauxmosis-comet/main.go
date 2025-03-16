package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Fauxmosis with CometBFT ===")
	fmt.Println("A mock Cosmos SDK application with CometBFT for testing and benchmarking")
	fmt.Println()

	if len(os.Args) <= 1 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		printHelp()
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: fauxmosis-comet [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help, --help, -h     Show this help message")
	fmt.Println("  version, --version, -v  Show version information")
	fmt.Println()
}

func printVersion() {
	fmt.Println("Fauxmosis with CometBFT v0.1.0")
	fmt.Println("Mock Cosmos SDK application with CometBFT for testing and benchmarking")
	fmt.Println()
}
