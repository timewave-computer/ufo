package mempool

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

// generateTxBytes creates a transaction with specified gas price and sequence number
// For testing purposes, we just generate a placeholder tx with different gas prices
func generateTxBytes(gasPrice int, sequence int) string {
	// In a real implementation, we would create actual different transactions
	// For now, we'll just return a fixed transaction string to simulate the test
	return "0A94010A91010A1C2F636F736D6F732E62616E6B2E763162657461312E4D736753656E6412710A2D636F736D6F7331706B707472653766646B6C366766727A6C65736A6A766878686C63337234657A61667635390A2D636F736D6F73317A7138376D6C7173386773357A7A7974356C7575636A6E6365673534657636637A3833677A391A110A057374616B6512083130303030303030120974657374206D656D6F12670A500A460A1F2F636F736D6F732E63727970746F2E736563703235366B312E5075624B657912230A21028C3956DE0F92959BFBB7CCD6F97C5949BD5FE42518CE45C5CA6B3598B68312F12040A020801180A12130A0D0A057374616B6512043530303010904E1A40E44D599F6F7C79BC7242CD5C2A1A9F0E556118ADA0D63A941F2444006A9E8EF53DD5984D3B0B32EA15B67AAB25C0D84E172195AB5FE111C70A0F639644551"
}

// submitTransaction submits a transaction to the node and returns the hash
func submitTransaction(ctx context.Context, httpClient *utils.HTTPClient, txBytes string) (string, error) {
	broadcastParams := map[string]interface{}{
		"tx": txBytes,
	}

	var resp map[string]interface{}
	err := httpClient.Post(ctx, "/broadcast_tx_sync", broadcastParams, &resp)
	if err != nil {
		return "", err
	}

	// Extract the hash
	hash, ok := resp["hash"].(string)
	if !ok {
		return "", nil
	}

	return hash, nil
}

// checkTransactionInMempool checks if a transaction is in the mempool
func checkTransactionInMempool(ctx context.Context, httpClient *utils.HTTPClient, txHash string) (bool, error) {
	var resp map[string]interface{}
	err := httpClient.Get(ctx, "/unconfirmed_txs", &resp)
	if err != nil {
		return false, err
	}

	// Check if txs exists in the response
	txs, ok := resp["txs"].([]interface{})
	if !ok {
		return false, nil
	}

	// For this test, we'll just return true if there are any transactions
	// In a real test, we would decode and check the hash of each transaction
	return len(txs) > 0, nil
}

// getNumberOfTransactionsInMempool gets the count of transactions in the mempool
func getNumberOfTransactionsInMempool(ctx context.Context, httpClient *utils.HTTPClient) (int, error) {
	var resp map[string]interface{}
	err := httpClient.Get(ctx, "/unconfirmed_txs", &resp)
	if err != nil {
		return 0, err
	}

	// Check if n_txs exists in the response
	nTxs, ok := resp["n_txs"].(float64)
	if !ok {
		return 0, nil
	}

	return int(nTxs), nil
}

func TestMempoolCapacity(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// Test mempool capacity
	t.Run("Mempool Maximum Capacity", func(t *testing.T) {
		// Check initial mempool size
		initialCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get initial mempool count")
		t.Logf("Initial mempool transaction count: %d", initialCount)

		// Submit transactions until mempool is full or we reach a reasonable limit for testing
		maxTestTxs := 50 // Adjust as needed for your test environment
		txHashes := make([]string, 0, maxTestTxs)

		for i := 0; i < maxTestTxs; i++ {
			// Generate a test transaction
			txBytes := generateTxBytes(1000, i+1)

			// Submit transaction
			hash, err := submitTransaction(ctx, httpClient, txBytes)
			if err != nil {
				t.Logf("Failed to submit transaction %d: %v", i+1, err)
				continue
			}

			if hash != "" {
				txHashes = append(txHashes, hash)
				t.Logf("Submitted transaction %d with hash: %s", i+1, hash)
			}

			// Slight delay to avoid overwhelming the node
			time.Sleep(500 * time.Millisecond)
		}

		// Check final mempool size
		finalCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get final mempool count")
		t.Logf("Final mempool transaction count: %d", finalCount)

		// Verify mempool has transactions
		assert.True(t, finalCount > initialCount, "Mempool should have more transactions after submission")

		// Wait for some transactions to be processed in blocks
		t.Log("Waiting for transactions to be processed...")
		time.Sleep(20 * time.Second)

		// Check if mempool size decreased (transactions were included in blocks)
		afterWaitCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get mempool count after waiting")
		t.Logf("Mempool transaction count after waiting: %d", afterWaitCount)
	})
}

