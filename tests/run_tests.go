// Package tests is the main entry point for the UFO test suite
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestPackages defines the test packages to run
var TestPackages = []string{
	"./core/rest",
	"./core/grpc",
	"./core/websocket",
	"./core/mempool",
	// Add more test packages as they are implemented
	// "./consensus",
	// "./ibc",
	// "./integration",
	// "./stress",
}

func main() {
	// Find the test directory
	testDir, err := findTestDir()
	if err != nil {
		fmt.Printf("Error finding test directory: %v\n", err)
		os.Exit(1)
	}

	// Run the tests
	success := runTests(testDir)
	if !success {
		os.Exit(1)
	}
}

// findTestDir attempts to find the test directory
func findTestDir() (string, error) {
	// Try the current directory first
	if _, err := os.Stat("./tests"); err == nil {
		return "./tests", nil
	}

	// Try to find from the current executable location
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Check if we're in the test directory
	dir := filepath.Dir(execPath)
	if filepath.Base(dir) == "tests" {
		return dir, nil
	}

	// Check parent directory
	parent := filepath.Dir(dir)
	testDir := filepath.Join(parent, "tests")
	if _, err := os.Stat(testDir); err == nil {
		return testDir, nil
	}

	// As a last resort, use the current directory and hope for the best
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	return cwd, nil
}

// runTests runs the tests in the specified packages
func runTests(testDir string) bool {
	allPassed := true

	for _, pkg := range TestPackages {
		fmt.Printf("Running tests in %s\n", pkg)

		// Build the command
		cmd := exec.Command("go", "test", "-v", pkg)
		cmd.Dir = testDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		if err := cmd.Run(); err != nil {
			fmt.Printf("Tests in %s failed: %v\n", pkg, err)
			allPassed = false
		} else {
			fmt.Printf("Tests in %s passed\n", pkg)
		}

		// Add a separator between packages
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print summary
	if allPassed {
		fmt.Println("All tests passed!")
	} else {
		fmt.Println("Some tests failed")
	}

	return allPassed
}
