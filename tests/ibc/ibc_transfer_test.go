package ibc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/timewave/ufo/tests/utils"
)

// TestBasicIBCTransfer tests a basic IBC transfer between two chains
func TestBasicIBCTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:       "ufo-chain-1",
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
		ChainID:       "ufo-chain-2",
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

	// Wait for chains to produce at least one block
	time.Sleep(2 * time.Second)

	// Verify chains are running by checking the node status
	nodeStatus1, err := chain1.Client.GetNodeStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, nodeStatus1)

	nodeStatus2, err := chain2.Client.GetNodeStatus(ctx)
	require.NoError(t, err)
	require.NotNil(t, nodeStatus2)

	// Configure and start the Hermes relayer
	chainConfigs := []utils.TestConfig{chain1Config, chain2Config}
	err = utils.CreateHermesConfig(relayerDir, chainConfigs)
	require.NoError(t, err)

	// Start Hermes relayer
	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Wait for relayer to initialize
	time.Sleep(5 * time.Second)

	// Verify relayer status
	hermesStatus, err := utils.GetHermesStatus(ctx, relayerDir)
	require.NoError(t, err)
	t.Logf("Hermes status: %s", hermesStatus)

	// Create IBC channel between chains
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)
	t.Logf("Created IBC channel: %s (source) -> %s (destination)", sourceChannelID, destChannelID)

	// Get initial balances for comparison after transfer
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	// Create test accounts if they don't exist
	chain1Address, err := chain1Client.CreateKey(ctx, "test-sender")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-receiver")
	require.NoError(t, err)

	// Fund the test accounts
	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Verify initial balances
	initialSenderBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	require.NotZero(t, initialSenderBalance)
	t.Logf("Initial sender balance: %s stake", initialSenderBalance)

	initialReceiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, "stake")
	require.NoError(t, err)
	t.Logf("Initial receiver balance: %s stake", initialReceiverBalance)

	// Perform IBC transfer from chain1 to chain2
	transferAmount := "10000"
	transferHash, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("IBC transfer hash: %s", transferHash)

	// Wait for the transfer to complete
	time.Sleep(10 * time.Second)

	// Relay any pending packets
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)

	// Wait for packet relay to complete
	time.Sleep(5 * time.Second)

	// Verify balances after the transfer
	finalSenderBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	t.Logf("Final sender balance: %s stake", finalSenderBalance)

	// The IBC denom will be different on the destination chain
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This should be determined dynamically in a real implementation
	finalReceiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	t.Logf("Final receiver balance: %s %s", finalReceiverBalance, ibcDenom)

	// Verify that the transfer was successful
	// The sender's balance should have decreased
	initialSenderBalanceInt, _ := strconv.Atoi(initialSenderBalance)
	finalSenderBalanceInt, _ := strconv.Atoi(finalSenderBalance)
	transferAmountInt, _ := strconv.Atoi(transferAmount)
	assert.Less(t, finalSenderBalanceInt, initialSenderBalanceInt-transferAmountInt)

	// The receiver's balance in the IBC denominated token should be at least the transfer amount
	finalReceiverBalanceInt, _ := strconv.Atoi(finalReceiverBalance)
	assert.GreaterOrEqual(t, finalReceiverBalanceInt, transferAmountInt)
}

