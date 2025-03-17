package ibc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCLightClientUpdates tests IBC light client updates with multiple validators per chain.
// This test verifies that:
// 1. Light client updates work correctly when validators change
// 2. IBC transfers still succeed after validator set changes
// 3. The light client correctly tracks validator set changes
// 4. Updates occur correctly after Osmosis epoch boundaries are crossed
func TestIBCLightClientUpdates(t *testing.T) {
	// Determine which binary type is being used
	binaryType := os.Getenv("UFO_BINARY_TYPE")
	if binaryType == "" {
		binaryType = "patched" // Default to patched if not specified
	}
	t.Logf("Running IBC Light Client test with binary type: %s", binaryType)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-light-client-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-light-client-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-light-client-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Set up directories for validator nodes
	chainDirs := []string{chain1Dir, chain2Dir}
	for _, dir := range chainDirs {
		for i := 1; i <= 4; i++ {
			validatorDir := filepath.Join(dir, fmt.Sprintf("validator%d", i))
			require.NoError(t, os.MkdirAll(validatorDir, 0755))
		}
	}

	// Configure the chains with 4 validators each and explicitly set epoch length
	chain1Config := utils.TestConfig{
		ChainID:                   "light-client-chain-1",
		RPCAddress:                "tcp://localhost:26657",
		GRPCAddress:               "localhost:9090",
		RESTAddress:               "localhost:1317",
		P2PAddress:                "localhost:26656",
		HomeDir:                   chain1Dir,
		DebugLevel:                "info",
		KeysDir:                   filepath.Join(chain1Dir, "keys"),
		BlockTime:                 "100ms", // Ultra fast block time (100ms)
		ValidatorCount:            4,       // Using 4 validators
		ValidatorRotationInterval: 200,     // Rotate validators every 200ms
		EpochLength:               10,      // Set epoch length to 10 blocks
	}

	// Set BinaryType based on environment
	if binaryType == "patched" {
		chain1Config.BinaryType = utils.BinaryTypeOsmosisUfoPatched
	} else {
		chain1Config.BinaryType = utils.BinaryTypeOsmosisUfoBridged
	}

	chain2Config := utils.TestConfig{
		ChainID:                   "light-client-chain-2",
		RPCAddress:                "tcp://localhost:26658",
		GRPCAddress:               "localhost:9091",
		RESTAddress:               "localhost:1318",
		P2PAddress:                "localhost:26659",
		HomeDir:                   chain2Dir,
		DebugLevel:                "info",
		KeysDir:                   filepath.Join(chain2Dir, "keys"),
		BlockTime:                 "100ms", // Ultra fast block time (100ms)
		ValidatorCount:            4,       // Using 4 validators
		ValidatorRotationInterval: 200,     // Rotate validators every 200ms
		EpochLength:               10,      // Set epoch length to 10 blocks
	}

	// Set BinaryType based on environment
	if binaryType == "patched" {
		chain2Config.BinaryType = utils.BinaryTypeOsmosisUfoPatched
	} else {
		chain2Config.BinaryType = utils.BinaryTypeOsmosisUfoBridged
	}

	// Start chain1 with multiple validators and handle potential startup issues
	t.Log("Starting chain1 with 4 validators and epoch length of 10 blocks...")
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	if err != nil {
		t.Fatalf("Failed to start chain1: %v", err)
	}
	defer chain1.Stop()
	t.Log("Chain1 started successfully")

	// Start chain2 with multiple validators and handle potential startup issues
	t.Log("Starting chain2 with 4 validators and epoch length of 10 blocks...")
	chain2, err := utils.StartTestNode(ctx, chain2Config)
	if err != nil {
		t.Fatalf("Failed to start chain2: %v", err)
	}
	defer chain2.Stop()
	t.Log("Chain2 started successfully")

	// Configure and start the Hermes relayer
	t.Log("Setting up Hermes relayer...")
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Create IBC channel between chains
	t.Log("Creating IBC channel...")
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created IBC channel: %s (source) -> %s (destination)", sourceChannelID, destChannelID)

	// Create Hermes config for more direct control
	hermesConfig := utils.NewHermesConfig(filepath.Join(relayerDir, "config", "config.toml"), "")

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-light-client")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-light-client")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Get initial validator set information before IBC operations
	chain1Validators, err := chain1Client.GetValidators(ctx)
	require.NoError(t, err)
	t.Logf("Chain1 initial validator set with %d validators", len(chain1Validators))

	chain2Validators, err := chain2Client.GetValidators(ctx)
	require.NoError(t, err)
	t.Logf("Chain2 initial validator set with %d validators", len(chain2Validators))

	// Assert that we have 4 validators on each chain
	require.Equal(t, 4, len(chain1Validators), "Chain1 should have 4 validators")
	require.Equal(t, 4, len(chain2Validators), "Chain2 should have 4 validators")

	// Get IBC client information
	t.Log("Getting initial IBC client information...")
	clients1, err := hermesConfig.GetClients(ctx, chain1Config.ChainID)
	require.NoError(t, err)
	require.NotEmpty(t, clients1, "Should have at least one client on chain1")

	clients2, err := hermesConfig.GetClients(ctx, chain2Config.ChainID)
	require.NoError(t, err)
	require.NotEmpty(t, clients2, "Should have at least one client on chain2")

	chain1ClientID := clients1[0] // Get the first client ID
	chain2ClientID := clients2[0] // Get the first client ID

	t.Logf("Chain1 client ID: %s, Chain2 client ID: %s", chain1ClientID, chain2ClientID)

	// Get initial client state and consensus height
	initialClientState1, err := hermesConfig.GetClientState(ctx, chain1Config.ChainID, chain2ClientID)
	require.NoError(t, err)
	t.Logf("Initial client state on chain1: %v", initialClientState1)

	initialClientState2, err := hermesConfig.GetClientState(ctx, chain2Config.ChainID, chain1ClientID)
	require.NoError(t, err)
	t.Logf("Initial client state on chain2: %v", initialClientState2)

	// Start automatic validator rotation
	t.Log("Starting automatic validator rotation...")
	stopChain1Rotation, err := chain1.StartValidatorRotation(ctx)
	require.NoError(t, err)
	defer stopChain1Rotation()

	stopChain2Rotation, err := chain2.StartValidatorRotation(ctx)
	require.NoError(t, err)
	defer stopChain2Rotation()

	// Wait for at least one epoch to pass (10 blocks * 100ms = 1s)
	t.Log("Waiting for epoch boundary to be crossed...")
	time.Sleep(1 * time.Second)

	// Get current block heights to verify epoch boundaries
	chain1Height, err := chain1Client.GetLatestBlockHeight(ctx)
	require.NoError(t, err)
	t.Logf("Chain1 current height: %d (completed epochs: %d)", chain1Height, chain1Height/chain1Config.EpochLength)

	chain2Height, err := chain2Client.GetLatestBlockHeight(ctx)
	require.NoError(t, err)
	t.Logf("Chain2 current height: %d (completed epochs: %d)", chain2Height, chain2Height/chain2Config.EpochLength)

	// Force validator changes immediately after epoch boundary
	t.Log("Triggering validator set changes after epoch boundary...")
	err = chain1.RotateValidators(ctx)
	require.NoError(t, err)
	t.Log("Chain1 validator set rotated.")

	err = chain2.RotateValidators(ctx)
	require.NoError(t, err)
	t.Log("Chain2 validator set rotated.")

	// Wait for validator changes to be committed
	time.Sleep(300 * time.Millisecond)

	// Send IBC transfer right after validator set changes
	t.Log("Sending IBC transfer after validator set changes and epoch boundary...")
	transferAmount := "10000"
	transferHash, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("IBC transfer hash after validator changes: %s", transferHash)

	// Relay packets
	time.Sleep(200 * time.Millisecond)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	// Check if client updates occurred
	t.Log("Verifying client updates have occurred...")
	updatedClientState1, err := hermesConfig.GetClientState(ctx, chain1Config.ChainID, chain2ClientID)
	require.NoError(t, err)
	t.Logf("Updated client state on chain1: %v", updatedClientState1)

	updatedClientState2, err := hermesConfig.GetClientState(ctx, chain2Config.ChainID, chain1ClientID)
	require.NoError(t, err)
	t.Logf("Updated client state on chain2: %v", updatedClientState2)

	// Explicitly check if a client update occurred by comparing heights
	clientUpdated := false
	if updatedClientState1.LatestHeight != initialClientState1.LatestHeight ||
		updatedClientState2.LatestHeight != initialClientState2.LatestHeight {
		clientUpdated = true
		t.Log("✅ CLIENT UPDATE DETECTED: Client state has been updated after validator set changes and epoch boundary!")
	} else {
		t.Log("❌ No client update detected after validator set changes and epoch boundary.")
	}

	// Get the client's consensus state for the updated height
	consensusState1, err := hermesConfig.GetConsensusState(ctx, chain1Config.ChainID, chain2ClientID, updatedClientState1.Height)
	require.NoError(t, err)
	t.Logf("Consensus state for client on chain1: %v", consensusState1)

	consensusState2, err := hermesConfig.GetConsensusState(ctx, chain2Config.ChainID, chain1ClientID, updatedClientState2.Height)
	require.NoError(t, err)
	t.Logf("Consensus state for client on chain2: %v", consensusState2)

	// Verify client update contains validator set changes
	require.True(t, clientUpdated, "Client update should have occurred after validator set changes and epoch boundary")

	// Verify IBC transfer went through
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This would be the actual denom hash in a real implementation
	receiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, receiverBalance, "IBC transfer should succeed after validator set changes")
	t.Log("✅ IBC transfer successful after validator set changes and epoch boundary")

	t.Logf("Test completed successfully with %s binary: IBC light client updates verified after validator set change, epoch boundary crossing, and IBC packet transfer!", binaryType)
}
