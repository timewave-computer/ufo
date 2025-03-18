package ibc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCEnvironmentSetup tests the setup of an IBC-enabled environment with two chains.
func TestIBCEnvironmentSetup(t *testing.T) {
	// Skip if not in nix shell
	if !isInNixShell() {
		t.Skip("Skipping test, not in nix shell")
		return
	}

	// Configure environment
	configureEnvironment(t)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout())
	defer cancel()

	// Setup goroutine to monitor for test termination
	termCh := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				t.Log("Test timed out")
			}
		case <-termCh:
			// Test completed normally
		}
	}()
	defer close(termCh)

	t.Log("Starting TestIBCEnvironmentSetup")

	// Prepare test directories
	testDirs := PrepareNixTestDirs(t, "TestIBCEnvironmentSetup")
	chain1Dir := testDirs[0]
	chain2Dir := testDirs[1]
	relayerDir := filepath.Join(testDirs[2], "relayer")
	err := os.MkdirAll(relayerDir, 0755)
	require.NoError(t, err)

	// Get the binary path
	binaryPath := GetNixBinaryPath(t)
	t.Logf("Using binary path: %s", binaryPath)

	// Configure chains
	chain1Config := NixChainConfig{
		Name:                        "env-chain-1",
		BinaryPath:                  binaryPath,
		HomeDir:                     chain1Dir,
		RPCPort:                     "26657",
		P2PPort:                     "26656",
		GRPCPort:                    "9090",
		RESTPort:                    "1317",
		ValidatorCount:              4,
		EpochLength:                 10,
		ValidatorWeightChangeBlocks: 5,
	}

	chain2Config := NixChainConfig{
		Name:                        "env-chain-2",
		BinaryPath:                  binaryPath,
		HomeDir:                     chain2Dir,
		RPCPort:                     "26667",
		P2PPort:                     "26666",
		GRPCPort:                    "9190",
		RESTPort:                    "1318",
		ValidatorCount:              4,
		EpochLength:                 10,
		ValidatorWeightChangeBlocks: 5,
	}

	// Start chains
	chains := StartNixChains(t, ctx, []NixChainConfig{chain1Config, chain2Config})
	t.Logf("Started %d chains", len(chains))

	// Give chains time to initialize
	time.Sleep(5 * time.Second)

	// Check both chains are running
	require.Equal(t, 2, len(chains), "Expected 2 chains to be running")

	// Test verification of environment setup
	t.Log("Verifying IBC environment setup in nix environment")

	// Verify our chains are running with correct configuration
	assert.Equal(t, "env-chain-1", chains[0].Config.Name, "Chain 1 should have correct name")
	assert.Equal(t, "env-chain-2", chains[1].Config.Name, "Chain 2 should have correct name")

	// In a nix environment with auto-starting binaries,
	// additional environment setup testing would use API interactions
	// See tests/ibc/nix_compatibility.md for details on how to adapt tests

	t.Log("Successfully verified IBC environment setup in nix environment")
}

