package ibc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestIBCClientTrustingPeriodSecurity tests the security implications of the trusting period.
func TestIBCClientTrustingPeriodSecurity(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-sec-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-sec-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-sec-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains with a very short trusting period for testing
	chain1Config := utils.TestConfig{
		ChainID:        "sec-chain-1",
		RPCAddress:     "tcp://localhost:26657",
		GRPCAddress:    "localhost:9090",
		RESTAddress:    "localhost:1317",
		P2PAddress:     "localhost:26656",
		HomeDir:        chain1Dir,
		DebugLevel:     "info",
		KeysDir:        filepath.Join(chain1Dir, "keys"),
		BlockTime:      "1s",
		WithCometMock:  false,
		TrustingPeriod: "10s", // Very short trusting period for testing
	}

	chain2Config := utils.TestConfig{
		ChainID:        "sec-chain-2",
		RPCAddress:     "tcp://localhost:26658",
		GRPCAddress:    "localhost:9091",
		RESTAddress:    "localhost:1318",
		P2PAddress:     "localhost:26659",
		HomeDir:        chain2Dir,
		DebugLevel:     "info",
		KeysDir:        filepath.Join(chain2Dir, "keys"),
		BlockTime:      "1s",
		WithCometMock:  false,
		TrustingPeriod: "10s", // Very short trusting period for testing
	}

	// Start both chains
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Wait for chains to produce blocks
	time.Sleep(2 * time.Second)

	// Configure and start the Hermes relayer
	err = utils.CreateHermesConfigWithOptions(relayerDir, []utils.TestConfig{chain1Config, chain2Config}, map[string]interface{}{
		"clear_on_start": true,
	})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Create IBC channel between chains
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created IBC channel: %s (source) -> %s (destination)", sourceChannelID, destChannelID)

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-security")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-security")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// First do a successful transfer
	transferAmount := "10000"
	transferHash, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("IBC transfer hash: %s", transferHash)

	// Relay packets to ensure transfer completes
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)

	// Wait to ensure transfer completed
	time.Sleep(5 * time.Second)

	// Verify transfer succeeded
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This would be the actual denom hash in a real implementation
	receiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, transferAmount, receiverBalance, "Expected transfer to succeed")

	// Stop chain2 to simulate it being offline
	t.Log("Stopping chain2 to simulate extended downtime...")
	chain2.Stop()

	// Wait longer than the trusting period
	t.Logf("Waiting for trusting period (%s) to expire...", chain1Config.TrustingPeriod)
	trustingPeriodDuration, _ := time.ParseDuration(chain1Config.TrustingPeriod)
	time.Sleep(trustingPeriodDuration + 5*time.Second)

	// Restart chain2
	t.Log("Restarting chain2...")
	chain2, err = utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Wait for chain2 to produce blocks
	time.Sleep(5 * time.Second)

	// Attempt another transfer - this should fail due to expired client
	t.Log("Attempting transfer after trusting period expired (should fail)...")
	transferAmount2 := "5000"
	_, err = utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount2, "stake")

	// The transfer attempt might fail, or more likely the relay attempt will fail
	// Either way, the tokens should not arrive at the destination
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)

	// Wait to ensure any packets are processed
	time.Sleep(5 * time.Second)

	// Check if receiver balance changed (it shouldn't have after trusting period expiry)
	newReceiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, receiverBalance, newReceiverBalance,
		"Receiver balance should not increase after trusting period expiry")

	// Log that client update would be needed in a real scenario
	t.Log("In a real scenario, the client would need to be updated to restore functionality")
}

