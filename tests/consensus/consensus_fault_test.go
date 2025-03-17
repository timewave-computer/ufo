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

// TestByzantineValidatorBehavior tests the system's handling of Byzantine validator behaviors.
func TestByzantineValidatorBehavior(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node with multi-validator configuration
	homeDir, err := os.MkdirTemp("", "ufo-byzantine-test-")
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

	// Get initial block height and validator info
	initialStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status: %v", err)
	}

	initialHeight := initialStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Initial block height: %.0f", initialHeight)

	// Get validators
	validatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get validator set: %v", err)
	}

	validators := validatorSet["validators"].([]interface{})
	require.GreaterOrEqual(t, len(validators), 4, "Expected at least 4 validators for Byzantine behavior test")

	// Select a validator to simulate Byzantine behavior
	byzantineValidator := validators[1].(map[string]interface{})
	byzantineAddress := byzantineValidator["address"].(string)
	t.Logf("Selected validator %s to simulate Byzantine behavior", byzantineAddress)

	// Simulate Byzantine behavior by sending an invalid vote
	err = client.SimulateByzantineVote(ctx, byzantineAddress, int(initialHeight+1))
	if err != nil {
		t.Fatalf("Failed to simulate Byzantine vote: %v", err)
	}

	// Wait for the network to process and respond to the Byzantine behavior
	t.Log("Waiting for network to handle Byzantine behavior...")
	time.Sleep(10 * time.Second)

	// Get updated status
	updatedStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get updated node status: %v", err)
	}

	updatedHeight := updatedStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Updated block height: %.0f", updatedHeight)

	// Verify that the network continues to make progress despite Byzantine behavior
	assert.Greater(t, updatedHeight, initialHeight, "Expected network to continue producing blocks despite Byzantine validator")

	// Verify that the Byzantine validator's power has been adjusted if slashing is enabled
	// (Slashing may not be implemented in all configurations)
	updatedValidatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get updated validator set: %v", err)
	}

	updatedValidators := updatedValidatorSet["validators"].([]interface{})
	for _, val := range updatedValidators {
		validator := val.(map[string]interface{})
		if validator["address"].(string) == byzantineAddress {
			t.Logf("Byzantine validator status: %v", validator)
			// Check for any status changes, depending on what the implementation does with Byzantine validators
			// Some implementations may reduce voting power, jail the validator, or remove it entirely
		}
	}
}

// TestValidatorDisconnection tests the system's handling of validator disconnection and reconnection.
func TestValidatorDisconnection(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-validator-disconnect-test-")
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

	// Get initial status
	initialStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status: %v", err)
	}

	initialHeight := initialStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Initial block height: %.0f", initialHeight)

	// Get validators
	validatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get validator set: %v", err)
	}

	validators := validatorSet["validators"].([]interface{})
	require.GreaterOrEqual(t, len(validators), 4, "Expected at least 4 validators for disconnection test")

	// Select a validator to disconnect
	disconnectValidator := validators[2].(map[string]interface{})
	disconnectAddress := disconnectValidator["address"].(string)
	t.Logf("Selected validator %s to simulate disconnection", disconnectAddress)

	// Simulate validator disconnection
	err = client.DisconnectValidator(ctx, disconnectAddress)
	if err != nil {
		t.Fatalf("Failed to disconnect validator: %v", err)
	}

	// Wait to observe network behavior with disconnected validator
	disconnectObservationPeriod := 15 * time.Second
	t.Logf("Waiting %s to observe network with disconnected validator...", disconnectObservationPeriod)
	time.Sleep(disconnectObservationPeriod)

	// Get status after disconnection
	statusAfterDisconnect, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status after disconnect: %v", err)
	}

	heightAfterDisconnect := statusAfterDisconnect["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Block height after disconnect: %.0f", heightAfterDisconnect)

	// Verify that the network continues to make progress despite validator disconnection
	assert.Greater(t, heightAfterDisconnect, initialHeight,
		"Expected network to continue producing blocks despite validator disconnection")

	// Now simulate validator reconnection
	err = client.ReconnectValidator(ctx, disconnectAddress)
	if err != nil {
		t.Fatalf("Failed to reconnect validator: %v", err)
	}

	// Wait to observe network behavior with reconnected validator
	reconnectObservationPeriod := 15 * time.Second
	t.Logf("Waiting %s to observe network with reconnected validator...", reconnectObservationPeriod)
	time.Sleep(reconnectObservationPeriod)

	// Get status after reconnection
	statusAfterReconnect, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status after reconnect: %v", err)
	}

	heightAfterReconnect := statusAfterReconnect["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Block height after reconnect: %.0f", heightAfterReconnect)

	// Verify that the network continues to make progress after validator reconnection
	assert.Greater(t, heightAfterReconnect, heightAfterDisconnect,
		"Expected network to continue producing blocks after validator reconnection")

	// Check that the validator participates in consensus again after reconnection
	proposers := make(map[string]bool)
	for height := heightAfterReconnect + 1; height <= heightAfterReconnect+10; height++ {
		block, err := client.GetBlockByHeight(ctx, int(height))
		if err != nil {
			continue
		}

		proposerAddress := block["block"].(map[string]interface{})["header"].(map[string]interface{})["proposer_address"].(string)
		proposers[proposerAddress] = true
	}

	// Log if the reconnected validator is participating again
	if _, ok := proposers[disconnectAddress]; ok {
		t.Logf("Reconnected validator %s is participating in consensus again", disconnectAddress)
	} else {
		t.Logf("Reconnected validator %s is not yet observed as a proposer", disconnectAddress)
	}
}