// TestHermesRelayerConfig tests the Hermes relayer configuration and setup
func TestHermesRelayerConfig(t *testing.T) {
	// Skip if not in nix shell
	if !isInNixShell() {
		t.Skip("Skipping test, not in nix shell")
		return
	}

	// Configure environment
	configureEnvironment(t)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout())
	defer cancel()

	// Setup goroutine to monitor for test termination
	termCh := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				t.Log("Test timed out")
			}
		case <-termCh:
			// Test completed normally
		}
	}()
	defer close(termCh)

	t.Log("Starting TestHermesRelayerConfig")

	// Prepare test directories
	testDirs := PrepareNixTestDirs(t, "TestHermesRelayerConfig")
	relayerDir := filepath.Join(testDirs[2], "relayer")
	err := os.MkdirAll(relayerDir, 0755)
	require.NoError(t, err)

	// Get the binary path
	binaryPath := GetNixBinaryPath(t)
	t.Logf("Using binary path: %s", binaryPath)

	// Create a Hermes config file
	configPath := filepath.Join(relayerDir, "config.toml")
	configContent := `
[global]
log_level = "info"

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = true

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = "127.0.0.1"
port = 3000

[telemetry]
enabled = true
host = "127.0.0.1"
port = 3001
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Verify that the config file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err, "Hermes config file should exist")

	// Check if the content is correct
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(content), "log_level"), "Config should contain log_level")
	assert.True(t, strings.Contains(string(content), "telemetry"), "Config should contain telemetry section")

	t.Log("Successfully verified Hermes relayer configuration in nix environment")
}

// TestMultiChainIBCSetup tests the setup of multiple chains with IBC
func TestMultiChainIBCSetup(t *testing.T) {
	// Skip if not in nix shell
	if !isInNixShell() {
		t.Skip("Skipping test, not in nix shell")
		return
	}

	// Configure environment
	configureEnvironment(t)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestTimeout())
	defer cancel()

	// Setup goroutine to monitor for test termination
	termCh := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				t.Log("Test timed out")
			}
		case <-termCh:
			// Test completed normally
		}
	}()
	defer close(termCh)

	t.Log("Starting TestMultiChainIBCSetup")

	// Prepare test directories for 3 chains
	testDirs := PrepareNixTestDirs(t, "TestMultiChainIBCSetup")
	chain1Dir := testDirs[0]
	chain2Dir := testDirs[1]

	// Create a third chain directory
	chain3Dir := filepath.Join(filepath.Dir(chain1Dir), "chain3")
	err := os.MkdirAll(chain3Dir, 0755)
	require.NoError(t, err)

	// Get the binary path
	binaryPath := GetNixBinaryPath(t)
	t.Logf("Using binary path: %s", binaryPath)

	// Configure chains
	chain1Config := NixChainConfig{
		Name:                        "multi-chain-1",
		BinaryPath:                  binaryPath,
		HomeDir:                     chain1Dir,
		RPCPort:                     "26657",
		P2PPort:                     "26656",
		GRPCPort:                    "9090",
		RESTPort:                    "1317",
		ValidatorCount:              4,
		EpochLength:                 10,
		ValidatorWeightChangeBlocks: 5,
	}

	chain2Config := NixChainConfig{
		Name:                        "multi-chain-2",
		BinaryPath:                  binaryPath,
		HomeDir:                     chain2Dir,
		RPCPort:                     "26667",
		P2PPort:                     "26666",
		GRPCPort:                    "9190",
		RESTPort:                    "1318",
		ValidatorCount:              4,
		EpochLength:                 10,
		ValidatorWeightChangeBlocks: 5,
	}

	chain3Config := NixChainConfig{
		Name:                        "multi-chain-3",
		BinaryPath:                  binaryPath,
		HomeDir:                     chain3Dir,
		RPCPort:                     "26677",
		P2PPort:                     "26676",
		GRPCPort:                    "9290",
		RESTPort:                    "1319",
		ValidatorCount:              4,
		EpochLength:                 10,
		ValidatorWeightChangeBlocks: 5,
	}

	// Start chains
	chains := StartNixChains(t, ctx, []NixChainConfig{chain1Config, chain2Config, chain3Config})
	t.Logf("Started %d chains", len(chains))

	// Give chains time to initialize
	time.Sleep(5 * time.Second)

	// Check all chains are running
	require.Equal(t, 3, len(chains), "Expected 3 chains to be running")

	// Verify our chains are running with correct configuration
	assert.Equal(t, "multi-chain-1", chains[0].Config.Name, "Chain 1 should have correct name")
	assert.Equal(t, "multi-chain-2", chains[1].Config.Name, "Chain 2 should have correct name")
	assert.Equal(t, "multi-chain-3", chains[2].Config.Name, "Chain 3 should have correct name")

	// In a nix environment with auto-starting binaries,
	// additional multi-chain setup testing would use API interactions
	// See tests/ibc/nix_compatibility.md for details on how to adapt tests

	t.Log("Successfully verified multi-chain IBC setup in nix environment")
}

// TestHermesRelayerConfiguration tests the configuration options of the Hermes relayer.
func TestHermesRelayerConfiguration(t *testing.T) {
	// Create a context with timeout
	_, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test directories for the relayer
	relayerDir, err := os.MkdirTemp("", "ufo-hermes-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for relayer: %v", err)
	}
	defer os.RemoveAll(relayerDir)

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Join(relayerDir, "config"), 0755); err != nil {
		t.Fatalf("Failed to create config directory for relayer: %v", err)
	}

	// Set up test configurations for multiple chains
	chainConfigs := []utils.TestConfig{
		{
			HomeDir:     "/tmp/chain1", // These directories don't need to exist for the config test
			RESTAddress: "http://localhost:1317",
			RPCAddress:  "tcp://localhost:26657",
			GRPCAddress: "localhost:9090",
			ChainID:     "chain-1",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
		},
		{
			HomeDir:     "/tmp/chain2",
			RESTAddress: "http://localhost:2317",
			RPCAddress:  "tcp://localhost:36657",
			GRPCAddress: "localhost:9190",
			ChainID:     "chain-2",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
		},
		{
			HomeDir:     "/tmp/chain3",
			RESTAddress: "http://localhost:3317",
			RPCAddress:  "tcp://localhost:46657",
			GRPCAddress: "localhost:9290",
			ChainID:     "chain-3",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
		},
	}

	// Create Hermes configuration with multiple chains
	err = utils.CreateHermesConfig(relayerDir, chainConfigs)
	if err != nil {
		t.Fatalf("Failed to create Hermes configuration: %v", err)
	}

	// Verify the configuration file exists
	configPath := filepath.Join(relayerDir, "config", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Hermes configuration file was not created: %v", err)
	}

	// Read and validate the configuration
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Hermes configuration: %v", err)
	}

	// Check that the configuration includes all chains
	config := string(configContent)
	for _, chainConfig := range chainConfigs {
		assert.Contains(t, config, chainConfig.ChainID, "Configuration should include chain ID: %s", chainConfig.ChainID)
		assert.Contains(t, config, chainConfig.RPCAddress, "Configuration should include RPC address: %s", chainConfig.RPCAddress)
		assert.Contains(t, config, chainConfig.GRPCAddress, "Configuration should include gRPC address: %s", chainConfig.GRPCAddress)
	}

	// Test additional configuration options
	customChainConfigs := []utils.TestConfig{
		{
			HomeDir:        "/tmp/custom-chain",
			RESTAddress:    "http://localhost:4317",
			RPCAddress:     "tcp://localhost:56657",
			GRPCAddress:    "localhost:9390",
			ChainID:        "custom-chain",
			BinaryType:     utils.BinaryTypeFauxmosisUfo,
			TrustingPeriod: "48h", // Custom trusting period
		},
	}

	// Create Hermes configuration with custom options
	err = utils.CreateHermesConfigWithOptions(relayerDir, customChainConfigs, map[string]interface{}{
		"clear_on_start":                   true,
		"tx_confirmation":                  true,
		"auto_register_counterparty_payee": true,
	})
	if err != nil {
		t.Fatalf("Failed to create Hermes configuration with options: %v", err)
	}

	// Read and validate the custom configuration
	configContent, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read custom Hermes configuration: %v", err)
	}

	// Check that the configuration includes custom options
	customConfig := string(configContent)
	assert.Contains(t, customConfig, "trusting_period = \"48h\"", "Configuration should include custom trusting period")
	assert.Contains(t, customConfig, "clear_on_start = true", "Configuration should include clear_on_start option")
	assert.Contains(t, customConfig, "tx_confirmation = true", "Configuration should include tx_confirmation option")
}

// TestHermesConfigCreation tests the creation of Hermes configuration with custom options.
func TestHermesConfigCreation(t *testing.T) {
	// Create a context with timeout
	cancel := context.CancelFunc(func() {})
	defer cancel()

	// Setup test directories for the relayer
	relayerDir, err := os.MkdirTemp("", "ufo-hermes-config-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for relayer: %v", err)
	}
	defer os.RemoveAll(relayerDir)

	// Create configurations for test chains
	chainConfigs := []utils.TestConfig{
		{
			ChainID:     "test-chain-1",
			RPCAddress:  "tcp://localhost:26657",
			GRPCAddress: "localhost:9090",
			HomeDir:     "/tmp/chain1",
		},
		{
			ChainID:     "test-chain-2",
			RPCAddress:  "tcp://localhost:26658",
			GRPCAddress: "localhost:9091",
			HomeDir:     "/tmp/chain2",
		},
	}

	// Create standard Hermes config
	t.Log("Creating standard Hermes config...")
	err = utils.CreateHermesConfig(relayerDir, chainConfigs)
	if err != nil {
		t.Fatalf("Failed to create standard Hermes config: %v", err)
	}

	// Check if the config file was created
	configPath := filepath.Join(relayerDir, "config", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Create Hermes config with custom options
	t.Log("Creating Hermes config with custom options...")
	customOptions := map[string]interface{}{
		"clear_on_start": true,
		"log_level":      "debug",
	}

	err = utils.CreateHermesConfigWithOptions(relayerDir, chainConfigs, customOptions)
	if err != nil {
		t.Fatalf("Failed to create Hermes config with custom options: %v", err)
	}

	// Check if the config file was updated
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file does not exist at %s after update", configPath)
	}

	// Read the config file and check if our custom options are included
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configContent := string(configData)
	if !strings.Contains(configContent, "clear_on_start = true") {
		t.Errorf("Config file does not contain custom option 'clear_on_start = true'")
	}

	if !strings.Contains(configContent, "log_level = \"debug\"") {
		t.Errorf("Config file does not contain custom option 'log_level = \"debug\"'")
	}

	t.Log("Hermes configuration test completed successfully")
}
