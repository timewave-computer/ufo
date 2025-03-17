package consensus

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

// TestBlockProduction tests the basic block production capability of UFO.
func TestBlockProduction(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-block-production-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Ensure directories exist
	os.MkdirAll(filepath.Join(homeDir, "data"), 0755)
	os.MkdirAll(filepath.Join(homeDir, "config"), 0755)

	// Set up test configuration
	config := utils.TestConfig{
		HomeDir:     homeDir,
		RESTAddress: "http://localhost:1317",
		RPCAddress:  "http://localhost:26657",
		ChainID:     "test-chain",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
	}

	// Start the node
	err = utils.SetupTestNode(ctx, config)
	if err != nil {
		t.Fatalf("Failed to setup test node: %v", err)
	}
	defer utils.CleanupTestNode(ctx, config)

	// Wait for the node to start producing blocks
	t.Log("Waiting for the node to start producing blocks...")
	time.Sleep(5 * time.Second)

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Get initial block height
	initialStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status: %v", err)
	}

	initialHeight := initialStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Initial block height: %.0f", initialHeight)

	// Wait for a few more blocks to be produced
	t.Log("Waiting for more blocks to be produced...")
	time.Sleep(5 * time.Second)

	// Get updated block height
	updatedStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get updated node status: %v", err)
	}

	updatedHeight := updatedStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Updated block height: %.0f", updatedHeight)

	// Verify that new blocks were produced
	assert.Greater(t, updatedHeight, initialHeight, "Expected new blocks to be produced")
}

// TestBlockValidation tests that produced blocks are valid according to the consensus rules.
func TestBlockValidation(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-block-validation-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Ensure directories exist
	os.MkdirAll(filepath.Join(homeDir, "data"), 0755)
	os.MkdirAll(filepath.Join(homeDir, "config"), 0755)

	// Set up test configuration
	config := utils.TestConfig{
		HomeDir:     homeDir,
		RESTAddress: "http://localhost:1317",
		RPCAddress:  "http://localhost:26657",
		ChainID:     "test-chain",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
	}

	// Start the node
	err = utils.SetupTestNode(ctx, config)
	if err != nil {
		t.Fatalf("Failed to setup test node: %v", err)
	}
	defer utils.CleanupTestNode(ctx, config)

	// Wait for the node to start producing blocks
	t.Log("Waiting for the node to start producing blocks...")
	time.Sleep(5 * time.Second)

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Get the latest block
	latestBlock, err := client.GetLatestBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}

	// Verify the block has the required fields
	require.NotNil(t, latestBlock["block"], "Block data is missing")

	// Get the block header
	header := latestBlock["block"].(map[string]interface{})["header"].(map[string]interface{})

	// Check required header fields
	require.NotNil(t, header["height"], "Block height is missing")
	require.NotNil(t, header["time"], "Block time is missing")
	require.NotNil(t, header["chain_id"], "Chain ID is missing")
	require.NotNil(t, header["last_block_id"], "Last block ID is missing")

	// Verify chain ID matches our configuration
	assert.Equal(t, config.ChainID, header["chain_id"].(string), "Chain ID mismatch")

	// Get the previous block
	height := int(header["height"].(float64))
	if height > 1 {
		prevBlock, err := client.GetBlockByHeight(ctx, height-1)
		if err != nil {
			t.Fatalf("Failed to get previous block: %v", err)
		}

		// Verify the current block references the previous block correctly
		prevBlockID := prevBlock["block"].(map[string]interface{})["header"].(map[string]interface{})["block_id"].(map[string]interface{})
		lastBlockID := header["last_block_id"].(map[string]interface{})

		assert.Equal(t, prevBlockID["hash"], lastBlockID["hash"], "Previous block hash mismatch")
	}
}