// TestBidirectionalIBCTransfer tests transferring tokens back and forth between two chains
func TestBidirectionalIBCTransfer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1-bidir")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2-bidir")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer-bidir")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:       "ufo-chain-1-bidir",
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
		ChainID:       "ufo-chain-2-bidir",
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

	// Start both chains and set up relayer following same steps as TestBasicIBCTransfer
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Configure and start the Hermes relayer
	chainConfigs := []utils.TestConfig{chain1Config, chain2Config}
	err = utils.CreateHermesConfig(relayerDir, chainConfigs)
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

	chain1Address, err := chain1Client.CreateKey(ctx, "test-account1")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-account2")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	err = chain2Client.FundAccount(ctx, chain2Address, "1000000stake")
	require.NoError(t, err)

	// === First Transfer: Chain1 -> Chain2 ===
	// Perform IBC transfer from chain1 to chain2
	transferAmount1 := "10000"
	transferHash1, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount1, "stake")
	require.NoError(t, err)
	t.Logf("First IBC transfer hash: %s", transferHash1)

	// Wait and relay packets
	time.Sleep(10 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Determine the IBC denom on chain2
	chain1Denom := "stake"
	chain2IBCDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This should be determined dynamically in real implementation

	// Verify chain2 received the tokens
	chain2Balance, err := chain2Client.GetBalance(ctx, chain2Address, chain2IBCDenom)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, chain2Balance, transferAmount1)
	t.Logf("Chain2 balance after first transfer: %s %s", chain2Balance, chain2IBCDenom)

	// === Second Transfer: Chain2 -> Chain1 (sending back the IBC tokens) ===
	transferAmount2 := "5000" // Sending back half
	transferHash2, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain2Config.ChainID, chain1Config.ChainID,
		destChannelID, chain2Address, chain1Address,
		transferAmount2, chain2IBCDenom)
	require.NoError(t, err)
	t.Logf("Second IBC transfer hash: %s", transferHash2)

	// Wait and relay packets
	time.Sleep(10 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain2Config.ChainID, chain1Config.ChainID,
		destChannelID, sourceChannelID)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Verify final balances
	// For chain1, the tokens should have returned to their original denomination
	chain1FinalBalance, err := chain1Client.GetBalance(ctx, chain1Address, chain1Denom)
	require.NoError(t, err)
	t.Logf("Chain1 final balance: %s %s", chain1FinalBalance, chain1Denom)

	// For chain2, the balance should be reduced by the amount sent back
	chain2FinalBalance, err := chain2Client.GetBalance(ctx, chain2Address, chain2IBCDenom)
	require.NoError(t, err)
	t.Logf("Chain2 final balance: %s %s", chain2FinalBalance, chain2IBCDenom)

	// Verify the amounts
	// Chain2 balance should be reduced by the amount sent back
	chain2BalanceInt, _ := strconv.Atoi(chain2Balance)
	chain2FinalBalanceInt, _ := strconv.Atoi(chain2FinalBalance)
	transferAmount2Int, _ := strconv.Atoi(transferAmount2)
	expectedChain2Balance := chain2BalanceInt - transferAmount2Int
	assert.GreaterOrEqual(t, chain2FinalBalanceInt, expectedChain2Balance)

	// The return transfer should have increased chain1's balance
	chain1FinalBalanceInt, _ := strconv.Atoi(chain1FinalBalance)
	assert.Greater(t, chain1FinalBalanceInt, 990000) // Initial balance minus first transfer plus returned amount
}