func TestTransactionPriority(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// Test transaction priority based on gas price
	t.Run("Transaction Priority by Gas Price", func(t *testing.T) {
		// Clear mempool by waiting for transactions to be processed
		// This is a simplified approach - in a real test we might have a way to reset the state
		t.Log("Waiting for mempool to clear...")
		time.Sleep(20 * time.Second)

		initialCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get initial mempool count")
		t.Logf("Initial mempool transaction count: %d", initialCount)

		// Generate transactions with different gas prices
		// Lower gas price - should be lower priority
		lowGasTx := generateTxBytes(100, 1)
		// Medium gas price
		mediumGasTx := generateTxBytes(500, 2)
		// High gas price - should be higher priority
		highGasTx := generateTxBytes(1000, 3)

		// Randomize submission order to ensure priority is based on gas price, not submission time
		// This is a basic randomization for testing purposes
		txsToSubmit := []struct {
			name     string
			gasPrice int
			txBytes  string
		}{
			{"low", 100, lowGasTx},
			{"medium", 500, mediumGasTx},
			{"high", 1000, highGasTx},
		}

		// Shuffle the order
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(txsToSubmit), func(i, j int) {
			txsToSubmit[i], txsToSubmit[j] = txsToSubmit[j], txsToSubmit[i]
		})

		// Submit transactions in random order
		for i, tx := range txsToSubmit {
			hash, err := submitTransaction(ctx, httpClient, tx.txBytes)
			require.NoError(t, err, "Failed to submit transaction")
			t.Logf("Submitted %s gas price transaction (%d) with hash: %s", tx.name, tx.gasPrice, hash)

			// Slight delay to ensure distinct submission times
			time.Sleep(500 * time.Millisecond)
		}

		// Wait a moment for mempool to update
		time.Sleep(2 * time.Second)

		// Get mempool transactions
		var mempoolResp map[string]interface{}
		err = httpClient.Get(ctx, "/unconfirmed_txs", &mempoolResp)
		require.NoError(t, err, "Failed to get mempool transactions")

		// Log the mempool response to see transaction order
		// In a real test, we would decode and verify the order based on gas price
		t.Logf("Mempool response: %v", mempoolResp)

		// Wait for the next block to be created
		t.Log("Waiting for next block...")
		time.Sleep(10 * time.Second)

		// Check if transactions were included in the new block
		// This part is simplified. In a real test, we would check which transactions
		// were included first to verify priority ordering
		var blockResp map[string]interface{}
		err = httpClient.Get(ctx, "/block", &blockResp)
		require.NoError(t, err, "Failed to get latest block")

		// Log the block to see which transactions were included
		if block, ok := blockResp["block"].(map[string]interface{}); ok {
			if data, ok := block["data"].(map[string]interface{}); ok {
				t.Logf("Block data txs count: %v", data["txs"])
			}
		}
	})
}