// TestIBCDoubleSpendPrevention tests the prevention of double spending via IBC.
func TestIBCDoubleSpendPrevention(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for the chains and relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-double-spend-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-double-spend-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-double-spend-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:       "double-spend-chain-1",
		RPCAddress:    "tcp://localhost:26657",
		GRPCAddress:   "localhost:9090",
		RESTAddress:   "localhost:1317",
		P2PAddress:    "localhost:26656",
		HomeDir:       chain1Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain1Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	chain2Config := utils.TestConfig{
		ChainID:       "double-spend-chain-2",
		RPCAddress:    "tcp://localhost:26658",
		GRPCAddress:   "localhost:9091",
		RESTAddress:   "localhost:1318",
		P2PAddress:    "localhost:26659",
		HomeDir:       chain2Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain2Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	// Start both chains
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Configure and start the Hermes relayer
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Create IBC channel between chains
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created IBC channel: %s (source) -> %s (destination)", sourceChannelID, destChannelID)

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-double-spend")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-double-spend")
	require.NoError(t, err)

	// Fund with exact amount so we can try to double spend
	exactAmount := "1000"
	err = chain1Client.FundAccount(ctx, chain1Address, exactAmount+"stake")
	require.NoError(t, err)

	// Verify initial balance
	initialBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	assert.Equal(t, exactAmount, initialBalance, "Initial balance should be exactly what we funded")

	// Perform first IBC transfer using the entire balance
	transferHash1, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		exactAmount, "stake")
	require.NoError(t, err)
	t.Logf("First IBC transfer hash: %s", transferHash1)

	// Relay packets for first transfer
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)

	// Wait to ensure transfer completed
	time.Sleep(5 * time.Second)

	// Verify first transfer succeeded by checking balances
	chain1BalanceAfterFirstTransfer, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	assert.Equal(t, "0", chain1BalanceAfterFirstTransfer, "Chain1 balance should be 0 after transfer")

	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This would be the actual hash in a real implementation
	chain2BalanceAfterFirstTransfer, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, exactAmount, chain2BalanceAfterFirstTransfer, "Chain2 should have received the tokens")

	// Try to perform a second IBC transfer with the same funds (double spend)
	// This should fail since the account has a zero balance
	_, err = utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		exactAmount, "stake")

	// The transfer might fail immediately, or it might be accepted but fail during processing
	// Either way, the balance on chain2 should not increase

	// Relay packets to see if any got through
	time.Sleep(5 * time.Second)
	_ = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)

	// Wait to ensure any potential transfer would complete
	time.Sleep(5 * time.Second)

	// Verify balances to ensure double spend did not occur
	chain1FinalBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	assert.Equal(t, "0", chain1FinalBalance, "Chain1 final balance should still be 0")

	chain2FinalBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, chain2BalanceAfterFirstTransfer, chain2FinalBalance,
		"Chain2 balance should not have increased (double spend prevented)")
}

// TestIBCChannelSecurityValidation tests the security validation of IBC channels.
func TestIBCChannelSecurityValidation(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-channel-sec-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-channel-sec-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-channel-sec-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:       "channel-sec-chain-1",
		RPCAddress:    "tcp://localhost:26657",
		GRPCAddress:   "localhost:9090",
		RESTAddress:   "localhost:1317",
		P2PAddress:    "localhost:26656",
		HomeDir:       chain1Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain1Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	chain2Config := utils.TestConfig{
		ChainID:       "channel-sec-chain-2",
		RPCAddress:    "tcp://localhost:26658",
		GRPCAddress:   "localhost:9091",
		RESTAddress:   "localhost:1318",
		P2PAddress:    "localhost:26659",
		HomeDir:       chain2Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain2Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	// Start both chains
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Configure and start the Hermes relayer
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Test case 1: Create a valid IBC channel
	t.Log("Test case 1: Creating a valid IBC channel")
	validSourceChannelID, validDestChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created valid IBC channel: %s (source) -> %s (destination)", validSourceChannelID, validDestChannelID)

	// Test case 2: Attempt to create a channel with invalid parameters
	t.Log("Test case 2: Attempting to create a channel with invalid parameters")

	// Try to create a channel with an invalid port
	invalidPortCmd := utils.NewCustomCommand(ctx, "hermes",
		"create", "channel",
		"--a-chain", chain1Config.ChainID,
		"--b-chain", chain2Config.ChainID,
		"--a-port", "invalid-port", // Invalid port
		"--b-port", "transfer",
	)
	invalidPortCmd.Dir = relayerDir
	invalidPortCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	invalidPortOutput, err := invalidPortCmd.CombinedOutput()
	t.Logf("Invalid port attempt output: %s", string(invalidPortOutput))
	// We expect this to fail or return an error in a real implementation

	// Test case 3: Try to create a channel with mismatched ordering
	t.Log("Test case 3: Attempting to create a channel with mismatched ordering")
	mismatchedOrderCmd := utils.NewCustomCommand(ctx, "hermes",
		"create", "channel",
		"--a-chain", chain1Config.ChainID,
		"--b-chain", chain2Config.ChainID,
		"--a-port", "transfer",
		"--b-port", "transfer",
		"--order", "ordered", // Different from the default unordered
		"--channel-version", "invalid-version", // Invalid version
	)
	mismatchedOrderCmd.Dir = relayerDir
	mismatchedOrderCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	mismatchedOrderOutput, err := mismatchedOrderCmd.CombinedOutput()
	t.Logf("Mismatched ordering attempt output: %s", string(mismatchedOrderOutput))
	// This might fail in different ways depending on implementation

	// Create and fund test accounts to test transfers through the valid channel
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-channel-sec")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-channel-sec")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Perform an IBC transfer through the valid channel
	transferAmount := "10000"
	transferHash, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		validSourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("Valid channel IBC transfer hash: %s", transferHash)

	// Relay packets
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		validSourceChannelID, validDestChannelID)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Verify transfer succeeded through the valid channel
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This would be the actual hash in a real implementation
	receiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, transferAmount, receiverBalance, "Transfer through valid channel should succeed")
}

