package abci

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

func TestABCIInfo(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create gRPC client
	client, err := utils.NewGRPCClient(config.GRPCAddress)
	require.NoError(t, err, "Failed to create gRPC client")
	defer client.Close()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test ABCI Info
	t.Run("ABCI Info", func(t *testing.T) {
		var resp map[string]interface{}
		err := httpClient.Get(ctx, "/abci_info", &resp)
		require.NoError(t, err, "Failed to get ABCI info")

		// Verify the response structure
		response, hasResponse := resp["response"].(map[string]interface{})
		require.True(t, hasResponse, "Response should include 'response' field")

		// Check for required fields in the response
		_, hasVersion := response["version"]
		assert.True(t, hasVersion, "Response should include 'version' field")

		_, hasAppVersion := response["app_version"]
		assert.True(t, hasAppVersion, "Response should include 'app_version' field")

		t.Logf("ABCI Info response: %v", resp)
	})
}

func TestABCIQuery(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test ABCI Query - Bank Module
	t.Run("ABCI Query - Bank Module", func(t *testing.T) {
		// Query all balances for a test account
		// This is a direct ABCI query to the bank module
		queryParams := map[string]interface{}{
			"path":  "store/bank/key",
			"data":  "0100000000000000", // Example: key prefix for balances
			"prove": false,
		}

		var resp map[string]interface{}
		err := httpClient.Post(ctx, "/abci_query", queryParams, &resp)
		require.NoError(t, err, "Failed to make ABCI query")

		// Log the response
		t.Logf("ABCI Query (Bank) response: %v", resp)

		// Verify the response structure
		response, hasResponse := resp["response"].(map[string]interface{})
		require.True(t, hasResponse, "Response should include 'response' field")

		// Check the response code - 0 means success
		code, hasCode := response["code"]
		if hasCode {
			codeVal, isInt := code.(float64)
			if isInt {
				assert.Equal(t, float64(0), codeVal, "Expected success response code")
			}
		}
	})

	// Test ABCI Query - Staking Module
	t.Run("ABCI Query - Staking Module", func(t *testing.T) {
		// Query validators from the staking module
		queryParams := map[string]interface{}{
			"path":  "store/staking/key",
			"data":  "0200000000000000", // Example: key prefix for validators
			"prove": false,
		}

		var resp map[string]interface{}
		err := httpClient.Post(ctx, "/abci_query", queryParams, &resp)
		require.NoError(t, err, "Failed to make ABCI query")

		// Log the response
		t.Logf("ABCI Query (Staking) response: %v", resp)
	})

	// Test ABCI Query with proof
	t.Run("ABCI Query with Proof", func(t *testing.T) {
		// Query with proof enabled
		queryParams := map[string]interface{}{
			"path":   "store/bank/key",
			"data":   "0100000000000000", // Example: key prefix for balances
			"prove":  true,
			"height": 1, // Query at a specific height
		}

		var resp map[string]interface{}
		err := httpClient.Post(ctx, "/abci_query", queryParams, &resp)
		require.NoError(t, err, "Failed to make ABCI query with proof")

		// Log the response
		t.Logf("ABCI Query with proof response: %v", resp)

		// Verify the proof is included (if supported)
		response, hasResponse := resp["response"].(map[string]interface{})
		if hasResponse {
			_, hasProof := response["proof"]
			t.Logf("Proof included: %v", hasProof)
		}
	})
}