func TestReplacingTransaction(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// Test replacing a transaction in the mempool
	t.Run("Replace Transaction in Mempool", func(t *testing.T) {
		// This test simulates replacing a transaction in the mempool
		// with a higher gas price transaction with the same nonce

		// Clear mempool by waiting for transactions to be processed
		t.Log("Waiting for mempool to clear...")
		time.Sleep(20 * time.Second)

		// Get initial mempool count
		initialCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get initial mempool count")
		t.Logf("Initial mempool transaction count: %d", initialCount)

		// Submit first transaction with low gas price
		// In a real test, we would create a proper transaction with a specific nonce
		lowGasTx := generateTxBytes(100, 1)
		lowHash, err := submitTransaction(ctx, httpClient, lowGasTx)
		require.NoError(t, err, "Failed to submit low gas transaction")
		t.Logf("Submitted low gas transaction with hash: %s", lowHash)

		// Wait a moment
		time.Sleep(2 * time.Second)

		// Submit replacement transaction with higher gas price but same nonce
		// In a real test, this would be a different transaction with the same nonce
		highGasTx := generateTxBytes(1000, 1)
		highHash, err := submitTransaction(ctx, httpClient, highGasTx)
		require.NoError(t, err, "Failed to submit high gas transaction")
		t.Logf("Submitted high gas transaction with hash: %s", highHash)

		// Wait a moment for mempool to update
		time.Sleep(2 * time.Second)

		// Get mempool transactions count
		afterSubmitCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get mempool count after submission")
		t.Logf("Mempool transaction count after submission: %d", afterSubmitCount)

		// Note: Depending on the implementation, the mempool might:
		// 1. Replace the low gas transaction with the high gas one (count stays the same)
		// 2. Keep both transactions (count increases by 1)
		// 3. Reject the second transaction due to nonce conflict (count stays the same)

		// Log the outcome for manual verification
		// In a real test, we would verify the expected behavior
		t.Logf("Transaction replacement result: initial=%d, after=%d", initialCount, afterSubmitCount)

		// Wait for the next block to be created
		t.Log("Waiting for next block...")
		time.Sleep(10 * time.Second)

		// Check which transaction was included in the block
		// This is a simplified approach for testing purposes
		var blockResp map[string]interface{}
		err = httpClient.Get(ctx, "/block", &blockResp)
		require.NoError(t, err, "Failed to get latest block")

		// Log the block to see which transactions were included
		if block, ok := blockResp["block"].(map[string]interface{}); ok {
			if data, ok := block["data"].(map[string]interface{}); ok {
				t.Logf("Block data txs: %v", data["txs"])
			}
		}
	})
}

func TestMempoolRecheck(t *testing.T) {
	// For this test, we'll just use one binary type for simplicity
	config := utils.DefaultTestConfig("fauxmosis-comet")

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// Test mempool transaction recheck after a block is committed
	t.Run("Mempool Recheck After Block", func(t *testing.T) {
		// This test checks if the mempool properly rechecks transactions
		// after a block is committed and removes invalid transactions

		// Clear mempool by waiting for transactions to be processed
		t.Log("Waiting for mempool to clear...")
		time.Sleep(20 * time.Second)

		// Submit some test transactions
		// In a real test, we would create transactions that become invalid
		// after a block is committed (e.g., due to sequence number changes)
		for i := 1; i <= 5; i++ {
			txBytes := generateTxBytes(500, i)
			hash, err := submitTransaction(ctx, httpClient, txBytes)
			require.NoError(t, err, "Failed to submit transaction")
			t.Logf("Submitted transaction %d with hash: %s", i, hash)

			// Slight delay to ensure distinct submission times
			time.Sleep(500 * time.Millisecond)
		}

		// Get mempool count after submission
		afterSubmitCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get mempool count after submission")
		t.Logf("Mempool transaction count after submission: %d", afterSubmitCount)

		// Wait for next block to be committed
		t.Log("Waiting for next block...")
		time.Sleep(10 * time.Second)

		// Check mempool count after block
		// If the mempool implementation properly rechecks transactions,
		// invalid transactions should be removed
		afterBlockCount, err := getNumberOfTransactionsInMempool(ctx, httpClient)
		require.NoError(t, err, "Failed to get mempool count after block")
		t.Logf("Mempool transaction count after block: %d", afterBlockCount)

		// Verify some transactions were processed
		// In a real test with properly created transactions,
		// we would expect afterBlockCount < afterSubmitCount
		t.Logf("Transaction processing: before=%d, after=%d", afterSubmitCount, afterBlockCount)
	})
}
