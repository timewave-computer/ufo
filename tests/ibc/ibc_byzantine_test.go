package ibc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCByzantineValidators tests IBC behavior with Byzantine validators
func TestIBCByzantineValidators(t *testing.T) {
	ctx := context.Background()

	// Set up the chains with IBC
	chainA, chainB, hermes, err := utils.SetupIBCChains(ctx, "ufo", "ufo", "hermes/config.toml")
	require.NoError(t, err, "Failed to set up IBC chains")

	// Create IBC clients between the chains
	t.Log("Setting up IBC clients")
	err = utils.SetupIBCConnection(ctx, hermes, chainA, chainB)
	require.NoError(t, err, "Failed to set up IBC connection")

	// Create IBC channel
	t.Log("Setting up IBC channel")
	err = utils.SetupIBCChannel(ctx, hermes, chainA, chainB)
	require.NoError(t, err, "Failed to set up IBC channel")

	// Start the relayer
	t.Log("Starting the relayer")
	relayerCmd, err := utils.SetupRelayer(ctx, hermes, chainA, chainB)
	require.NoError(t, err, "Failed to start the relayer")

	// Wait for a few blocks to ensure the relayer is working
	err = utils.WaitForBlockHeight(ctx, chainA.HTTPClient, 5, 30*time.Second)
	require.NoError(t, err, "Failed to wait for blocks on chain A")

	err = utils.WaitForBlockHeight(ctx, chainB.HTTPClient, 5, 30*time.Second)
	require.NoError(t, err, "Failed to wait for blocks on chain B")

	// Get validator set for chain A
	t.Log("Getting validator set for chain A")
	validatorSetA, err := chainA.HTTPClient.GetValidatorSet(ctx)
	require.NoError(t, err, "Failed to get validator set for chain A")
	require.NotNil(t, validatorSetA, "Validator set for chain A is nil")

	// Get the first validator from chain A
	validators, ok := validatorSetA["validators"].([]interface{})
	require.True(t, ok, "Failed to get validators from validator set")
	require.NotEmpty(t, validators, "No validators found in validator set")

	validator, ok := validators[0].(map[string]interface{})
	require.True(t, ok, "Failed to get validator from validators")

	validatorAddress, ok := validator["address"].(string)
	require.True(t, ok, "Failed to get validator address")
	require.NotEmpty(t, validatorAddress, "Validator address is empty")

	// Simulate Byzantine behavior by having a validator send an invalid vote
	t.Log("Simulating Byzantine behavior by having a validator send an invalid vote")
	currentHeight := 10 // Assuming we're at least at height 10
	err = chainA.HTTPClient.SimulateByzantineVote(ctx, validatorAddress, currentHeight)
	require.NoError(t, err, "Failed to simulate Byzantine vote")

	// Wait for a few blocks to see if the chain continues to make progress
	t.Log("Waiting to see if the chain continues to make progress")
	err = utils.WaitForBlockHeight(ctx, chainA.HTTPClient, currentHeight+5, 60*time.Second)
	require.NoError(t, err, "Chain A failed to make progress after Byzantine vote")

	// Perform IBC transfer from chain A to chain B to verify IBC still works
	t.Log("Performing IBC transfer from chain A to chain B")
	transferAmount := "1000"
	err = utils.PerformIBCTransfer(ctx, hermes, chainA, chainB, transferAmount, "stake")
	require.NoError(t, err, "Failed to perform IBC transfer")

	// Wait for the transfer to be processed
	t.Log("Waiting for IBC transfer to be processed")
	time.Sleep(15 * time.Second)

	// Verify the transfer by checking the balance on chain B
	t.Log("Verifying IBC transfer")
	err = utils.VerifyIBCTransfer(ctx, chainB, transferAmount, chainA.ID, "stake")
	require.NoError(t, err, "Failed to verify IBC transfer")

	// Simulate network partition
	t.Log("Simulating network partition")
	partition1 := []string{validatorAddress}
	partition2 := []string{} // All other validators

	// Get the rest of the validators
	for i, v := range validators {
		if i == 0 {
			continue // Skip the first validator as it's already in partition1
		}
		val, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		addr, ok := val["address"].(string)
		if !ok {
			continue
		}
		partition2 = append(partition2, addr)
	}

	err = chainA.HTTPClient.SimulateNetworkPartition(ctx, partition1, partition2)
	require.NoError(t, err, "Failed to simulate network partition")

	// Wait to see if the chain continues to make progress
	t.Log("Waiting to see if the chain continues to make progress after network partition")
	err = utils.WaitForBlockHeight(ctx, chainA.HTTPClient, currentHeight+10, 60*time.Second)
	require.NoError(t, err, "Chain A failed to make progress after network partition")

	// Heal the network partition
	t.Log("Healing network partition")
	err = chainA.HTTPClient.HealNetworkPartition(ctx)
	require.NoError(t, err, "Failed to heal network partition")

	// Wait for a few blocks to ensure the network is healed
	t.Log("Waiting for a few blocks to ensure the network is healed")
	err = utils.WaitForBlockHeight(ctx, chainA.HTTPClient, currentHeight+15, 60*time.Second)
	require.NoError(t, err, "Chain A failed to make progress after healing network partition")

	// Perform another IBC transfer to verify IBC still works after healing
	t.Log("Performing another IBC transfer to verify IBC still works after healing")
	transferAmount = "2000"
	err = utils.PerformIBCTransfer(ctx, hermes, chainA, chainB, transferAmount, "stake")
	require.NoError(t, err, "Failed to perform IBC transfer after healing")

	// Wait for the transfer to be processed
	t.Log("Waiting for IBC transfer to be processed")
	time.Sleep(15 * time.Second)

	// Verify the transfer by checking the balance on chain B
	t.Log("Verifying IBC transfer")
	err = utils.VerifyIBCTransfer(ctx, chainB, transferAmount, chainA.ID, "stake")
	require.NoError(t, err, "Failed to verify IBC transfer after healing")

	// Clean up the test environment
	t.Log("Cleaning up test environment")
	utils.CleanupTestEnvironment(chainA, chainB, relayerCmd)
}