// TestIBCTimeouts tests that IBC transfers respect timeout parameters
func TestIBCTimeouts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories for each chain and the relayer
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1-timeout")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2-timeout")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer-timeout")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains with similar setup as before
	chain1Config := utils.TestConfig{
		ChainID:       "ufo-chain-1-timeout",
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
		ChainID:       "ufo-chain-2-timeout",
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

	// Start first chain
	chain1, err := utils.StartTestNode(ctx, chain1Config)
	require.NoError(t, err)
	defer chain1.Stop()

	// Start the relayer with only the first chain (to simulate receiver unavailability)
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config})
	require.NoError(t, err)

	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Now start the second chain
	chain2, err := utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Update Hermes config to include the second chain
	err = utils.CreateHermesConfig(relayerDir, []utils.TestConfig{chain1Config, chain2Config})
	require.NoError(t, err)

	// Restart Hermes with the updated config
	hermesProcess.Stop()
	hermesProcess, err = utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Create IBC channel with a very short timeout
	sourceChannelID, destChannelID, err := utils.CreateIBCChannel(ctx, relayerDir, chain1Config.ChainID, chain2Config.ChainID)
	require.NoError(t, err)

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-timeout-sender")
	require.NoError(t, err)

	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)
	chain2Address, err := chain2Client.CreateKey(ctx, "test-timeout-receiver")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Get initial balances
	initialSenderBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	t.Logf("Initial sender balance: %s stake", initialSenderBalance)

	// Stop chain2 to simulate receiver chain being down
	chain2.Stop()
	time.Sleep(2 * time.Second)

	// Make an IBC transfer with a very short timeout (1 block)
	transferAmount := "10000"
	// Custom TransferTokensIBC call with very short timeout
	cmd := utils.NewCustomCommand(ctx, "hermes",
		"tx", "ft-transfer",
		"--src-chain", chain1Config.ChainID,
		"--dst-chain", chain2Config.ChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
		"--amount", transferAmount,
		"--denom", "stake",
		"--timeout-height-offset", "1", // Very short timeout (1 block)
		"--timeout-seconds", "5", // Very short timeout (5 seconds)
		"--receiver", chain2Address,
	)
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "IBC transfer failed: %s", string(output))
	t.Logf("IBC transfer with timeout initiated: %s", string(output))

	// Wait for timeout
	time.Sleep(10 * time.Second)

	// Check sender's balance - should be refunded after timeout
	finalSenderBalance, err := chain1Client.GetBalance(ctx, chain1Address, "stake")
	require.NoError(t, err)
	t.Logf("Final sender balance: %s stake", finalSenderBalance)

	// The balance should be the same as initial or slightly less due to gas fees
	initialSenderBalanceInt, _ := strconv.Atoi(initialSenderBalance)
	finalSenderBalanceInt, _ := strconv.Atoi(finalSenderBalance)
	expectedBalance := initialSenderBalanceInt - 1000 // Accounting for gas fees
	assert.GreaterOrEqual(t, finalSenderBalanceInt, expectedBalance,
		"Balance should be refunded after timeout except for gas fees")

	// Restart chain2 to see if late acknowledgment causes issues
	chain2, err = utils.StartTestNode(ctx, chain2Config)
	require.NoError(t, err)
	defer chain2.Stop()

	// Relay packets to ensure any pending operations are processed
	time.Sleep(5 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)

	// Verify chain2 did not receive the tokens (the transfer timed out)
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This should be determined dynamically
	chain2Balance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	t.Logf("Chain2 balance of IBC tokens: %s", chain2Balance)
	assert.Equal(t, "0", chain2Balance, "Receiver should not have received any tokens due to timeout")
}