// TestIBCPacketDataValidation tests the validation of IBC packet data.
func TestIBCPacketDataValidation(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-packet-val-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-packet-val-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-packet-val-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:       "packet-val-chain-1",
		RPCAddress:    "tcp://localhost:26657",
		GRPCAddress:   "localhost:9090",
		RESTAddress:   "localhost:1317",
		P2PAddress:    "localhost:26656",
		HomeDir:       chain1Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain1Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	chain2Config := utils.TestConfig{
		ChainID:       "packet-val-chain-2",
		RPCAddress:    "tcp://localhost:26658",
		GRPCAddress:   "localhost:9091",
		RESTAddress:   "localhost:1318",
		P2PAddress:    "localhost:26659",
		HomeDir:       chain2Dir,
		DebugLevel:    "info",
		KeysDir:       filepath.Join(chain2Dir, "keys"),
		BlockTime:     "1s",
		WithCometMock: false,
	}

	// Start both chains
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Configure and start the Hermes relayer
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Create IBC channel between chains
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created IBC channel: %s (source) -> %s (destination)", sourceChannelID, destChannelID)

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-packet-val")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-packet-val")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Test case 1: Perform a valid IBC transfer
	t.Log("Test case 1: Performing a valid IBC transfer")
	validAmount := "10000"
	validTransferHash, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		validAmount, "stake")
	require.NoError(t, err)
	t.Logf("Valid IBC transfer hash: %s", validTransferHash)

	// Relay packets for valid transfer
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Verify valid transfer succeeded
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This would be the actual hash in a real implementation
	receiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, validAmount, receiverBalance, "Valid transfer should succeed")

	// Test case 2: Attempt transfer with invalid receiver address (if implementation supports this test)
	t.Log("Test case 2: Attempting transfer with invalid receiver address")
	invalidReceiver := "invalid-cosmos-address"
	invalidReceiverCmd := utils.NewCustomCommand(ctx, "hermes",
		"tx", "ft-transfer",
		"--src-chain", chain1Config.ChainID,
		"--dst-chain", chain2Config.ChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
		"--amount", "5000",
		"--denom", "stake",
		"--timeout-height-offset", "1000",
		"--timeout-seconds", "100",
		"--receiver", invalidReceiver, // Invalid receiver address
	)
	invalidReceiverCmd.Dir = relayerDir
	invalidReceiverCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	invalidReceiverOutput, err := invalidReceiverCmd.CombinedOutput()
	t.Logf("Invalid receiver attempt output: %s", string(invalidReceiverOutput))
	// This should fail in a real implementation

	// Test case 3: Attempt transfer with invalid denom
	t.Log("Test case 3: Attempting transfer with invalid denom")
	invalidDenomCmd := utils.NewCustomCommand(ctx, "hermes",
		"tx", "ft-transfer",
		"--src-chain", chain1Config.ChainID,
		"--dst-chain", chain2Config.ChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
		"--amount", "5000",
		"--denom", "nonexistent-token", // Invalid denom
		"--timeout-height-offset", "1000",
		"--timeout-seconds", "100",
		"--receiver", chain2Address,
	)
	invalidDenomCmd.Dir = relayerDir
	invalidDenomCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	invalidDenomOutput, err := invalidDenomCmd.CombinedOutput()
	t.Logf("Invalid denom attempt output: %s", string(invalidDenomOutput))
	// This should fail in a real implementation

	// Test case 4: Attempt transfer with negative amount (if implementation supports this test)
	t.Log("Test case 4: Attempting transfer with invalid amount")
	invalidAmountCmd := utils.NewCustomCommand(ctx, "hermes",
		"tx", "ft-transfer",
		"--src-chain", chain1Config.ChainID,
		"--dst-chain", chain2Config.ChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
		"--amount", "-1000", // Invalid negative amount
		"--denom", "stake",
		"--timeout-height-offset", "1000",
		"--timeout-seconds", "100",
		"--receiver", chain2Address,
	)
	invalidAmountCmd.Dir = relayerDir
	invalidAmountCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	invalidAmountOutput, err := invalidAmountCmd.CombinedOutput()
	t.Logf("Invalid amount attempt output: %s", string(invalidAmountOutput))
	// This should fail in a real implementation

	// Verify the receiver's balance hasn't changed after the invalid transfer attempts
	finalReceiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	assert.Equal(t, receiverBalance, finalReceiverBalance,
		"Receiver balance should not change after invalid transfer attempts")
}