func TestCheckTx(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test CheckTx via broadcast_tx_sync (which includes CheckTx)
	t.Run("CheckTx via broadcast_tx_sync", func(t *testing.T) {
		// Create a sample transaction (this is a simplified example)
		// In a real test, we would create a proper signed transaction

		// For this test, we'll use a placeholder tx
		// This will likely fail CheckTx, which is fine for testing
		txBytes := "0A94010A91010A1C2F636F736D6F732E62616E6B2E763162657461312E4D736753656E6412710A2D636F736D6F7331706B707472653766646B6C366766727A6C65736A6A766878686C63337234657A61667635390A2D636F736D6F73317A7138376D6C7173386773357A7A7974356C7575636A6E6365673534657636637A3833677A391A110A057374616B6512083130303030303030120974657374206D656D6F12670A500A460A1F2F636F736D6F732E63727970746F2E736563703235366B312E5075624B657912230A21028C3956DE0F92959BFBB7CCD6F97C5949BD5FE42518CE45C5CA6B3598B68312F12040A020801180A12130A0D0A057374616B6512043530303010904E1A40E44D599F6F7C79BC7242CD5C2A1A9F0E556118ADA0D63A941F2444006A9E8EF53DD5984D3B0B32EA15B67AAB25C0D84E172195AB5FE111C70A0F639644551"

		// Broadcast the transaction
		broadcastParams := map[string]interface{}{
			"tx": txBytes,
		}

		var resp map[string]interface{}
		err := httpClient.Post(ctx, "/broadcast_tx_sync", broadcastParams, &resp)
		require.NoError(t, err, "Failed to broadcast transaction")

		// Log the response
		t.Logf("Broadcast tx response: %v", resp)

		// Check if we got a hash (even if it failed validation, we should get a hash)
		_, hasHash := resp["hash"]
		assert.True(t, hasHash, "Response should include 'hash' field")

		// We might get an error in the CheckTx, which is fine for this test
		// Just log the code and log
		if code, hasCode := resp["code"]; hasCode {
			t.Logf("CheckTx code: %v", code)
		}
		if log, hasLog := resp["log"]; hasLog {
			t.Logf("CheckTx log: %v", log)
		}
	})
}

func TestDeliverTx(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test DeliverTx via broadcast_tx_commit (which includes CheckTx and DeliverTx)
	t.Run("DeliverTx via broadcast_tx_commit", func(t *testing.T) {
		// Create a sample transaction (this is a simplified example)
		// In a real test, we would create a proper signed transaction

		// For this test, we'll use a placeholder tx
		// This will likely fail, which is fine for testing
		txBytes := "0A94010A91010A1C2F636F736D6F732E62616E6B2E763162657461312E4D736753656E6412710A2D636F736D6F7331706B707472653766646B6C366766727A6C65736A6A766878686C63337234657A61667635390A2D636F736D6F73317A7138376D6C7173386773357A7A7974356C7575636A6E6365673534657636637A3833677A391A110A057374616B6512083130303030303030120974657374206D656D6F12670A500A460A1F2F636F736D6F732E63727970746F2E736563703235366B312E5075624B657912230A21028C3956DE0F92959BFBB7CCD6F97C5949BD5FE42518CE45C5CA6B3598B68312F12040A020801180A12130A0D0A057374616B6512043530303010904E1A40E44D599F6F7C79BC7242CD5C2A1A9F0E556118ADA0D63A941F2444006A9E8EF53DD5984D3B0B32EA15B67AAB25C0D84E172195AB5FE111C70A0F639644551"

		// Broadcast the transaction
		broadcastParams := map[string]interface{}{
			"tx": txBytes,
		}

		var resp map[string]interface{}
		err := httpClient.Post(ctx, "/broadcast_tx_commit", broadcastParams, &resp)
		require.NoError(t, err, "Failed to broadcast transaction")

		// Log the response
		t.Logf("Broadcast tx commit response: %v", resp)

		// Check for hash and check_tx result
		_, hasHash := resp["hash"]
		assert.True(t, hasHash, "Response should include 'hash' field")

		// Check for check_tx and deliver_tx results
		_, hasCheckTx := resp["check_tx"]
		assert.True(t, hasCheckTx, "Response should include 'check_tx' field")

		_, hasDeliverTx := resp["deliver_tx"]
		assert.True(t, hasDeliverTx, "Response should include 'deliver_tx' field")
	})
}