// TestBlockCommitSignatures tests that blocks have valid commit signatures from validators.
func TestBlockCommitSignatures(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-block-commit-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Ensure directories exist
	os.MkdirAll(filepath.Join(homeDir, "data"), 0755)
	os.MkdirAll(filepath.Join(homeDir, "config"), 0755)

	// Set up test configuration
	config := utils.TestConfig{
		HomeDir:     homeDir,
		RESTAddress: "http://localhost:1317",
		RPCAddress:  "http://localhost:26657",
		ChainID:     "test-chain",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
	}

	// Start the node
	err = utils.SetupTestNode(ctx, config)
	if err != nil {
		t.Fatalf("Failed to setup test node: %v", err)
	}
	defer utils.CleanupTestNode(ctx, config)

	// Wait for the node to start producing blocks
	t.Log("Waiting for the node to start producing blocks...")
	time.Sleep(5 * time.Second)

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Get the latest block
	latestBlock, err := client.GetLatestBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}

	// Verify the block has a valid commit with signatures
	require.NotNil(t, latestBlock["block"], "Block data is missing")
	lastCommit := latestBlock["block"].(map[string]interface{})["last_commit"]
	require.NotNil(t, lastCommit, "Last commit data is missing")

	// Check for signatures in the commit
	signatures := lastCommit.(map[string]interface{})["signatures"]
	require.NotNil(t, signatures, "Commit signatures are missing")

	// Make sure we have at least one signature (single validator case)
	sigArray := signatures.([]interface{})
	assert.GreaterOrEqual(t, len(sigArray), 1, "Expected at least one validator signature")

	// Check that each signature has the required fields
	for i, sig := range sigArray {
		sigData := sig.(map[string]interface{})
		require.NotNil(t, sigData["validator_address"], "Validator address missing in signature %d", i)
		require.NotNil(t, sigData["signature"], "Signature data missing in signature %d", i)
	}
}

// TestConsensusConfigurationParameters tests different consensus configuration parameters.
func TestConsensusConfigurationParameters(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Test different block times to verify configuration is respected
	blockTimes := []int{1000, 500, 200} // Block times in milliseconds

	for _, blockTimeMS := range blockTimes {
		t.Run("BlockTime-"+string(rune(blockTimeMS)), func(t *testing.T) {
			// Setup test node
			homeDir, err := os.MkdirTemp("", "ufo-consensus-config-test-")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(homeDir)

			// Ensure directories exist
			os.MkdirAll(filepath.Join(homeDir, "data"), 0755)
			os.MkdirAll(filepath.Join(homeDir, "config"), 0755)

			// Set up test configuration with specific block time
			config := utils.TestConfig{
				HomeDir:     homeDir,
				RESTAddress: "http://localhost:1317",
				RPCAddress:  "http://localhost:26657",
				ChainID:     "test-chain",
				BinaryType:  utils.BinaryTypeFauxmosisUfo,
				BlockTimeMS: blockTimeMS,
			}

			// Start the node
			err = utils.SetupTestNode(ctx, config)
			if err != nil {
				t.Fatalf("Failed to setup test node: %v", err)
			}
			defer utils.CleanupTestNode(ctx, config)

			// Wait for the node to start producing blocks
			t.Log("Waiting for the node to start producing blocks...")
			time.Sleep(3 * time.Second)

			// Create HTTP client
			client := utils.NewHTTPClient(config.RESTAddress)

			// Get initial block height
			initialStatus, err := client.GetNodeStatus(ctx)
			if err != nil {
				t.Fatalf("Failed to get node status: %v", err)
			}

			initialHeight := initialStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
			t.Logf("Initial block height: %.0f", initialHeight)

			// Wait for a specified period to observe block production rate
			observationPeriod := 10 * time.Second
			t.Logf("Waiting %s to observe block production rate...", observationPeriod)
			time.Sleep(observationPeriod)

			// Get updated block height
			updatedStatus, err := client.GetNodeStatus(ctx)
			if err != nil {
				t.Fatalf("Failed to get updated node status: %v", err)
			}

			updatedHeight := updatedStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
			t.Logf("Updated block height: %.0f", updatedHeight)

			// Calculate blocks produced during observation period
			blocksProduced := updatedHeight - initialHeight
			expectedBlocks := float64(observationPeriod.Milliseconds()) / float64(blockTimeMS)

			// Allow for some variance in timing
			// The number of blocks produced should be roughly proportional to the block time
			minExpectedBlocks := 0.7 * expectedBlocks
			maxExpectedBlocks := 1.3 * expectedBlocks

			t.Logf("Blocks produced: %.0f, Expected range: %.1f - %.1f blocks (block time: %d ms)",
				blocksProduced, minExpectedBlocks, maxExpectedBlocks, blockTimeMS)

			assert.GreaterOrEqual(t, blocksProduced, minExpectedBlocks,
				"Block production rate too slow for block time %d ms", blockTimeMS)
			assert.LessOrEqual(t, blocksProduced, maxExpectedBlocks,
				"Block production rate too fast for block time %d ms", blockTimeMS)
		})
	}
}