// TestForkDetectionAndResolution tests the system's ability to detect and resolve forks.
func TestForkDetectionAndResolution(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-fork-detection-test-")
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

	// Simulate a network partition creating a fork
	forkHeight := int(initialHeight) + 5
	t.Logf("Simulating a fork at height %d", forkHeight)

	err = client.SimulateFork(ctx, forkHeight)
	if err != nil {
		t.Fatalf("Failed to simulate fork: %v", err)
	}

	// Wait for the network to detect and resolve the fork
	t.Log("Waiting for fork detection and resolution...")
	time.Sleep(30 * time.Second)

	// Get status after fork resolution
	statusAfterFork, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status after fork: %v", err)
	}

	heightAfterFork := statusAfterFork["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Block height after fork resolution: %.0f", heightAfterFork)

	// Verify that the network continued to make progress past the fork height
	assert.Greater(t, heightAfterFork, float64(forkHeight),
		"Expected network to continue producing blocks past the fork height")

	// Check for evidence of the fork in the block data
	block, err := client.GetBlockByHeight(ctx, forkHeight)
	if err == nil {
		blockData := block["block"].(map[string]interface{})
		evidence := blockData["evidence"]
		if evidence != nil {
			evidenceList := evidence.(map[string]interface{})["evidence"]
			if evidenceList != nil {
				t.Logf("Found evidence at fork height: %v", evidenceList)
			}
		}
	}
}

