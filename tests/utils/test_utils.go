package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Binary type constants
const (
	BinaryTypeFauxmosisComet    = "fauxmosis-comet"
	BinaryTypeFauxmosisUfo      = "fauxmosis-ufo"
	BinaryTypeOsmosisUfoBridged = "osmosis-ufo-bridged"
	BinaryTypeOsmosisUfoPatched = "osmosis-ufo-patched"
)

// TestConfig represents the configuration for running a test
type TestConfig struct {
	BinaryType      string // Options: fauxmosis-comet, fauxmosis-ufo, osmosis-ufo-bridged, osmosis-ufo-patched
	HomeDir         string
	RPCAddress      string
	RESTAddress     string
	GRPCAddress     string
	WebSocketURL    string
	NodeID          string
	ChainID         string
	BlockTimeMS     int
	LogLevel        string
	TestTimeoutSecs int
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig(binaryType string) TestConfig {
	homeDir := filepath.Join(os.TempDir(), fmt.Sprintf("ufo-test-%s-%d", binaryType, time.Now().Unix()))

	return TestConfig{
		BinaryType:      binaryType,
		HomeDir:         homeDir,
		RPCAddress:      "tcp://localhost:26657",
		RESTAddress:     "http://localhost:1317",
		GRPCAddress:     "localhost:9090",
		WebSocketURL:    "ws://localhost:26657/websocket",
		NodeID:          "testnode",
		ChainID:         "test-chain",
		BlockTimeMS:     1000, // Default to 1 second blocks
		LogLevel:        "info",
		TestTimeoutSecs: 60, // Default timeout of 60 seconds
	}
}

// GetBinaryPath returns the path to the binary for the specified type
func GetBinaryPath(binaryType string) (string, error) {
	// In a real environment, we'd look for the binaries in a standard location
	// For our test suite, we'll use the binaries in the project's bin directory
	binaries := map[string]string{
		"fauxmosis-comet":     "fauxmosis-comet",
		"fauxmosis-ufo":       "fauxmosis-ufo",
		"osmosis-ufo-bridged": "osmosisd-bridged",
		"osmosis-ufo-patched": "osmosisd-patched",
	}

	binary, ok := binaries[binaryType]
	if !ok {
		return "", fmt.Errorf("unknown binary type: %s", binaryType)
	}

	// First, look in the bin directory
	projectRoot := getProjectRoot()
	fmt.Printf("Project root: %s\n", projectRoot)

	// Check for the binary in multiple locations
	// 1. First in the bin directory
	binPath := filepath.Join(projectRoot, "bin", binary)
	fmt.Printf("Checking for binary at: %s\n", binPath)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	// 2. Next, look in the tests/bin directory
	binPath = filepath.Join(projectRoot, "tests", "bin", binary)
	fmt.Printf("Checking for binary at: %s\n", binPath)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	// 3. Check if the binary is available in the PATH
	fmt.Printf("Looking for binary in PATH: %s\n", binary)
	path, err := exec.LookPath(binary)
	if err == nil {
		return path, nil
	}

	// 4. If not in PATH, look in the project's root directory
	binPath = filepath.Join(projectRoot, binary)
	fmt.Printf("Checking for binary at: %s\n", binPath)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	// Last resort - try the current directory
	wd, _ := os.Getwd()
	binPath = filepath.Join(wd, binary)
	fmt.Printf("Checking for binary at: %s\n", binPath)
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	return "", fmt.Errorf("binary %s not found in any expected location", binary)
}

// getProjectRoot attempts to find the root directory of the project
func getProjectRoot() string {
	// First try to use git to find the repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	// If git command fails, try to use the working directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Check if we're in a subdirectory of the project
	for {
		// Check if this directory contains tests/bin or bin
		if _, err := os.Stat(filepath.Join(dir, "tests", "bin")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "bin")); err == nil {
			return dir
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// If all else fails, return the current directory
	return "."
}

// SetupTestNode initializes and starts a test node with the specified configuration
func SetupTestNode(ctx context.Context, config TestConfig) error {
	// Get the binary path
	binaryPath, err := GetBinaryPath(config.BinaryType)
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	fmt.Printf("Using binary at: %s\n", binaryPath)

	// Create home directory if it doesn't exist
	if err := os.MkdirAll(config.HomeDir, 0755); err != nil {
		return fmt.Errorf("failed to create home directory: %w", err)
	}

	// Create bin/data directories
	if err := os.MkdirAll(filepath.Join(config.HomeDir, "data"), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(config.HomeDir, "config"), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize the node
	fmt.Printf("Initializing node with command: %s init %s --chain-id %s --home %s\n",
		binaryPath, config.NodeID, config.ChainID, config.HomeDir)

	initCmd := exec.CommandContext(ctx, binaryPath, "init", config.NodeID, "--chain-id", config.ChainID, "--home", config.HomeDir)
	if output, err := initCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to initialize node: %w, output: %s", err, string(output))
	}

	// Start the node
	// This command would vary based on the binary type
	// For now, we'll use a simplified approach
	fmt.Printf("Starting node with command: %s start --home %s --rpc.laddr %s --grpc.address %s\n",
		binaryPath, config.HomeDir, config.RPCAddress, config.GRPCAddress)

	startCmd := exec.CommandContext(ctx, binaryPath, "start", "--home", config.HomeDir, "--rpc.laddr", config.RPCAddress, "--grpc.address", config.GRPCAddress)

	// Start the node in the background
	if err := startCmd.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// Wait for the node to be ready
	// In a real implementation, we'd have a more robust way to check if the node is ready
	fmt.Println("Waiting for node to start...")
	time.Sleep(5 * time.Second)
	fmt.Println("Node should be ready now")

	return nil
}

// CleanupTestNode stops the node and cleans up temporary files
func CleanupTestNode(ctx context.Context, config TestConfig) error {
	// Get the binary path
	binaryPath, err := GetBinaryPath(config.BinaryType)
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	// Stop the node
	// This command would vary based on the binary type
	// For now, we'll use a simplified approach
	stopCmd := exec.CommandContext(ctx, binaryPath, "tendermint", "unsafe-reset-all", "--home", config.HomeDir)
	if output, err := stopCmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to stop node: %v, output: %s\n", err, string(output))
		// Continue with cleanup even if stopping the node fails
	}

	// Clean up temporary files
	if err := os.RemoveAll(config.HomeDir); err != nil {
		return fmt.Errorf("failed to clean up home directory: %w", err)
	}

	return nil
}
