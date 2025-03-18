package ibc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// NixChainConfig represents configuration for a chain in nix environment
type NixChainConfig struct {
	Name                        string
	HomeDir                     string
	BinaryPath                  string
	RPCPort                     string
	P2PPort                     string
	GRPCPort                    string
	RESTPort                    string
	ValidatorCount              int
	EpochLength                 int
	ValidatorWeightChangeBlocks int
	Byzantine                   bool
}

// NixChain represents a running chain in the nix environment
type NixChain struct {
	Config   NixChainConfig
	Cmd      *exec.Cmd
	Finished bool
	T        *testing.T
	Process  *os.Process
}

// logInfo logs information with a timestamp
func logInfo(t *testing.T, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	t.Logf("[%s] %s", time.Now().Format("15:04:05.000"), message)
}

// isInNixShell checks if the test is running in a nix shell
func isInNixShell() bool {
	// Check for IN_NIX_SHELL environment variable
	if os.Getenv("IN_NIX_SHELL") != "" {
		return true
	}

	// Check if PATH contains /nix/store
	path := os.Getenv("PATH")
	if strings.Contains(path, "/nix/store") {
		return true
	}

	return false
}

// configureEnvironment sets up the environment for running tests in nix
func configureEnvironment(t *testing.T) {
	// Log nix environment information
	if isInNixShell() {
		t.Log("Running in nix environment")

		// Set environment variables needed for tests
		os.Setenv("UFO_BINARY_TYPE", "patched")

		// Check if we have the necessary environment variables
		if os.Getenv("UFO_BIN") == "" {
			// Try to find the binary
			binaryPath := GetNixBinaryPath(t)
			os.Setenv("UFO_BIN", binaryPath)
		}
	} else {
		t.Log("Not running in nix environment")
	}
}

// defaultTestTimeout returns the default timeout for tests
func defaultTestTimeout() time.Duration {
	return 10 * time.Minute
}

// findProjectRoot finds the root directory of the project
func findProjectRoot() string {
	// Start from the current directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Walk up the directory tree looking for a go.mod file or a flake.nix file
	for {
		// Check if this directory has go.mod or flake.nix
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "flake.nix")); err == nil {
			return dir
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// We've reached the root
			break
		}
		dir = parent
	}

	// If we couldn't find it, return current directory
	currentDir, _ := os.Getwd()
	return currentDir
}

// StartNixChains starts multiple chains in the nix environment
func StartNixChains(t *testing.T, ctx context.Context, configs []NixChainConfig) []NixChain {
	chains := make([]NixChain, 0, len(configs))

	for i, config := range configs {
		chain, err := StartNixChain(t, ctx, config)
		if err != nil {
			t.Fatalf("Failed to start chain %s: %v", config.Name, err)
		}
		chains = append(chains, chain)
		logInfo(t, "Started chain %d: %s", i+1, config.Name)
	}

	// Wait for chains to initialize
	time.Sleep(5 * time.Second)

	return chains
}

// StartNixChain starts a chain in the nix environment
func StartNixChain(t *testing.T, ctx context.Context, config NixChainConfig) (NixChain, error) {
	// Ensure binary exists and is executable
	if _, err := os.Stat(config.BinaryPath); err != nil {
		return NixChain{}, fmt.Errorf("binary not found at %s: %w", config.BinaryPath, err)
	}

	if err := os.Chmod(config.BinaryPath, 0755); err != nil {
		return NixChain{}, fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, config.BinaryPath)

	// Set environment variables to customize behavior
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", config.HomeDir),
		fmt.Sprintf("CHAIN_ID=%s", config.Name),
		fmt.Sprintf("RPC_PORT=%s", config.RPCPort),
		fmt.Sprintf("P2P_PORT=%s", config.P2PPort),
		fmt.Sprintf("GRPC_PORT=%s", config.GRPCPort),
		fmt.Sprintf("REST_PORT=%s", config.RESTPort),
		fmt.Sprintf("VALIDATOR_COUNT=%d", config.ValidatorCount),
		fmt.Sprintf("EPOCH_LENGTH=%d", config.EpochLength),
		fmt.Sprintf("VALIDATOR_WEIGHT_CHANGE=%d", config.ValidatorWeightChangeBlocks),
	)

	if config.Byzantine {
		cmd.Env = append(cmd.Env, "BYZANTINE=true")
	}

	// Capture output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Log the chain configuration
	logInfo(t, "Starting chain %s with configuration: ValidatorCount=%d, EpochLength=%d",
		config.Name, config.ValidatorCount, config.EpochLength)

	// Start the process
	logInfo(t, "Starting chain %s with binary %s (data: %s)",
		config.Name, config.BinaryPath, config.HomeDir)

	if err := cmd.Start(); err != nil {
		return NixChain{}, fmt.Errorf("failed to start chain: %w", err)
	}

	// Create chain object
	chain := NixChain{
		Config:   config,
		Cmd:      cmd,
		Finished: false,
		T:        t,
		Process:  cmd.Process,
	}

	// Set up goroutine to monitor the process
	go func() {
		err := cmd.Wait()
		chain.Finished = true
		if err != nil && ctx.Err() == nil { // Only log if not due to context cancellation
			logInfo(t, "Chain process %s exited with error: %v", config.Name, err)
		}
	}()

	return chain, nil
}

