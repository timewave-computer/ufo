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

// TestMultiValidatorSetup tests a network with multiple validators.
func TestMultiValidatorSetup(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Setup test node with multi-validator configuration
	homeDir, err := os.MkdirTemp("", "ufo-multi-validator-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Ensure directories exist
	os.MkdirAll(filepath.Join(homeDir, "data"), 0755)
	os.MkdirAll(filepath.Join(homeDir, "config"), 0755)

	// Set up test configuration for multi-validator node
	config := utils.TestConfig{
		HomeDir:     homeDir,
		RESTAddress: "http://localhost:1317",
		RPCAddress:  "http://localhost:26657",
		ChainID:     "test-chain",
		BinaryType:  utils.BinaryTypeFauxmosisUfo,
		BlockTimeMS: 500, // Fast blocks for testing
		// The multi-validator setup will be handled by the node itself
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

	// Get validator set information
	validatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get validator set: %v", err)
	}

	// Verify we have multiple validators
	validators := validatorSet["validators"].([]interface{})
	require.GreaterOrEqual(t, len(validators), 4, "Expected at least 4 validators in the validator set")

	// Verify each validator has the required fields
	for i, val := range validators {
		validator := val.(map[string]interface{})
		require.NotNil(t, validator["address"], "Validator %d missing address", i)
		require.NotNil(t, validator["voting_power"], "Validator %d missing voting power", i)
		require.NotNil(t, validator["pub_key"], "Validator %d missing public key", i)

		// Log validator info
		t.Logf("Validator %d: Address: %s, Voting Power: %v",
			i, validator["address"].(string), validator["voting_power"])
	}
}

// TestValidatorRotation tests the rotation of validators in the consensus protocol.
func TestValidatorRotation(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-validator-rotation-test-")
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

	// Get initial block and proposer info
	initialBlock, err := client.GetLatestBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial block: %v", err)
	}

	initialHeight := initialBlock["block"].(map[string]interface{})["header"].(map[string]interface{})["height"].(float64)
	proposerAddress := initialBlock["block"].(map[string]interface{})["header"].(map[string]interface{})["proposer_address"].(string)
	t.Logf("Initial block height: %.0f, Proposer: %s", initialHeight, proposerAddress)

	// Wait for several blocks to be produced to observe proposer rotation
	observationPeriod := 20 * time.Second
	t.Logf("Waiting %s to observe proposer rotation...", observationPeriod)
	time.Sleep(observationPeriod)

	// Get blocks produced during the observation period
	currentBlock, err := client.GetLatestBlock(ctx)
	if err != nil {
		t.Fatalf("Failed to get current block: %v", err)
	}
	currentHeight := currentBlock["block"].(map[string]interface{})["header"].(map[string]interface{})["height"].(float64)

	// Collect proposers for all blocks in the range
	proposers := make(map[string]int)
	for height := initialHeight + 1; height <= currentHeight; height++ {
		block, err := client.GetBlockByHeight(ctx, int(height))
		if err != nil {
			t.Fatalf("Failed to get block at height %.0f: %v", height, err)
		}

		blockProposer := block["block"].(map[string]interface{})["header"].(map[string]interface{})["proposer_address"].(string)
		proposers[blockProposer]++
		t.Logf("Block height: %.0f, Proposer: %s", height, blockProposer)
	}

	// Verify that multiple proposers were used
	assert.Greater(t, len(proposers), 1, "Expected multiple different proposers to be used in rotation")

	// Log the distribution of proposers
	t.Log("Proposer distribution:")
	for proposer, count := range proposers {
		t.Logf("  Proposer %s: %d blocks", proposer, count)
	}
}

// TestVotingPowerDistribution tests that voting power is distributed correctly among validators.
func TestVotingPowerDistribution(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-voting-power-test-")
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

	// Get validator set information
	validatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get validator set: %v", err)
	}

	// Analyze voting power distribution
	validators := validatorSet["validators"].([]interface{})
	require.NotEmpty(t, validators, "Validator set is empty")

	var totalVotingPower int64
	votingPowers := make(map[string]int64)

	for _, val := range validators {
		validator := val.(map[string]interface{})
		address := validator["address"].(string)
		votingPower := int64(validator["voting_power"].(float64))

		votingPowers[address] = votingPower
		totalVotingPower += votingPower
	}

	// Verify each validator has a valid voting power
	for address, power := range votingPowers {
		require.Greater(t, power, int64(0), "Validator %s has zero or negative voting power", address)
		t.Logf("Validator %s: Voting Power: %d (%.2f%%)",
			address, power, float64(power*100)/float64(totalVotingPower))
	}

	// Verify that the total voting power is sensible
	require.Greater(t, totalVotingPower, int64(0), "Total voting power should be positive")
	t.Logf("Total voting power: %d", totalVotingPower)
}

// TestValidatorSetUpdate tests adding and removing validators.
func TestValidatorSetUpdate(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-validator-update-test-")
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

	// Get initial validator set
	initialValidatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial validator set: %v", err)
	}

	initialValidators := initialValidatorSet["validators"].([]interface{})
	t.Logf("Initial validator set size: %d", len(initialValidators))

	// Create a new validator
	newValidatorKey, err := client.CreateValidatorKey(ctx, "new-validator")
	if err != nil {
		t.Fatalf("Failed to create new validator key: %v", err)
	}

	// Add the new validator to the validator set
	err = client.AddValidator(ctx, newValidatorKey, 10) // 10 is the voting power
	if err != nil {
		t.Fatalf("Failed to add new validator: %v", err)
	}

	// Wait for the validator set update to be processed
	t.Log("Waiting for validator set update to be processed...")
	time.Sleep(10 * time.Second)

	// Verify the new validator was added
	updatedValidatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get updated validator set: %v", err)
	}

	updatedValidators := updatedValidatorSet["validators"].([]interface{})
	t.Logf("Updated validator set size: %d", len(updatedValidators))

	// The updated validator set should include the new validator
	assert.Greater(t, len(updatedValidators), len(initialValidators),
		"Expected validator set to grow after adding a validator")

	// Find the new validator in the updated set
	found := false
	for _, val := range updatedValidators {
		validator := val.(map[string]interface{})
		pubKey := validator["pub_key"].(map[string]interface{})["value"].(string)
		if pubKey == newValidatorKey {
			found = true
			break
		}
	}
	assert.True(t, found, "Could not find the newly added validator in the validator set")

	// Now remove the validator
	err = client.RemoveValidator(ctx, newValidatorKey)
	if err != nil {
		t.Fatalf("Failed to remove validator: %v", err)
	}

	// Wait for the validator removal to be processed
	t.Log("Waiting for validator removal to be processed...")
	time.Sleep(10 * time.Second)

	// Verify the validator was removed
	finalValidatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get final validator set: %v", err)
	}

	finalValidators := finalValidatorSet["validators"].([]interface{})
	t.Logf("Final validator set size: %d", len(finalValidators))

	// The final validator set should be back to the original size
	assert.Equal(t, len(initialValidators), len(finalValidators),
		"Expected validator set to return to original size after removing the validator")

	// Verify the removed validator is no longer in the set
	for _, val := range finalValidators {
		validator := val.(map[string]interface{})
		pubKey := validator["pub_key"].(map[string]interface{})["value"].(string)
		assert.NotEqual(t, pubKey, newValidatorKey,
			"Removed validator still present in validator set")
	}
}
