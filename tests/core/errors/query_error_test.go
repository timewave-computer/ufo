package errors

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

// TestQueryParameterValidationErrors tests that invalid query parameters result in appropriate errors.
func TestQueryParameterValidationErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Create HTTP client
	httpClient := utils.NewHTTPClient(node.Config.RESTAddress)

	// Wait for a block to be produced
	time.Sleep(5 * time.Second)

	// Get the latest block height for reference
	latestBlockResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest")
	require.NoError(t, err)
	require.NotNil(t, latestBlockResp["block"])

	latestHeight := int64(0)
	if blockHeader, ok := latestBlockResp["block"].(map[string]interface{})["header"].(map[string]interface{}); ok {
		latestHeight = int64(blockHeader["height"].(float64))
	}
	require.Greater(t, latestHeight, int64(0), "Failed to get latest block height")
	t.Logf("Latest block height: %d", latestHeight)

	// Test case 1: Invalid block height (negative)
	negativeHeightResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "height must be greater than 0", "Expected error for negative block height")
	t.Logf("Got expected error for negative block height: %v", err)

	// Test case 2: Invalid block height (too high)
	tooHighHeight := latestHeight + 1000
	tooHighHeightResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/"+strconv.FormatInt(tooHighHeight, 10))
	require.Error(t, err)
	require.Contains(t, err.Error(), "height", "Expected error for too high block height")
	t.Logf("Got expected error for too high block height: %v", err)

	// Test case 3: Malformed transaction hash query
	invalidTxHashResp, err := httpClient.Get(ctx, "/cosmos/tx/v1beta1/txs/invalid_hash")
	require.Error(t, err)
	require.Contains(t, err.Error(), "hash", "Expected error for invalid transaction hash")
	t.Logf("Got expected error for invalid transaction hash: %v", err)

	// Test case 4: Invalid pagination parameters
	invalidPaginationResp, err := httpClient.Get(ctx, "/cosmos/tx/v1beta1/txs?pagination.limit=-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "pagination", "Expected error for invalid pagination parameter")
	t.Logf("Got expected error for invalid pagination parameter: %v", err)

	// Test case 5: Invalid account address format
	invalidAddressResp, err := httpClient.Get(ctx, "/cosmos/auth/v1beta1/accounts/not_an_address")
	require.Error(t, err)
	require.Contains(t, err.Error(), "address", "Expected error for invalid address format")
	t.Logf("Got expected error for invalid address format: %v", err)

	// Test case 6: Invalid parameter type (string instead of number)
	invalidParamTypeResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/abc")
	require.Error(t, err)
	require.Contains(t, err.Error(), "height", "Expected error for invalid parameter type")
	t.Logf("Got expected error for invalid parameter type: %v", err)

	// Test case 7: Missing required parameter
	missingParamResp, err := httpClient.Post(ctx, "/cosmos/tx/v1beta1/simulate", map[string]interface{}{
		// Missing required "tx" field
		"tx_bytes": "",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing", "Expected error for missing required parameter")
	t.Logf("Got expected error for missing required parameter: %v", err)

	// Verify a valid query works
	validBlockResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err)
	require.NotNil(t, validBlockResp["block"])
	t.Logf("Valid block query succeeded")
}

// TestNonExistentDataQueries tests querying for non-existent data.
func TestNonExistentDataQueries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Create HTTP client
	httpClient := utils.NewHTTPClient(node.Config.RESTAddress)

	// Wait for a block to be produced
	time.Sleep(5 * time.Second)

	// Get the latest block height for reference
	latestBlockResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest")
	require.NoError(t, err)

	latestHeight := int64(0)
	if blockHeader, ok := latestBlockResp["block"].(map[string]interface{})["header"].(map[string]interface{}); ok {
		latestHeight = int64(blockHeader["height"].(float64))
	}
	require.Greater(t, latestHeight, int64(0), "Failed to get latest block height")
	t.Logf("Latest block height: %d", latestHeight)

	// Test case 1: Query for a non-existent block (future block)
	nonExistentBlockHeight := latestHeight + 100
	_, err = httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/"+strconv.FormatInt(nonExistentBlockHeight, 10))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found", "Expected not found error for non-existent block")
	t.Logf("Got expected error for non-existent block: %v", err)

	// Test case 2: Query for a non-existent transaction
	nonExistentTxHash := "0000000000000000000000000000000000000000000000000000000000000000"
	_, err = httpClient.Get(ctx, "/cosmos/tx/v1beta1/txs/"+nonExistentTxHash)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found", "Expected not found error for non-existent transaction")
	t.Logf("Got expected error for non-existent transaction: %v", err)

	// Test case 3: Query for a non-existent account
	// Generate a valid but unused address
	nonExistentKeyName := "non-existent-account-key"
	nonExistentAccount, err := node.CreateAccount(ctx, nonExistentKeyName)
	require.NoError(t, err)

	// Don't fund this account, so it doesn't exist on chain
	_, err = httpClient.Get(ctx, "/cosmos/bank/v1beta1/balances/"+nonExistentAccount.Address)
	// This might not return an error, just empty balances
	if err != nil {
		require.Contains(t, err.Error(), "not found", "Expected not found error for non-existent account")
		t.Logf("Got expected error for non-existent account: %v", err)
	} else {
		t.Logf("Non-existent account query returned empty balances instead of error")
	}

	// Test case 4: Query for a non-existent validator
	nonExistentValidatorAddr := "cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj"
	_, err = httpClient.Get(ctx, "/cosmos/staking/v1beta1/validators/"+nonExistentValidatorAddr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found", "Expected not found error for non-existent validator")
	t.Logf("Got expected error for non-existent validator: %v", err)

	// Test case 5: Query for a non-existent proposal
	_, err = httpClient.Get(ctx, "/cosmos/gov/v1beta1/proposals/99999")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found", "Expected not found error for non-existent proposal")
	t.Logf("Got expected error for non-existent proposal: %v", err)

	// Verify a valid query works
	validBlockResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err)
	require.NotNil(t, validBlockResp["block"])
	t.Logf("Valid block query succeeded")
}