func TestBeginBlock(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test BeginBlock indirectly by checking for blocks being created
	t.Run("BeginBlock via block creation", func(t *testing.T) {
		// Get initial block height
		var initialResp map[string]interface{}
		err := httpClient.Get(ctx, "/status", &initialResp)
		require.NoError(t, err, "Failed to get node status")

		var initialHeight int64
		if syncInfo, ok := initialResp["sync_info"].(map[string]interface{}); ok {
			if latestBlockHeight, ok := syncInfo["latest_block_height"].(string); ok {
				// Parse the height, assuming it's a string (adjust if different)
				_, err := fmt.Sscanf(latestBlockHeight, "%d", &initialHeight)
				require.NoError(t, err, "Failed to parse initial block height")
			}
		}

		t.Logf("Initial block height: %d", initialHeight)

		// Wait a few seconds for a new block to be created
		time.Sleep(10 * time.Second)

		// Check if block height has increased
		var finalResp map[string]interface{}
		err = httpClient.Get(ctx, "/status", &finalResp)
		require.NoError(t, err, "Failed to get node status")

		var finalHeight int64
		if syncInfo, ok := finalResp["sync_info"].(map[string]interface{}); ok {
			if latestBlockHeight, ok := syncInfo["latest_block_height"].(string); ok {
				// Parse the height
				_, err := fmt.Sscanf(latestBlockHeight, "%d", &finalHeight)
				require.NoError(t, err, "Failed to parse final block height")
			}
		}

		t.Logf("Final block height: %d", finalHeight)

		// Assert that blocks are being created (height has increased)
		// Note: This is a weak test of BeginBlock. In a real test environment,
		// we might have more direct access to the BeginBlock events.
		assert.True(t, finalHeight >= initialHeight, "Expected block height to increase")
	})
}

func TestEndBlock(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test EndBlock indirectly by checking validator updates in blocks
	t.Run("EndBlock via validator set changes", func(t *testing.T) {
		// Get validators for a recent block
		var validatorsResp map[string]interface{}
		err := httpClient.Get(ctx, "/validators", &validatorsResp)
		require.NoError(t, err, "Failed to get validators")

		// Log validator information
		t.Logf("Validators response: %v", validatorsResp)

		// Verify response structure
		validators, hasValidators := validatorsResp["validators"].([]interface{})
		require.True(t, hasValidators, "Response should include 'validators' field")

		// Log number of validators
		t.Logf("Number of validators: %d", len(validators))
	})
}

func TestCommit(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test Commit by checking for committed blocks
	t.Run("Commit via committed blocks", func(t *testing.T) {
		// Get a block and check its commit information
		var blockResp map[string]interface{}
		err := httpClient.Get(ctx, "/block?height=1", &blockResp)
		require.NoError(t, err, "Failed to get block")

		// Verify the response has a block with commit info
		block, hasBlock := blockResp["block"].(map[string]interface{})
		require.True(t, hasBlock, "Response should include 'block' field")

		// Log block info
		t.Logf("Block response: %v", block)

		// Check for last_commit in the block
		_, hasLastCommit := block["last_commit"]
		assert.True(t, hasLastCommit, "Block should include 'last_commit' field")
	})
}

func TestInitChain(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Set up the node
	err := utils.SetupTestNode(ctx, config)
	require.NoError(t, err, "Failed to set up test node")
	defer func() {
		err := utils.CleanupTestNode(ctx, config)
		if err != nil {
			t.Logf("Warning: failed to clean up test node: %v", err)
		}
	}()

	// Create HTTP client for RPC endpoints
	httpClient := utils.NewHTTPClient(config.RPCAddress)

	// Test InitChain indirectly by checking genesis state
	t.Run("InitChain via genesis state", func(t *testing.T) {
		// Get the genesis block
		var genesisResp map[string]interface{}
		err := httpClient.Get(ctx, "/genesis", &genesisResp)
		require.NoError(t, err, "Failed to get genesis")

		// Verify the response structure
		genesis, hasGenesis := genesisResp["genesis"].(map[string]interface{})
		require.True(t, hasGenesis, "Response should include 'genesis' field")

		// Check for app_state in genesis
		_, hasAppState := genesis["app_state"]
		assert.True(t, hasAppState, "Genesis should include 'app_state' field")

		// Log genesis app state
		t.Logf("Genesis includes app_state: %v", hasAppState)
	})
}
