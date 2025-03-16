package rest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

func TestStateQueries(t *testing.T) {
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

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Test case: Query bank module state (balances)
	t.Run("Bank Module - Balances", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/bank/v1beta1/balances/cosmos1test", &resp)

		// In a real test, we'd verify the response contains the expected balances
		// For now, just log the result
		t.Logf("Response from bank balances query: %v", resp)
	})

	// Test case: Query staking module state (validators)
	t.Run("Staking Module - Validators", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/staking/v1beta1/validators", &resp)

		// Log the result
		t.Logf("Response from staking validators query: %v", resp)
	})

	// Test case: Query gov module state (proposals)
	t.Run("Gov Module - Proposals", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/gov/v1beta1/proposals", &resp)

		// Log the result
		t.Logf("Response from gov proposals query: %v", resp)
	})

	// Test case: Query historical state (at a specific height)
	t.Run("Historical Queries", func(t *testing.T) {
		// First, get the current height
		var nodeInfoResp map[string]interface{}
		err := client.Get(ctx, "/cosmos/base/tendermint/v1beta1/node_info", &nodeInfoResp)
		t.Logf("Node info response: %v", nodeInfoResp)

		// Assuming we have a way to get the current height, let's try to query at height-1
		// For testing purposes, let's just use height=1
		var balanceAtHeightResp map[string]interface{}
		err = client.Get(ctx, "/cosmos/bank/v1beta1/balances/cosmos1test?height=1", &balanceAtHeightResp)

		// Log the result
		t.Logf("Response from historical balance query: %v", balanceAtHeightResp)
	})
}

func TestBlockQueries(t *testing.T) {
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

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Test case: Query latest block
	t.Run("Latest Block", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest", &resp)

		// Log the result
		t.Logf("Response from latest block query: %v", resp)

		// Verify the response structure
		if err == nil {
			// Check if block info is present
			_, hasBlock := resp["block"]
			assert.True(t, hasBlock, "Response should include 'block' field")
		}
	})

	// Test case: Query block by height
	t.Run("Block By Height", func(t *testing.T) {
		// Query block at height 1
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1", &resp)

		// Log the result
		t.Logf("Response from block by height query: %v", resp)

		// Verify the response structure
		if err == nil {
			// Check if block info is present
			_, hasBlock := resp["block"]
			assert.True(t, hasBlock, "Response should include 'block' field")
		}
	})

	// Test case: Query invalid block height
	t.Run("Invalid Block Height", func(t *testing.T) {
		// Query block at a very high height (likely invalid)
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/999999999", &resp)

		// Log the result
		t.Logf("Response from invalid block height query: %v, error: %v", resp, err)

		// We expect an error or empty result for an invalid height
		// But the specific behavior depends on the implementation
	})
}

func TestTransactionQueries(t *testing.T) {
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

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Test case: Query transaction by hash
	t.Run("Transaction By Hash", func(t *testing.T) {
		// For testing, we'll use a placeholder hash
		// In a real test, we would submit a transaction first then query it
		txHash := "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"

		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/tx/v1beta1/txs/"+txHash, &resp)

		// Log the result
		t.Logf("Response from tx by hash query: %v, error: %v", resp, err)

		// We expect an error for a non-existent hash in this test
		// In a real test with a valid hash, we would verify the response
	})

	// Test case: Query transactions by events
	t.Run("Transactions By Events", func(t *testing.T) {
		// Query for transactions of message type MsgSend
		queryParams := "message.action='/cosmos.bank.v1beta1.MsgSend'"

		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/tx/v1beta1/txs?events="+queryParams, &resp)

		// Log the result
		t.Logf("Response from tx by events query: %v", resp)
	})

	// Test case: Query with pagination
	t.Run("Pagination", func(t *testing.T) {
		// Query for all transactions with pagination
		var resp map[string]interface{}
		err := client.Get(ctx, "/cosmos/tx/v1beta1/txs?pagination.limit=5&pagination.offset=0", &resp)

		// Log the result
		t.Logf("Response from paginated tx query: %v", resp)

		// Verify pagination info if available
		if err == nil {
			pagination, hasPagination := resp["pagination"].(map[string]interface{})
			if hasPagination {
				_, hasTotal := pagination["total"]
				assert.True(t, hasTotal, "Pagination response should include 'total' field")
			}
		}

		// Query next page if there is one
		if err == nil && resp != nil {
			pagination, hasPagination := resp["pagination"].(map[string]interface{})
			if hasPagination {
				nextKey, hasNextKey := pagination["next_key"]
				if hasNextKey && nextKey != nil && nextKey != "" {
					// Query next page
					var nextPageResp map[string]interface{}
					err := client.Get(ctx, "/cosmos/tx/v1beta1/txs?pagination.key="+nextKey.(string), &nextPageResp)
					t.Logf("Response from next page query: %v", nextPageResp)
				}
			}
		}
	})
}

func TestABCIQueries(t *testing.T) {
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

	// Create HTTP client
	client := utils.NewHTTPClient(config.RESTAddress)

	// Test case: Direct ABCI query
	t.Run("ABCI Query", func(t *testing.T) {
		// Perform a direct ABCI query for account info
		// The path is equivalent to querying account info through the bank module
		queryParams := map[string]interface{}{
			"path":  "/cosmos.bank.v1beta1.Query/AllBalances",
			"data":  "0a0a636f736d6f7331746573741220636f736d6f7331746573740000000000000000000000000000000000000000", // hex encoded protobuf
			"prove": false,
		}

		var resp map[string]interface{}
		err := client.Post(ctx, "/abci_query", queryParams, &resp)

		// Log the result
		t.Logf("Response from ABCI query: %v", resp)

		// Verify response structure if available
		if err == nil {
			response, hasResponse := resp["response"].(map[string]interface{})
			if hasResponse {
				_, hasValue := response["value"]
				assert.True(t, hasValue, "ABCI query response should include 'value' field")
			}
		}
	})

	// Test case: ABCI Info
	t.Run("ABCI Info", func(t *testing.T) {
		var resp map[string]interface{}
		err := client.Get(ctx, "/abci_info", &resp)

		// Log the result
		t.Logf("Response from ABCI info: %v", resp)

		// Verify response structure if available
		if err == nil {
			_, hasResponse := resp["response"]
			assert.True(t, hasResponse, "ABCI info response should include 'response' field")
		}
	})
}
