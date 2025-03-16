package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Osmosis with CometBFT ===")
	fmt.Println("Standard Osmosis with CometBFT consensus")
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
	fmt.Println("Usage: osmosis-comet [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help, --help, -h     Show this help message")
	fmt.Println("  version, --version, -v  Show version information")
	fmt.Println()
}

func printVersion() {
	fmt.Println("Osmosis with CometBFT v0.1.0")
	fmt.Println("Standard Osmosis implementation with CometBFT consensus")
	fmt.Println()
}