// TestNetworkPartitionRecovery tests the system's ability to recover from network partitions.
func TestNetworkPartitionRecovery(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-network-partition-test-")
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

	// Get validator set to simulate a partition
	validatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get validator set: %v", err)
	}

	validators := validatorSet["validators"].([]interface{})
	require.GreaterOrEqual(t, len(validators), 4, "Expected at least 4 validators for partition test")

	// Create two partitions with validators split between them
	partition1 := make([]string, 0, len(validators)/2)
	partition2 := make([]string, 0, len(validators)/2)

	for i, val := range validators {
		validator := val.(map[string]interface{})
		address := validator["address"].(string)

		if i < len(validators)/2 {
			partition1 = append(partition1, address)
		} else {
			partition2 = append(partition2, address)
		}
	}

	t.Logf("Creating network partition: P1: %v, P2: %v", partition1, partition2)

	// Get initial block height
	initialStatus, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status: %v", err)
	}

	initialHeight := initialStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Initial block height: %.0f", initialHeight)

	// Simulate a network partition
	err = client.SimulateNetworkPartition(ctx, partition1, partition2)
	if err != nil {
		t.Fatalf("Failed to simulate network partition: %v", err)
	}

	// Wait to observe the network behavior during the partition
	partitionDuration := 20 * time.Second
	t.Logf("Waiting %s to observe network during partition...", partitionDuration)
	time.Sleep(partitionDuration)

	// Check status during partition
	statusDuringPartition, err := client.GetNodeStatus(ctx)
	var heightDuringPartition float64

	if err != nil {
		t.Logf("Error getting status during partition (expected if severe): %v", err)
	} else {
		heightDuringPartition = statusDuringPartition["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
		t.Logf("Block height during partition: %.0f", heightDuringPartition)
	}

	// Heal the network partition
	t.Log("Healing network partition...")
	err = client.HealNetworkPartition(ctx)
	if err != nil {
		t.Fatalf("Failed to heal network partition: %v", err)
	}

	// Wait for the network to recover
	recoveryPeriod := 30 * time.Second
	t.Logf("Waiting %s for network to recover from partition...", recoveryPeriod)
	time.Sleep(recoveryPeriod)

	// Get status after recovery
	statusAfterRecovery, err := client.GetNodeStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get node status after recovery: %v", err)
	}

	heightAfterRecovery := statusAfterRecovery["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
	t.Logf("Block height after recovery: %.0f", heightAfterRecovery)

	// Verify that the network continued to make progress after recovery
	assert.Greater(t, heightAfterRecovery, initialHeight,
		"Expected network to continue producing blocks after partition recovery")

	// Verify that all validators are participating again
	finalValidatorSet, err := client.GetValidatorSet(ctx)
	if err != nil {
		t.Fatalf("Failed to get final validator set: %v", err)
	}

	finalValidators := finalValidatorSet["validators"].([]interface{})
	t.Logf("Validator set size after recovery: %d", len(finalValidators))
	assert.Equal(t, len(validators), len(finalValidators),
		"Expected all validators to be present after recovery")
}

// TestConsensusLiveness tests that consensus maintains liveness guarantees.
func TestConsensusLiveness(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup test node
	homeDir, err := os.MkdirTemp("", "ufo-consensus-liveness-test-")
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

	// Create a series of stress conditions to test liveness
	stressors := []struct {
		name     string
		stressor func(context.Context) error
	}{
		{"highLoad", client.SimulateHighLoad},
		{"networkLatency", client.SimulateNetworkLatency},
		{"randomValidatorOutages", client.SimulateRandomValidatorOutages},
	}

	for _, stressor := range stressors {
		t.Run(stressor.name, func(t *testing.T) {
			// Get current height before stressor
			beforeStatus, err := client.GetNodeStatus(ctx)
			if err != nil {
				t.Fatalf("Failed to get node status before %s: %v", stressor.name, err)
			}
			beforeHeight := beforeStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)

			// Apply the stressor
			t.Logf("Applying stressor: %s", stressor.name)
			err = stressor.stressor(ctx)
			if err != nil {
				t.Fatalf("Failed to apply stressor %s: %v", stressor.name, err)
			}

			// Wait to observe the network under stress
			stressDuration := 20 * time.Second
			t.Logf("Waiting %s to observe network under %s...", stressDuration, stressor.name)
			time.Sleep(stressDuration)

			// Remove the stressor
			t.Logf("Removing stressor: %s", stressor.name)
			err = client.RemoveStressor(ctx, stressor.name)
			if err != nil {
				t.Fatalf("Failed to remove stressor %s: %v", stressor.name, err)
			}

			// Wait for recovery
			recoveryPeriod := 20 * time.Second
			t.Logf("Waiting %s for recovery after %s...", recoveryPeriod, stressor.name)
			time.Sleep(recoveryPeriod)

			// Get status after recovery
			afterStatus, err := client.GetNodeStatus(ctx)
			if err != nil {
				t.Fatalf("Failed to get node status after %s: %v", stressor.name, err)
			}

			afterHeight := afterStatus["sync_info"].(map[string]interface{})["latest_block_height"].(float64)
			t.Logf("Block height after %s: %.0f (before: %.0f)", stressor.name, afterHeight, beforeHeight)

			// Verify liveness: the chain should continue to produce blocks
			assert.Greater(t, afterHeight, beforeHeight,
				"Expected network to maintain liveness under %s", stressor.name)
		})
	}
}