// TestMalformedQueriesHandling tests handling of malformed queries.
func TestMalformedQueriesHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Create HTTP client
	httpClient := utils.NewHTTPClient(node.Config.RESTAddress)

	// Wait for a block to be produced
	time.Sleep(5 * time.Second)

	// Test case 1: Send malformed JSON in request body
	malformedJSON := []byte(`{"this is not valid json": "missing closing brace"`)
	err = httpClient.PostRaw(ctx, "/cosmos/tx/v1beta1/simulate", malformedJSON)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid", "Expected error for malformed JSON")
	t.Logf("Got expected error for malformed JSON: %v", err)

	// Test case 2: Send request with missing required parameters
	missingParamsResp, err := httpClient.Post(ctx, "/cosmos/tx/v1beta1/simulate", map[string]interface{}{
		// Missing the required tx_bytes field
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing", "Expected error for missing required parameters")
	t.Logf("Got expected error for missing required parameters: %v", err)

	// Test case 3: Send request with parameters of wrong type
	wrongTypeParams, err := httpClient.Post(ctx, "/cosmos/tx/v1beta1/simulate", map[string]interface{}{
		"tx_bytes": 12345, // Should be a string, not a number
	})
	require.Error(t, err)
	t.Logf("Got expected error for parameters of wrong type: %v", err)

	// Test case 4: Test invalid endpoint paths
	invalidPathResp, err := httpClient.Get(ctx, "/invalid/path/that/does/not/exist")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not", "Expected error for invalid endpoint path")
	t.Logf("Got expected error for invalid endpoint path: %v", err)

	// Test case 5: Test invalid HTTP method
	err = httpClient.PostRaw(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest", []byte(`{}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "method", "Expected error for invalid HTTP method")
	t.Logf("Got expected error for invalid HTTP method: %v", err)

	// Test case 6: Invalid query parameters
	invalidQueryParams, err := httpClient.Get(ctx, "/cosmos/tx/v1beta1/txs?events=invalid:event")
	require.Error(t, err)
	t.Logf("Got expected error for invalid query parameters: %v", err)

	// Verify a valid query works
	validQueryResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err)
	require.NotNil(t, validQueryResp["block"])
	t.Logf("Valid query succeeded")
}
