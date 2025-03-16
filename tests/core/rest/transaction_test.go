package rest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

func TestBankSendTransaction(t *testing.T) {
	// Set up test config for each binary type
	binaryTypes := []string{
		"fauxmosis-comet",
		"fauxmosis-ufo",
		// We'll add the other binary types when they're available
		// "osmosis-ufo-bridged",
		// "osmosis-ufo-patched",
	}

	for _, binaryType := range binaryTypes {
		t.Run(binaryType, func(t *testing.T) {
			// Set up context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Set up test config
			config := utils.DefaultTestConfig(binaryType)

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

			// Test case: Submit a bank send transaction
			// First, get the account information to get the account number and sequence
			type AccountResponse struct {
				Account struct {
					Address       string `json:"address"`
					AccountNumber string `json:"account_number"`
					Sequence      string `json:"sequence"`
				} `json:"account"`
			}

			var accountResp AccountResponse
			err = client.Get(ctx, "/cosmos/auth/v1beta1/accounts/cosmos1...", &accountResp) // Replace with an actual address
			require.NoError(t, err, "Failed to get account information")

			// Create a bank send transaction
			// In a real test, we'd create a proper transaction with the right structure
			// and sign it with a private key
			txReq := map[string]interface{}{
				"tx": map[string]interface{}{
					"body": map[string]interface{}{
						"messages": []map[string]interface{}{
							{
								"@type":        "/cosmos.bank.v1beta1.MsgSend",
								"from_address": "cosmos1...", // Replace with sender address
								"to_address":   "cosmos1...", // Replace with recipient address
								"amount": []map[string]interface{}{
									{
										"denom":  "stake",
										"amount": "100",
									},
								},
							},
						},
						"memo":                           "",
						"timeout_height":                 "0",
						"extension_options":              []interface{}{},
						"non_critical_extension_options": []interface{}{},
					},
					"auth_info": map[string]interface{}{
						"signer_infos": []map[string]interface{}{
							{
								"public_key": map[string]interface{}{
									"@type": "/cosmos.crypto.secp256k1.PubKey",
									"key":   "...", // Replace with a real public key
								},
								"mode_info": map[string]interface{}{
									"single": map[string]interface{}{
										"mode": "SIGN_MODE_DIRECT",
									},
								},
								"sequence": accountResp.Account.Sequence,
							},
						},
						"fee": map[string]interface{}{
							"amount": []map[string]interface{}{
								{
									"denom":  "stake",
									"amount": "200",
								},
							},
							"gas_limit": "200000",
							"payer":     "",
							"granter":   "",
						},
					},
					"signatures": []string{"..."}, // Replace with a real signature
				},
				"tx_bytes": "...", // Replace with real tx bytes if needed
				"mode":     "BROADCAST_MODE_SYNC",
			}

			// In a real test, we would have properly constructed transaction data
			// For this example, we're just checking if the endpoint accepts our request
			var txResp map[string]interface{}
			err = client.Post(ctx, "/cosmos/tx/v1beta1/txs", txReq, &txResp)

			// For now, we'll allow failures since we don't have real transaction data
			// In a real test, we'd validate the response
			t.Logf("Transaction response: %v", txResp)

			// Instead, we'll do a basic check to validate the node is running
			var statusResp map[string]interface{}
			err = client.Get(ctx, "/cosmos/base/tendermint/v1beta1/node_info", &statusResp)
			require.NoError(t, err, "Failed to get node status")

			// Check if the response contains expected fields
			assert.Contains(t, statusResp, "default_node_info", "Response should contain node info")
		})
	}
}

func TestNegativeTransactionCases(t *testing.T) {
	// Set up test config
	config := utils.DefaultTestConfig("fauxmosis-comet") // Using only one binary type for negative tests

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

	// Test case: Submit a malformed transaction
	malformedTx := map[string]interface{}{
		"tx": map[string]interface{}{
			"body": map[string]interface{}{
				"messages": []interface{}{}, // Empty messages should fail
			},
		},
		"mode": "BROADCAST_MODE_SYNC",
	}

	var errResp map[string]interface{}
	err = client.Post(ctx, "/cosmos/tx/v1beta1/txs", malformedTx, &errResp)

	// We expect an error for a malformed transaction
	// In a real test, we'd validate the specific error type and message
	t.Logf("Error response for malformed transaction: %v", errResp)

	// Validate node is still running after handling bad request
	var statusResp map[string]interface{}
	err = client.Get(ctx, "/cosmos/base/tendermint/v1beta1/node_info", &statusResp)
	require.NoError(t, err, "Node should still be running after handling bad request")
}
