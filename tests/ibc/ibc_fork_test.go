package ibc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCFork tests IBC behavior during a chain fork
func TestIBCFork(t *testing.T) {
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

	// Perform IBC transfer from chain A to chain B
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

	// Get current height of chain A
	status, err := chainA.HTTPClient.GetNodeStatus(ctx)
	require.NoError(t, err, "Failed to get node status for chain A")

	syncInfo, ok := status["sync_info"].(map[string]interface{})
	require.True(t, ok, "Failed to get sync_info from node status")

	latestHeight, ok := syncInfo["latest_block_height"].(string)
	require.True(t, ok, "Failed to get latest_block_height from sync_info")

	var currentHeight int
	_, err = fmt.Sscanf(latestHeight, "%d", &currentHeight)
	require.NoError(t, err, "Failed to parse latest_block_height")

	// Simulate a fork at the current height
	t.Log("Simulating a fork at the current height")
	err = chainA.HTTPClient.SimulateFork(ctx, currentHeight)
	require.NoError(t, err, "Failed to simulate fork")

	// Wait for a few blocks to see if the chain continues to make progress
	t.Log("Waiting to see if the chain continues to make progress after fork")
	err = utils.WaitForBlockHeight(ctx, chainA.HTTPClient, currentHeight+5, 60*time.Second)
	require.NoError(t, err, "Chain A failed to make progress after fork")

	// Get clients before the fork
	t.Log("Getting clients before the fork")
	clientsBeforeFork, err := hermes.GetClients(ctx, chainB.ID)
	require.NoError(t, err, "Failed to get clients on chain B before fork")

	// Update the client on chain B to reflect the fork on chain A
	t.Log("Updating the client on chain B to reflect the fork on chain A")
	for _, clientID := range clientsBeforeFork {
		err = hermes.UpdateClient(ctx, chainB.ID, clientID)
		require.NoError(t, err, "Failed to update client on chain B")
	}

	// Perform another IBC transfer from chain A to chain B after the fork
	t.Log("Performing another IBC transfer from chain A to chain B after the fork")
	transferAmount = "2000"
	err = utils.PerformIBCTransfer(ctx, hermes, chainA, chainB, transferAmount, "stake")
	require.NoError(t, err, "Failed to perform IBC transfer after fork")

	// Wait for the transfer to be processed
	t.Log("Waiting for IBC transfer to be processed")
	time.Sleep(15 * time.Second)

	// Verify the transfer by checking the balance on chain B
	t.Log("Verifying IBC transfer after fork")
	err = utils.VerifyIBCTransfer(ctx, chainB, transferAmount, chainA.ID, "stake")
	require.NoError(t, err, "Failed to verify IBC transfer after fork")

	// Clean up the test environment
	t.Log("Cleaning up test environment")
	utils.CleanupTestEnvironment(chainA, chainB, relayerCmd)
}
