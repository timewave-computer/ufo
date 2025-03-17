package ibc

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCEnvironmentSetup tests the setup of an IBC-enabled environment with two chains.
func TestIBCEnvironmentSetup(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test directories for two chains and relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for chain1: %v", err)
	}
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for chain2: %v", err)
	}
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for relayer: %v", err)
	}
	defer os.RemoveAll(relayerDir)

	// Ensure directories exist
	for _, dir := range []string{chain1Dir, chain2Dir, relayerDir} {
		os.MkdirAll(filepath.Join(dir, "data"), 0755)
		os.MkdirAll(filepath.Join(dir, "config"), 0755)
	}

	// Set up test configuration for chain1
	chain1Config := utils.TestConfig{
		HomeDir:     chain1Dir,
		RESTAddress: "http://localhost:1317",
		RPCAddress:  "tcp://localhost:26657",
		GRPCAddress: "localhost:9090",
		ChainID:     "ibc-chain-1",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
	}

	// Set up test configuration for chain2
	chain2Config := utils.TestConfig{
		HomeDir:     chain2Dir,
		RESTAddress: "http://localhost:2317",
		RPCAddress:  "tcp://localhost:36657",
		GRPCAddress: "localhost:9190",
		ChainID:     "ibc-chain-2",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
	}

	// Create Hermes configuration
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	if err != nil {
		t.Fatalf("Failed to create Hermes configuration: %v", err)
	}

	// Start both chains
	t.Log("Starting chain1...")
	_, err = utils.SetupTestNode(ctx, chain1Config)
	if err != nil {
		t.Fatalf("Failed to setup chain1: %v", err)
	}
	defer utils.CleanupTestNode(ctx, chain1Config)

	t.Log("Starting chain2...")
	_, err = utils.SetupTestNode(ctx, chain2Config)
	if err != nil {
		t.Fatalf("Failed to setup chain2: %v", err)
	}
	defer utils.CleanupTestNode(ctx, chain2Config)

	// Wait for the chains to start producing blocks
	t.Log("Waiting for chains to start producing blocks...")
	time.Sleep(5 * time.Second)

	// Create HTTP clients for both chains
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	// Verify both chains are running
	chain1Status, err := chain1Client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain1 status: %v", err)
	}

	chain2Status, err := chain2Client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain2 status: %v", err)
	}

	chain1Height := chain1Status["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	chain2Height := chain2Status["sync_info"].(map[string]interface{})["latest_block_height"].(float64)

	t.Logf("Chain1 height: %.0f, Chain2 height: %.0f", chain1Height, chain2Height)

	assert.Greater(t, chain1Height, float64(0), "Chain1 should be producing blocks")
	assert.Greater(t, chain2Height, float64(0), "Chain2 should be producing blocks")

	// Start the Hermes relayer
	t.Log("Starting Hermes relayer...")
	relayerProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	if err != nil {
		t.Fatalf("Failed to start Hermes relayer: %v", err)
	}
	defer relayerProcess.Stop()

	// Wait for the relayer to initialize
	t.Log("Waiting for relayer to initialize...")
	time.Sleep(10 * time.Second)

	// Verify the relayer is running
	relayerStatus, err := utils.GetHermesStatus(ctx, relayerDir)
	if err != nil {
		t.Fatalf("Failed to get relayer status: %v", err)
	}

	t.Logf("Relayer status: %v", relayerStatus)
	assert.Contains(t, relayerStatus, "Running", "Relayer should be running")
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
	os.MkdirAll(filepath.Join(relayerDir, "config"), 0755)

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

// TestMultiChainIBCSetup tests the setup of an IBC environment with more than two chains.
func TestMultiChainIBCSetup(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Setup test directories for three chains and relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for chain1: %v", err)
	}
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for chain2: %v", err)
	}
	defer os.RemoveAll(chain2Dir)

	chain3Dir, err := os.MkdirTemp("", "ufo-ibc-chain3-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for chain3: %v", err)
	}
	defer os.RemoveAll(chain3Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer-")
	if err != nil {
		t.Fatalf("Failed to create temp directory for relayer: %v", err)
	}
	defer os.RemoveAll(relayerDir)

	// Ensure directories exist
	for _, dir := range []string{chain1Dir, chain2Dir, chain3Dir, relayerDir} {
		os.MkdirAll(filepath.Join(dir, "data"), 0755)
		os.MkdirAll(filepath.Join(dir, "config"), 0755)
	}

	// Set up test configurations for three chains
	chainConfigs := []utils.TestConfig{
		{
			HomeDir:     chain1Dir,
			RESTAddress: "http://localhost:1317",
			RPCAddress:  "tcp://localhost:26657",
			GRPCAddress: "localhost:9090",
			ChainID:     "ibc-multi-chain-1",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
			BlockTimeMS: 500, // Fast blocks for testing
		},
		{
			HomeDir:     chain2Dir,
			RESTAddress: "http://localhost:2317",
			RPCAddress:  "tcp://localhost:36657",
			GRPCAddress: "localhost:9190",
			ChainID:     "ibc-multi-chain-2",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
			BlockTimeMS: 500, // Fast blocks for testing
		},
		{
			HomeDir:     chain3Dir,
			RESTAddress: "http://localhost:3317",
			RPCAddress:  "tcp://localhost:46657",
			GRPCAddress: "localhost:9290",
			ChainID:     "ibc-multi-chain-3",
			BinaryType:  utils.BinaryTypeFauxmosisUfo,
			BlockTimeMS: 500, // Fast blocks for testing
		},
	}

	// Create Hermes configuration for three chains
	err = utils.CreateHermesConfig(relayerDir, chainConfigs)
	if err != nil {
		t.Fatalf("Failed to create Hermes configuration: %v", err)
	}

	// Start all three chains
	t.Log("Starting chains...")
	for i, config := range chainConfigs {
		t.Logf("Starting chain%d...", i+1)
		_, err := utils.SetupTestNode(ctx, config)
		if err != nil {
			t.Fatalf("Failed to setup chain%d: %v", i+1, err)
		}
		defer utils.CleanupTestNode(ctx, config)
	}

	// Wait for the chains to start producing blocks
	t.Log("Waiting for chains to start producing blocks...")
	time.Sleep(5 * time.Second)

	// Create HTTP clients for all chains
	clients := make([]*utils.HTTPClient, len(chainConfigs))
	for i, config := range chainConfigs {
		clients[i] = utils.NewHTTPClient(config.RESTAddress)
	}

	// Verify all chains are running
	for i, client := range clients {
		status, err := client.GetNodeStatus(ctx)
		if err != nil {
			t.Fatalf("Failed to get chain%d status: %v", i+1, err)
		}

		height := status["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
		t.Logf("Chain%d height: %.0f", i+1, height)
		assert.Greater(t, height, float64(0), "Chain%d should be producing blocks", i+1)
	}

	// Start the Hermes relayer with multi-chain support
	t.Log("Starting Hermes relayer...")
	relayerProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	if err != nil {
		t.Fatalf("Failed to start Hermes relayer: %v", err)
	}
	defer relayerProcess.Stop()

	// Wait for the relayer to initialize
	t.Log("Waiting for relayer to initialize...")
	time.Sleep(10 * time.Second)

	// Verify the relayer is running
	relayerStatus, err := utils.GetHermesStatus(ctx, relayerDir)
	if err != nil {
		t.Fatalf("Failed to get relayer status: %v", err)
	}

	t.Logf("Relayer status: %v", relayerStatus)
	assert.Contains(t, relayerStatus, "Running", "Relayer should be running")

	// Create a set of channel pairs between all chains
	channelPairs := []struct {
		sourceChain      int
		destinationChain int
	}{
		{0, 1}, // chain1 to chain2
		{0, 2}, // chain1 to chain3
		{1, 2}, // chain2 to chain3
	}

	// Create channels between all pairs
	for _, pair := range channelPairs {
		sourceConfig := chainConfigs[pair.sourceChain]
		destConfig := chainConfigs[pair.destinationChain]

		t.Logf("Creating channel between %s and %s...", sourceConfig.ChainID, destConfig.ChainID)

		// TODO: Implement channel creation command with Hermes
		// For now, we'll just log the intention
		t.Logf("Would create channel between %s and %s", sourceConfig.ChainID, destConfig.ChainID)
	}

	// In a real implementation, we would verify that all channels are created correctly
	// For now, we'll just log the test completion
	t.Log("Multi-chain IBC setup test completed")
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