// Kill terminates the chain process
func (c *NixChain) Kill() {
	if c.Finished {
		return
	}
	logInfo(c.T, "Killing chain process %s", c.Config.Name)
	if err := c.Cmd.Process.Kill(); err != nil {
		logInfo(c.T, "Error killing chain process: %v", err)
	}
}

// Stop gracefully stops the chain
func (c *NixChain) Stop() {
	if c.Finished {
		return
	}
	logInfo(c.T, "Stopping chain %s", c.Config.Name)
	if c.Process != nil {
		c.Process.Signal(os.Interrupt)
		time.Sleep(2 * time.Second)
	}
	c.Kill()
}

// GetNixBinaryPath gets the path to the binary in a nix environment
func GetNixBinaryPath(t *testing.T) string {
	binaryPath := os.Getenv("UFO_BIN")
	if binaryPath == "" {
		// Try to find the binary in standard locations
		projectRoot := findProjectRoot()
		binPath := filepath.Join(projectRoot, "bin")
		resultPath := filepath.Join(projectRoot, "result")
		buildPath := filepath.Join(projectRoot, "build")

		// First look in bin/ directory (preferred location)
		if _, err := os.Stat(binPath); err == nil {
			binaryPath = filepath.Join(binPath, "osmosis-ufo-patched")
			if _, err := os.Stat(binaryPath); err != nil {
				// Binary not found in bin, continue checking other locations
				t.Logf("Binary not found in preferred bin/ directory, checking other locations")
				binaryPath = ""
			}
		}

		// If not found in bin, check other locations
		if binaryPath == "" {
			if _, err := os.Stat(resultPath); err == nil {
				binaryPath = filepath.Join(resultPath, "osmosis-ufo-patched")
			} else if _, err := os.Stat(buildPath); err == nil {
				binaryPath = filepath.Join(buildPath, "osmosis-ufo-patched")
			} else {
				t.Fatalf("Cannot find binary path. Set UFO_BIN environment variable or ensure binary exists in bin/, result/ or build/")
			}
		}
	}

	// Verify binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("Binary not found at %s: %v", binaryPath, err)
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		t.Fatalf("Failed to set executable permissions: %v", err)
	}

	return binaryPath
}

// PrepareNixTestDirs creates and prepares directories for a nix-compatible test
func PrepareNixTestDirs(t *testing.T, testName string) []string {
	// Create temporary directories for data
	tmpDir, err := os.MkdirTemp("", testName)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Set up directories for test data
	chain1Dir := filepath.Join(tmpDir, "chain1")
	chain2Dir := filepath.Join(tmpDir, "chain2")
	relayerDir := filepath.Join(tmpDir, "relayer")

	for _, dir := range []string{chain1Dir, chain2Dir, relayerDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	logInfo(t, "Created test directories: chain1=%s, chain2=%s, relayer=%s",
		chain1Dir, chain2Dir, relayerDir)

	// Register cleanup in test
	t.Cleanup(func() {
		if !t.Failed() {
			os.RemoveAll(tmpDir)
		} else {
			logInfo(t, "Test failed. Temporary directory preserved at: %s", tmpDir)
		}
	})

	return []string{chain1Dir, chain2Dir, relayerDir}
}

// ConfigureNixTest sets up a test for running in the nix environment
func ConfigureNixTest(t *testing.T) {
	if !isInNixShell() {
		t.Skip("This test requires nix environment")
	}

	configureEnvironment(t)
}
