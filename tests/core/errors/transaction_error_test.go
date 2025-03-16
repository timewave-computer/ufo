package errors

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

// TestTransactionValidationErrors tests various transaction validation errors.
func TestTransactionValidationErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Generate a new account/key pair
	keyName := "validation-test-key"
	accountInfo, err := node.CreateAccount(ctx, keyName)
	require.NoError(t, err)

	// Fund the account with some initial balance
	err = node.FundAccount(ctx, accountInfo.Address, "1000000stake")
	require.NoError(t, err)

	// Verify the account was funded
	account, err := node.GetAccount(ctx, accountInfo.Address)
	require.NoError(t, err)
	t.Logf("Account %s funded with balance: %v", accountInfo.Address, account.Balances)

	// Test cases for transaction validation errors
	testCases := []struct {
		name          string
		msg           map[string]interface{}
		txOptions     utils.TxOptions
		expectedError string
	}{
		{
			name: "Invalid memo (too long)",
			msg: map[string]interface{}{
				"@type":        "/cosmos.bank.v1beta1.MsgSend",
				"from_address": accountInfo.Address,
				"to_address":   accountInfo.Address,
				"amount": []map[string]interface{}{
					{
						"denom":  "stake",
						"amount": "1000",
					},
				},
			},
			txOptions: utils.TxOptions{
				// Create a memo that's too long (512+ characters)
				Memo: string(make([]byte, 513)),
				Gas:  200000,
				Fee:  "2000stake",
			},
			expectedError: "memo exceeds maximum size",
		},
		{
			name: "Invalid amount (zero)",
			msg: map[string]interface{}{
				"@type":        "/cosmos.bank.v1beta1.MsgSend",
				"from_address": accountInfo.Address,
				"to_address":   accountInfo.Address,
				"amount": []map[string]interface{}{
					{
						"denom":  "stake",
						"amount": "0",
					},
				},
			},
			txOptions: utils.TxOptions{
				Memo: "Testing zero amount",
				Gas:  200000,
				Fee:  "2000stake",
			},
			expectedError: "invalid coins",
		},
		{
			name: "Invalid recipient address",
			msg: map[string]interface{}{
				"@type":        "/cosmos.bank.v1beta1.MsgSend",
				"from_address": accountInfo.Address,
				"to_address":   "invalid-address",
				"amount": []map[string]interface{}{
					{
						"denom":  "stake",
						"amount": "1000",
					},
				},
			},
			txOptions: utils.TxOptions{
				Memo: "Testing invalid recipient",
				Gas:  200000,
				Fee:  "2000stake",
			},
			expectedError: "invalid address",
		},
		{
			name: "Missing required fields",
			msg: map[string]interface{}{
				"@type":        "/cosmos.bank.v1beta1.MsgSend",
				"from_address": accountInfo.Address,
				// Intentionally missing to_address
				"amount": []map[string]interface{}{
					{
						"denom":  "stake",
						"amount": "1000",
					},
				},
			},
			txOptions: utils.TxOptions{
				Memo: "Testing missing fields",
				Gas:  200000,
				Fee:  "2000stake",
			},
			expectedError: "invalid",
		},
		{
			name: "Non-existent token denom",
			msg: map[string]interface{}{
				"@type":        "/cosmos.bank.v1beta1.MsgSend",
				"from_address": accountInfo.Address,
				"to_address":   accountInfo.Address,
				"amount": []map[string]interface{}{
					{
						"denom":  "nonexistenttoken",
						"amount": "1000",
					},
				},
			},
			txOptions: utils.TxOptions{
				Memo: "Testing non-existent token",
				Gas:  200000,
				Fee:  "2000stake",
			},
			expectedError: "insufficient funds",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create and sign the transaction
			tx, err := node.CreateAndSignTx(
				ctx,
				keyName,
				[]map[string]interface{}{tc.msg},
				tc.txOptions,
			)

			// Some validation errors might occur during tx creation
			if err != nil {
				require.Contains(t, err.Error(), tc.expectedError,
					"Expected error containing '%s', got: %v", tc.expectedError, err)
				t.Logf("Got expected error during tx creation: %v", err)
				return
			}

			// Broadcast the transaction - should fail with the expected error
			_, err = node.BroadcastTx(ctx, tx, true)
			require.Error(t, err, "Expected transaction to fail with validation error")
			require.Contains(t, err.Error(), tc.expectedError,
				"Expected error containing '%s', got: %v", tc.expectedError, err)
			t.Logf("Got expected error during tx broadcast: %v", err)
		})
	}
}

