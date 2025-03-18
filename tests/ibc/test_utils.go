package ibc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestTimeoutError is returned when a test operation times out
type TestTimeoutError struct {
	Operation string
	Duration  time.Duration
}

func (e TestTimeoutError) Error() string {
	return "operation timed out after " + e.Duration.String() + ": " + e.Operation
}

// WaitWithTimeout waits for a condition to be met with a timeout
func WaitWithTimeout(t *testing.T, timeout time.Duration, description string, condition func() bool) error {
	infoLog(t, "Waiting for: %s (timeout: %v)", description, timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(startTime)
			errorLog(t, "⏰ TIMEOUT: %s did not complete within %v (elapsed: %v)", description, timeout, elapsed)
			return TestTimeoutError{Operation: description, Duration: timeout}
		case <-ticker.C:
			if condition() {
				elapsed := time.Since(startTime)
				infoLog(t, "✓ COMPLETED: %s completed successfully in %v", description, elapsed)
				return nil
			}
			// Log progress if taking more than 1 second
			if time.Since(startTime) > time.Second {
				debugLog(t, "Still waiting for: %s (elapsed: %v)", description, time.Since(startTime))
			}
		}
	}
}

// RunWithTimeout runs an operation with a timeout
func RunWithTimeout(t *testing.T, timeout time.Duration, description string, operation func() error) error {
	infoLog(t, "Starting operation: %s (timeout: %v)", description, timeout)

	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		defer wg.Done()
		err = operation()
	}()

	// Create a channel that will be closed when the wait group is done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either the operation to complete or the timeout to occur
	select {
	case <-ctx.Done():
		elapsed := timeout
		errorLog(t, "⏰ TIMEOUT: %s did not complete within %v", description, elapsed)
		return TestTimeoutError{Operation: description, Duration: timeout}
	case <-done:
		if err != nil {
			errorLog(t, "❌ FAILED: %s failed with error: %v", description, err)
		} else {
			infoLog(t, "✓ COMPLETED: %s completed successfully", description)
		}
		return err
	}
}

// DefaultTestTimeout returns the default timeout for tests
// This can be overridden by setting the GO_TEST_TIMEOUT environment variable
func DefaultTestTimeout() time.Duration {
	return 2 * time.Minute
}

// DefaultSubOperationTimeout returns the default timeout for sub-operations within tests
func DefaultSubOperationTimeout() time.Duration {
	return 30 * time.Second
}

// WaitForBlocks waits for a specific number of blocks to be produced
func WaitForBlocks(t *testing.T, ctx context.Context, getHeight func() (int64, error), startHeight int64, blocks int64, timeout time.Duration) error {
	targetHeight := startHeight + blocks
	description := fmt.Sprintf("wait for %d blocks", blocks)

	infoLog(t, "Waiting for %d blocks to be produced (current height: %d, target: %d)", blocks, startHeight, targetHeight)
	startTime := time.Now()

	err := WaitWithTimeout(t, timeout, description, func() bool {
		currentHeight, err := getHeight()
		if err != nil {
			warnLog(t, "Failed to get current height: %v", err)
			return false
		}

		if currentHeight >= targetHeight {
			infoLog(t, "Target height %d reached (current: %d) after %v", targetHeight, currentHeight, time.Since(startTime))
			return true
		}

		blocksLeft := targetHeight - currentHeight
		debugLog(t, "Waiting for %d more blocks (current height: %d, target: %d)", blocksLeft, currentHeight, targetHeight)
		return false
	})

	return err
}

