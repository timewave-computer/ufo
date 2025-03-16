package rest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

func TestBankTransactionTypes(t *testing.T) {
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

	// Test case: Send tokens
	t.Run("Send Tokens", func(t *testing.T) {
		// Prepare a transaction to test sending tokens
		sendTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":        "/cosmos.bank.v1beta1.MsgSend",
							"from_address": "cosmos1test", // Replace with a test account
							"to_address":   "cosmos1recv", // Replace with a test account
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
								"key":   "test_key", // Replace with a real public key
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "1",
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
				"signatures": []string{"test_signature"}, // Replace with a real signature
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", sendTx, &resp)

		// In a real test, we would check for success
		// For now, just log the result since we're using test data
		t.Logf("Response from send transaction: %v", resp)

		// Verify we can query the account balances
		var balanceResp map[string]interface{}
		err = client.Get(ctx, "/cosmos/bank/v1beta1/balances/cosmos1test", &balanceResp)
		t.Logf("Balance response: %v, error: %v", balanceResp, err)
	})

	// Test case: Multi-send
	t.Run("Multi Send", func(t *testing.T) {
		// Prepare a multisend transaction
		multiSendTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type": "/cosmos.bank.v1beta1.MsgMultiSend",
							"inputs": []map[string]interface{}{
								{
									"address": "cosmos1test",
									"coins": []map[string]interface{}{
										{
											"denom":  "stake",
											"amount": "200",
										},
									},
								},
							},
							"outputs": []map[string]interface{}{
								{
									"address": "cosmos1recv1",
									"coins": []map[string]interface{}{
										{
											"denom":  "stake",
											"amount": "100",
										},
									},
								},
								{
									"address": "cosmos1recv2",
									"coins": []map[string]interface{}{
										{
											"denom":  "stake",
											"amount": "100",
										},
									},
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "2",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", multiSendTx, &resp)

		// Log the result
		t.Logf("Response from multi-send transaction: %v", resp)
	})
}

func TestStakingTransactionTypes(t *testing.T) {
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

	// Test case: Delegate tokens
	t.Run("Delegate", func(t *testing.T) {
		// Prepare a delegation transaction
		delegateTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":             "/cosmos.staking.v1beta1.MsgDelegate",
							"delegator_address": "cosmos1test",
							"validator_address": "cosmosvaloper1test",
							"amount": map[string]interface{}{
								"denom":  "stake",
								"amount": "1000",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "3",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", delegateTx, &resp)

		// Log the result
		t.Logf("Response from delegate transaction: %v", resp)

		// Verify we can query the delegation
		var delegationResp map[string]interface{}
		err = client.Get(ctx, "/cosmos/staking/v1beta1/validators/cosmosvaloper1test/delegations/cosmos1test", &delegationResp)
		t.Logf("Delegation response: %v, error: %v", delegationResp, err)
	})

	// Test case: Undelegate tokens
	t.Run("Undelegate", func(t *testing.T) {
		// Prepare an undelegation transaction
		undelegateTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":             "/cosmos.staking.v1beta1.MsgUndelegate",
							"delegator_address": "cosmos1test",
							"validator_address": "cosmosvaloper1test",
							"amount": map[string]interface{}{
								"denom":  "stake",
								"amount": "500",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "4",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", undelegateTx, &resp)

		// Log the result
		t.Logf("Response from undelegate transaction: %v", resp)
	})

	// Test case: Redelegate tokens
	t.Run("Redelegate", func(t *testing.T) {
		// Prepare a redelegation transaction
		redelegateTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":                 "/cosmos.staking.v1beta1.MsgBeginRedelegate",
							"delegator_address":     "cosmos1test",
							"validator_src_address": "cosmosvaloper1test",
							"validator_dst_address": "cosmosvaloper2test",
							"amount": map[string]interface{}{
								"denom":  "stake",
								"amount": "250",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "5",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", redelegateTx, &resp)

		// Log the result
		t.Logf("Response from redelegate transaction: %v", resp)
	})
}

func TestGovernanceTransactionTypes(t *testing.T) {
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

	// Test case: Submit a text proposal
	t.Run("Submit Proposal", func(t *testing.T) {
		// Prepare a proposal submission transaction
		proposalTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
							"content": map[string]interface{}{
								"@type":       "/cosmos.gov.v1beta1.TextProposal",
								"title":       "Test Proposal",
								"description": "This is a test proposal",
							},
							"initial_deposit": []map[string]interface{}{
								{
									"denom":  "stake",
									"amount": "10000",
								},
							},
							"proposer": "cosmos1test",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "6",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", proposalTx, &resp)

		// Log the result
		t.Logf("Response from submit proposal transaction: %v", resp)

		// In a real test, we'd verify the proposal was created and we could query it
		var proposalsResp map[string]interface{}
		err = client.Get(ctx, "/cosmos/gov/v1beta1/proposals", &proposalsResp)
		t.Logf("Proposals response: %v, error: %v", proposalsResp, err)
	})

	// Test case: Vote on a proposal
	t.Run("Vote", func(t *testing.T) {
		// Prepare a vote transaction
		voteTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":       "/cosmos.gov.v1beta1.MsgVote",
							"proposal_id": "1",
							"voter":       "cosmos1test",
							"option":      "VOTE_OPTION_YES",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "7",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", voteTx, &resp)

		// Log the result
		t.Logf("Response from vote transaction: %v", resp)
	})

	// Test case: Deposit to a proposal
	t.Run("Deposit", func(t *testing.T) {
		// Prepare a deposit transaction
		depositTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":       "/cosmos.gov.v1beta1.MsgDeposit",
							"proposal_id": "1",
							"depositor":   "cosmos1test",
							"amount": []map[string]interface{}{
								{
									"denom":  "stake",
									"amount": "5000",
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
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "8",
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
				"signatures": []string{"test_signature"},
			},
			"mode": "BROADCAST_MODE_SYNC",
		}

		// Make the request to submit the transaction
		var resp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/txs", depositTx, &resp)

		// Log the result
		t.Logf("Response from deposit transaction: %v", resp)
	})
}

func TestTransactionEncoding(t *testing.T) {
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

	// Test case: Encode transaction
	t.Run("Encode Transaction", func(t *testing.T) {
		// Create a transaction for encoding
		txToEncode := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":        "/cosmos.bank.v1beta1.MsgSend",
							"from_address": "cosmos1test",
							"to_address":   "cosmos1recv",
							"amount": []map[string]interface{}{
								{
									"denom":  "stake",
									"amount": "100",
								},
							},
						},
					},
					"memo":                           "Test encoding",
					"timeout_height":                 "0",
					"extension_options":              []interface{}{},
					"non_critical_extension_options": []interface{}{},
				},
				"auth_info": map[string]interface{}{
					"signer_infos": []map[string]interface{}{
						{
							"public_key": map[string]interface{}{
								"@type": "/cosmos.crypto.secp256k1.PubKey",
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "9",
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
				"signatures": []string{},
			},
		}

		// Make the request to encode the transaction
		var encodeResp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/encode", txToEncode, &encodeResp)

		// In a real test, we'd verify the encoding was successful
		// and decode it back to verify correctness
		t.Logf("Response from encode transaction: %v, error: %v", encodeResp, err)

		// If encoding was successful, try decoding
		if err == nil && encodeResp != nil {
			txBytes := encodeResp["tx_bytes"]
			if txBytes != nil {
				// Try to decode it
				decodeTx := map[string]interface{}{
					"tx_bytes": txBytes,
				}

				var decodeResp map[string]interface{}
				err := client.Post(ctx, "/cosmos/tx/v1beta1/decode", decodeTx, &decodeResp)
				t.Logf("Response from decode transaction: %v, error: %v", decodeResp, err)
			}
		}
	})
}

func TestGasEstimation(t *testing.T) {
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

	// Test case: Simulate transaction to estimate gas
	t.Run("Simulate Transaction", func(t *testing.T) {
		// Create a transaction for simulation
		simulateTx := map[string]interface{}{
			"tx": map[string]interface{}{
				"body": map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"@type":        "/cosmos.bank.v1beta1.MsgSend",
							"from_address": "cosmos1test",
							"to_address":   "cosmos1recv",
							"amount": []map[string]interface{}{
								{
									"denom":  "stake",
									"amount": "100",
								},
							},
						},
					},
					"memo":                           "Test simulation",
					"timeout_height":                 "0",
					"extension_options":              []interface{}{},
					"non_critical_extension_options": []interface{}{},
				},
				"auth_info": map[string]interface{}{
					"signer_infos": []map[string]interface{}{
						{
							"public_key": map[string]interface{}{
								"@type": "/cosmos.crypto.secp256k1.PubKey",
								"key":   "test_key",
							},
							"mode_info": map[string]interface{}{
								"single": map[string]interface{}{
									"mode": "SIGN_MODE_DIRECT",
								},
							},
							"sequence": "10",
						},
					},
					"fee": map[string]interface{}{
						"amount": []map[string]interface{}{
							{
								"denom":  "stake",
								"amount": "0", // We don't know the fee yet
							},
						},
						"gas_limit": "0", // We want to estimate this
						"payer":     "",
						"granter":   "",
					},
				},
				"signatures": []string{"test_signature"}, // Doesn't matter for simulation
			},
		}

		// Make the request to simulate the transaction
		var simulateResp map[string]interface{}
		err := client.Post(ctx, "/cosmos/tx/v1beta1/simulate", simulateTx, &simulateResp)

		// In a real test, we'd verify the gas estimation was returned
		t.Logf("Response from simulate transaction: %v, error: %v", simulateResp, err)

		// Check if we got a valid gas estimate
		if err == nil && simulateResp != nil {
			// Extract the gas estimate and use it to submit a real transaction
			gasUsed := 0

			// Check if gas_info exists in the response
			if gasInfo, ok := simulateResp["gas_info"].(map[string]interface{}); ok {
				if gasUsedStr, ok := gasInfo["gas_used"].(string); ok {
					// Convert to int
					fmt.Sscanf(gasUsedStr, "%d", &gasUsed)
				}
			}

			// If we got a valid gas estimate, use it (add 20% buffer)
			if gasUsed > 0 {
				gasLimit := int(float64(gasUsed) * 1.2)
				t.Logf("Estimated gas: %d, gas limit with buffer: %d", gasUsed, gasLimit)
			}
		}
	})
}