// TestOutOfGasErrors tests scenarios where transactions run out of gas.
func TestOutOfGasErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Generate a new account/key pair
	keyName := "gas-test-key"
	accountInfo, err := node.CreateAccount(ctx, keyName)
	require.NoError(t, err)

	// Fund the account with some initial balance
	err = node.FundAccount(ctx, accountInfo.Address, "1000000stake")
	require.NoError(t, err)

	// Test case 1: Transaction with very low gas limit
	// First, get the gas estimate for a standard transaction
	msg := map[string]interface{}{
		"@type":        "/cosmos.bank.v1beta1.MsgSend",
		"from_address": accountInfo.Address,
		"to_address":   accountInfo.Address, // Send to self
		"amount": []map[string]interface{}{
			{
				"denom":  "stake",
				"amount": "1000",
			},
		},
	}

	// Create a standard transaction first to get a sense of the gas required
	stdTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{msg},
		utils.TxOptions{
			Memo: "Standard gas test",
			Gas:  200000,
			Fee:  "2000stake",
		},
	)
	require.NoError(t, err)

	// Simulate to get gas estimate
	simRes, err := node.SimulateTx(ctx, stdTx)
	require.NoError(t, err)
	estimatedGas := simRes.GasUsed
	t.Logf("Estimated gas for standard transaction: %d", estimatedGas)

	// Now create a transaction with insufficient gas (10% of estimated)
	lowGas := uint64(float64(estimatedGas) * 0.1)
	t.Logf("Using low gas limit: %d", lowGas)

	lowGasTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{msg},
		utils.TxOptions{
			Memo: "Low gas test",
			Gas:  lowGas,
			Fee:  "2000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction with low gas - should fail
	_, err = node.BroadcastTx(ctx, lowGasTx, true)
	require.Error(t, err, "Expected transaction with low gas to fail")
	require.Contains(t, err.Error(), "out of gas", "Expected out of gas error")
	t.Logf("Got expected out of gas error: %v", err)

	// Test case 2: Complex transaction with high computation but insufficient gas
	// Create a more complex transaction (multiple messages)
	multiMsg := []map[string]interface{}{}

	// Add multiple messages to increase complexity
	for i := 0; i < 10; i++ {
		multiMsg = append(multiMsg, map[string]interface{}{
			"@type":        "/cosmos.bank.v1beta1.MsgSend",
			"from_address": accountInfo.Address,
			"to_address":   accountInfo.Address,
			"amount": []map[string]interface{}{
				{
					"denom":  "stake",
					"amount": "100",
				},
			},
		})
	}

	// Create the complex transaction first to estimate gas
	complexTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		multiMsg,
		utils.TxOptions{
			Memo: "Complex transaction gas test",
			Gas:  500000, // Higher gas limit for simulation
			Fee:  "5000stake",
		},
	)
	require.NoError(t, err)

	// Simulate to get gas estimate
	complexSimRes, err := node.SimulateTx(ctx, complexTx)
	require.NoError(t, err)
	complexEstimatedGas := complexSimRes.GasUsed
	t.Logf("Estimated gas for complex transaction: %d", complexEstimatedGas)

	// Now create the same transaction with insufficient gas (30% of estimated)
	complexLowGas := uint64(float64(complexEstimatedGas) * 0.3)
	t.Logf("Using low gas limit for complex transaction: %d", complexLowGas)

	complexLowGasTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		multiMsg,
		utils.TxOptions{
			Memo: "Complex low gas test",
			Gas:  complexLowGas,
			Fee:  "5000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the complex transaction with low gas - should fail
	_, err = node.BroadcastTx(ctx, complexLowGasTx, true)
	require.Error(t, err, "Expected complex transaction with low gas to fail")
	require.Contains(t, err.Error(), "out of gas", "Expected out of gas error for complex transaction")
	t.Logf("Got expected out of gas error for complex transaction: %v", err)

	// Verify a transaction with sufficient gas succeeds
	sufficientGasTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{msg},
		utils.TxOptions{
			Memo: "Sufficient gas test",
			// Use estimated gas * 1.5 to ensure success
			Gas: uint64(float64(estimatedGas) * 1.5),
			Fee: "2000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction with sufficient gas - should succeed
	res, err := node.BroadcastTx(ctx, sufficientGasTx, true)
	require.NoError(t, err, "Expected transaction with sufficient gas to succeed")
	require.Equal(t, uint32(0), res.Code, "Expected successful transaction code")
	t.Logf("Transaction with sufficient gas succeeded with gas used: %d", res.GasUsed)
}

// TestInsufficientFundsErrors tests scenarios with insufficient funds.
func TestInsufficientFundsErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Generate a new account/key pair
	keyName := "funds-test-key"
	accountInfo, err := node.CreateAccount(ctx, keyName)
	require.NoError(t, err)

	// Fund the account with a small initial balance
	initialFunds := "10000stake"
	err = node.FundAccount(ctx, accountInfo.Address, initialFunds)
	require.NoError(t, err)

	// Verify the account was funded with the correct amount
	account, err := node.GetAccount(ctx, accountInfo.Address)
	require.NoError(t, err)
	t.Logf("Account %s funded with balance: %v", accountInfo.Address, account.Balances)

	// Test case 1: Sending more than the account balance
	// Extract the actual balance value for use in tests
	var availableBalance uint64
	for _, balance := range account.Balances {
		if balance.Denom == "stake" {
			availableBalance, err = strconv.ParseUint(balance.Amount, 10, 64)
			require.NoError(t, err)
			break
		}
	}
	require.NotEqual(t, uint64(0), availableBalance, "Failed to find stake balance")

	// Create a transaction attempting to send more than the available balance
	excessAmount := availableBalance + 1000
	msg := map[string]interface{}{
		"@type":        "/cosmos.bank.v1beta1.MsgSend",
		"from_address": accountInfo.Address,
		"to_address":   accountInfo.Address, // Send to self
		"amount": []map[string]interface{}{
			{
				"denom":  "stake",
				"amount": strconv.FormatUint(excessAmount, 10),
			},
		},
	}

	excessTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{msg},
		utils.TxOptions{
			Memo: "Excess amount test",
			Gas:  200000,
			Fee:  "2000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction - should fail with insufficient funds
	_, err = node.BroadcastTx(ctx, excessTx, true)
	require.Error(t, err, "Expected transaction with excess amount to fail")
	require.Contains(t, err.Error(), "insufficient funds", "Expected insufficient funds error")
	t.Logf("Got expected insufficient funds error: %v", err)

	// Test case 2: Sending exactly the account balance but not accounting for fees
	// Create a transaction with amount equal to balance (not leaving enough for fees)
	exactMsg := map[string]interface{}{
		"@type":        "/cosmos.bank.v1beta1.MsgSend",
		"from_address": accountInfo.Address,
		"to_address":   accountInfo.Address, // Send to self
		"amount": []map[string]interface{}{
			{
				"denom":  "stake",
				"amount": strconv.FormatUint(availableBalance, 10),
			},
		},
	}

	exactTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{exactMsg},
		utils.TxOptions{
			Memo: "Exact amount test",
			Gas:  200000,
			Fee:  "2000stake", // This fee can't be paid if sending entire balance
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction - should fail with insufficient funds for fees
	_, err = node.BroadcastTx(ctx, exactTx, true)
	require.Error(t, err, "Expected transaction with exact amount to fail due to fees")
	require.Contains(t, err.Error(), "insufficient funds", "Expected insufficient funds error")
	t.Logf("Got expected insufficient funds error for fees: %v", err)

	// Test case 3: Successfully sending with sufficient balance for amount + fees
	// Create a transaction with amount that leaves enough for fees
	sufficientAmount := availableBalance - 5000 // Leave 5000 for fees
	sufficientMsg := map[string]interface{}{
		"@type":        "/cosmos.bank.v1beta1.MsgSend",
		"from_address": accountInfo.Address,
		"to_address":   accountInfo.Address, // Send to self
		"amount": []map[string]interface{}{
			{
				"denom":  "stake",
				"amount": strconv.FormatUint(sufficientAmount, 10),
			},
		},
	}

	sufficientTx, err := node.CreateAndSignTx(
		ctx,
		keyName,
		[]map[string]interface{}{sufficientMsg},
		utils.TxOptions{
			Memo: "Sufficient amount test",
			Gas:  200000,
			Fee:  "2000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction - should succeed
	res, err := node.BroadcastTx(ctx, sufficientTx, true)
	require.NoError(t, err, "Expected transaction with sufficient funds to succeed")
	require.Equal(t, uint32(0), res.Code, "Expected successful transaction code")
	t.Logf("Transaction with sufficient funds succeeded")

	// Verify the balance was updated correctly
	updatedAccount, err := node.GetAccount(ctx, accountInfo.Address)
	require.NoError(t, err)
	t.Logf("Updated account balance: %v", updatedAccount.Balances)

	// Final test: attempt to send from an account with zero balance
	// Generate a new account without funding it
	emptyKeyName := "empty-funds-test-key"
	emptyAccountInfo, err := node.CreateAccount(ctx, emptyKeyName)
	require.NoError(t, err)

	// Create a transaction from the unfunded account
	emptyMsg := map[string]interface{}{
		"@type":        "/cosmos.bank.v1beta1.MsgSend",
		"from_address": emptyAccountInfo.Address,
		"to_address":   accountInfo.Address,
		"amount": []map[string]interface{}{
			{
				"denom":  "stake",
				"amount": "1000",
			},
		},
	}

	emptyTx, err := node.CreateAndSignTx(
		ctx,
		emptyKeyName,
		[]map[string]interface{}{emptyMsg},
		utils.TxOptions{
			Memo: "Empty account test",
			Gas:  200000,
			Fee:  "2000stake",
		},
	)
	require.NoError(t, err)

	// Broadcast the transaction from unfunded account - should fail
	_, err = node.BroadcastTx(ctx, emptyTx, true)
	require.Error(t, err, "Expected transaction from unfunded account to fail")
	require.Contains(t, err.Error(), "insufficient funds", "Expected insufficient funds error")
	t.Logf("Got expected insufficient funds error for unfunded account: %v", err)
}