// ConfigureTestEnvironment configures the test environment for IBC testing
func ConfigureTestEnvironment(t *testing.T) {
	// Check if we're in the nix environment
	if isInNixEnvironment() {
		warnLog(t, "⚠️ Running in nix environment. IBC tests have limitations with nix-built binaries.")
		warnLog(t, "⚠️ See 'tests/ibc/nix_compatibility.md' for details and alternative approaches.")
	}

	// Ensure we're using the correct binary path from the nix environment
	infoLog(t, "Configuring test environment...")

	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		warnLog(t, "Failed to get current working directory: %v", err)
		pwd = os.Getenv("PWD")
	}

	// Build fresh binaries using nix command before each test
	if isInNixEnvironment() {
		infoLog(t, "Building fresh test binaries using nix...")
		projectRoot := findProjectRoot()
		buildCmd := exec.Command("bash", "-c", "source /dev/stdin <<< \"$(declare -f build_test_binaries)\" && build_test_binaries")
		buildCmd.Dir = projectRoot
		buildOutput, err := buildCmd.CombinedOutput()
		if err != nil {
			errorLog(t, "Failed to build test binaries: %v\nOutput: %s", err, string(buildOutput))
		} else {
			infoLog(t, "✅ Built fresh test binaries successfully")
			debugLog(t, "Build output: %s", string(buildOutput))
		}
	} else {
		infoLog(t, "Not in nix environment, skipping nix build")
	}

	// Set the result path to the absolute path from the project root
	projectRoot := findProjectRoot()
	resultPath := filepath.Join(projectRoot, "result")
	binPath := filepath.Join(projectRoot, "bin")

	debugLog(t, "Checking for binary paths: result=%s, bin=%s", resultPath, binPath)

	// Prioritize the nix build result path
	if _, err := os.Stat(resultPath); err == nil {
		infoLog(t, "Using binaries from nix result directory: %s", resultPath)
		newPath := fmt.Sprintf("%s:%s", resultPath, os.Getenv("PATH"))
		os.Setenv("PATH", newPath)
		// Also set binary locations in test config
		patchedBin := filepath.Join(resultPath, "osmosis-ufo-patched")
		bridgedBin := filepath.Join(resultPath, "osmosis-ufo-bridged")

		// Make binaries executable
		os.Chmod(patchedBin, 0755)
		os.Chmod(bridgedBin, 0755)

		os.Setenv("UFO_BIN", patchedBin)
		infoLog(t, "Set UFO_BIN to %s", patchedBin)
	} else if _, err := os.Stat(binPath); err == nil {
		infoLog(t, "Using binaries from project bin directory: %s", binPath)
		newPath := fmt.Sprintf("%s:%s", binPath, os.Getenv("PATH"))
		os.Setenv("PATH", newPath)
		// Also set binary locations in test config
		patchedBin := filepath.Join(binPath, "osmosis-ufo-patched")
		bridgedBin := filepath.Join(binPath, "osmosis-ufo-bridged")

		// Make binaries executable
		os.Chmod(patchedBin, 0755)
		os.Chmod(bridgedBin, 0755)

		os.Setenv("UFO_BIN", patchedBin)
		infoLog(t, "Set UFO_BIN to %s", patchedBin)
	} else {
		errorLog(t, "Neither result nor bin directory found at %s, tests will likely fail", projectRoot)
	}

	// Set default binary type if not already set
	if os.Getenv("UFO_BINARY_TYPE") == "" {
		os.Setenv("UFO_BINARY_TYPE", "patched")
		infoLog(t, "Setting default UFO_BINARY_TYPE to 'patched'")
	} else {
		infoLog(t, "Using UFO_BINARY_TYPE: %s", os.Getenv("UFO_BINARY_TYPE"))
	}

	// Attempt to make sure binaries are executable
	if binPath, ok := os.LookupEnv("UFO_BIN"); ok {
		if _, err := os.Stat(binPath); err == nil {
			if err := os.Chmod(binPath, 0755); err != nil {
				warnLog(t, "Failed to set executable permissions on binary: %v", err)
			} else {
				debugLog(t, "Set executable permissions on binary: %s", binPath)
			}
		} else {
			errorLog(t, "Binary not found at %s: %v", binPath, err)
		}
	}

	// Log key environment variables that could impact test execution
	debugLog(t, "Test environment configuration:")
	debugLog(t, "  - PWD: %s", pwd)
	debugLog(t, "  - PROJECT_ROOT: %s", projectRoot)
	debugLog(t, "  - PATH: %s", os.Getenv("PATH"))
	debugLog(t, "  - UFO_BINARY_TYPE: %s", os.Getenv("UFO_BINARY_TYPE"))
	debugLog(t, "  - UFO_BIN: %s", os.Getenv("UFO_BIN"))
	debugLog(t, "  - HERMES_CONFIG: %s", os.Getenv("HERMES_CONFIG"))
	debugLog(t, "  - IN_NIX_SHELL: %s", os.Getenv("IN_NIX_SHELL"))
}

// isInNixEnvironment checks if we're running inside a nix shell
func isInNixEnvironment() bool {
	// Check for common nix environment indicators
	if _, exists := os.LookupEnv("IN_NIX_SHELL"); exists {
		return true
	}

	// Check if common nix paths are in PATH
	path := os.Getenv("PATH")
	return strings.Contains(path, "/nix/store")
}

// findProjectRoot function removed as it's now in nix_utils.go

// BinaryType represents the type of binary to use for tests
type BinaryType string

const (
	BinaryTypeFauxmosisUfo BinaryType = "fauxmosis-ufo"
	BinaryTypeOsmosisUfo   BinaryType = "osmosis-ufo"
)

// GetBinaryPath returns the path to the binary of the given type.
// If binary doesn't exist, it will be built.
func GetBinaryPath(t *testing.T, binaryType BinaryType) string {
	// Try to find binary in the project
	projectRoot := findProjectRoot()

	// Check standard locations
	binDir := filepath.Join(projectRoot, "bin")
	buildDir := filepath.Join(projectRoot, "build")
	resultDir := filepath.Join(projectRoot, "result")

	// Look in bin/ directory
	if _, err := os.Stat(binDir); err == nil {
		binPath := filepath.Join(binDir, string(binaryType))
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
	}

	// Look in build/ directory
	if _, err := os.Stat(buildDir); err == nil {
		binPath := filepath.Join(buildDir, string(binaryType))
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
	}

	// Look in result/ directory (for nix)
	if _, err := os.Stat(resultDir); err == nil {
		binPath := filepath.Join(resultDir, string(binaryType))
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
	}

	// If we can't find it, return empty string
	return ""
}

// GetIBCTestRootDir returns the root directory for IBC tests
func GetIBCTestRootDir() string {
	projectRoot := findProjectRoot()
	return filepath.Join(projectRoot, "tests", "ibc")
}