// TestCustomIBCChannelConfig tests creating IBC channels with custom configuration options
func TestCustomIBCChannelConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create temporary directories
	chain1Dir, err := os.MkdirTemp("", "ufo-ibc-chain1-custom")
	require.NoError(t, err)
	defer os.RemoveAll(chain1Dir)

	chain2Dir, err := os.MkdirTemp("", "ufo-ibc-chain2-custom")
	require.NoError(t, err)
	defer os.RemoveAll(chain2Dir)

	relayerDir, err := os.MkdirTemp("", "ufo-ibc-relayer-custom")
	require.NoError(t, err)
	defer os.RemoveAll(relayerDir)

	// Configure the chains
	chain1Config := utils.TestConfig{
		ChainID:        "ufo-chain-1-custom",
		RPCAddress:     "tcp://localhost:26657",
		GRPCAddress:    "localhost:9090",
		RESTAddress:    "localhost:1317",
		P2PAddress:     "localhost:26656",
		HomeDir:        chain1Dir,
		DebugLevel:     "info",
		KeysDir:        filepath.Join(chain1Dir, "keys"),
		BlockTime:      "1s",
		WithCometMock:  false,
		TrustingPeriod: "1h", // Short trusting period for testing
	}

	chain2Config := utils.TestConfig{
		ChainID:        "ufo-chain-2-custom",
		RPCAddress:     "tcp://localhost:26658",
		GRPCAddress:    "localhost:9091",
		RESTAddress:    "localhost:1318",
		P2PAddress:     "localhost:26659",
		HomeDir:        chain2Dir,
		DebugLevel:     "info",
		KeysDir:        filepath.Join(chain2Dir, "keys"),
		BlockTime:      "1s",
		WithCometMock:  false,
		TrustingPeriod: "1h", // Short trusting period for testing
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

	// Configure the Hermes relayer with custom options
	customOptions := map[string]interface{}{
		"clear_packets_interval": "10",
		"tx_confirmation":        true,
		"refresh_enabled":        true,
		"refresh_interval":       "20",
	}

	err = utils.CreateHermesConfigWithOptions(relayerDir, []utils.TestConfig{chain1Config, chain2Config}, customOptions)
	require.NoError(t, err)

	// Start Hermes relayer
	hermesProcess, err := utils.StartHermesRelayer(ctx, relayerDir)
	require.NoError(t, err)
	defer hermesProcess.Stop()

	// Wait for relayer to initialize
	time.Sleep(5 * time.Second)

	// Create IBC channel with custom parameters
	// For example, creating an ordered channel instead of the default unordered
	cmd := utils.NewCustomCommand(ctx, "hermes",
		"create", "channel",
		"--a-chain", chain1Config.ChainID,
		"--b-chain", chain2Config.ChainID,
		"--a-port", "transfer",
		"--b-port", "transfer",
		"--order", "ordered", // Using ordered channel instead of unordered
		"--new-client-connection", // Create new client and connection
	)
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Creating custom IBC channel failed: %s", string(output))
	t.Logf("Custom IBC channel creation output: %s", string(output))

	// Extract channel IDs (in a real implementation, parse the output)
	sourceChannelID := "channel-0" // Placeholder
	destChannelID := "channel-0"   // Placeholder

	// Create and fund test accounts
	chain1Client := utils.NewHTTPClient(chain1Config.RESTAddress)
	chain2Client := utils.NewHTTPClient(chain2Config.RESTAddress)

	chain1Address, err := chain1Client.CreateKey(ctx, "test-custom-sender")
	require.NoError(t, err)

	chain2Address, err := chain2Client.CreateKey(ctx, "test-custom-receiver")
	require.NoError(t, err)

	err = chain1Client.FundAccount(ctx, chain1Address, "1000000stake")
	require.NoError(t, err)

	// Verify that the ordered channel maintains the order of packets
	// Send multiple transfers in quick succession
	transferAmount := "1000"

	// Send first transfer
	transferHash1, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("First IBC transfer hash: %s", transferHash1)

	// Send second transfer immediately
	transferHash2, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("Second IBC transfer hash: %s", transferHash2)

	// Send third transfer immediately
	transferHash3, err := utils.TransferTokensIBC(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, chain1Address, chain2Address,
		transferAmount, "stake")
	require.NoError(t, err)
	t.Logf("Third IBC transfer hash: %s", transferHash3)

	// Wait and relay packets
	time.Sleep(10 * time.Second)
	err = utils.RelayPackets(ctx, relayerDir,
		chain1Config.ChainID, chain2Config.ChainID,
		sourceChannelID, destChannelID)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Verify receiver got all the tokens
	ibcDenom := fmt.Sprintf("ibc/%s", "hash_placeholder") // This should be determined dynamically
	receiverBalance, err := chain2Client.GetBalance(ctx, chain2Address, ibcDenom)
	require.NoError(t, err)
	transferAmountInt, _ := strconv.Atoi(transferAmount)
	expectedAmount := 3 * transferAmountInt
	expectedAmountStr := strconv.Itoa(expectedAmount)
	t.Logf("Receiver balance: %s %s (expected at least %s)", receiverBalance, ibcDenom, expectedAmountStr)

	// The receiver should have received all three transfers in order
	// For ordered channels, we expect all packets to be processed in order
	assert.GreaterOrEqual(t, receiverBalance, expectedAmount)
}

func NewCustomCommand(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd
}
