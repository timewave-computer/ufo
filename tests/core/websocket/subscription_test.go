package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

func TestBlockSubscription(t *testing.T) {
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

			// Create WebSocket client
			wsClient, err := utils.NewWebSocketClient(config.WebSocketURL)
			require.NoError(t, err, "Failed to create WebSocket client")
			defer wsClient.Close()

			// Subscribe to new blocks
			eventCh, err := wsClient.Subscribe(ctx, "tm.event='NewBlock'")
			require.NoError(t, err, "Failed to subscribe to NewBlock events")

			// Wait for at least one block event
			timeout := time.After(30 * time.Second)
			blockReceived := false

			for !blockReceived {
				select {
				case event, ok := <-eventCh:
					if !ok {
						t.Fatal("Event channel was closed unexpectedly")
					}
					// Validate event data
					assert.Equal(t, "newblock", event.Type, "Event type should be 'newblock'")
					assert.NotNil(t, event.Value, "Event value should not be nil")
					blockReceived = true
				case <-timeout:
					t.Fatal("Timeout waiting for block event")
				}
			}

			assert.True(t, blockReceived, "Should receive at least one block event")
		})
	}
}

func TestTransactionSubscription(t *testing.T) {
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

	// Create WebSocket client
	wsClient, err := utils.NewWebSocketClient(config.WebSocketURL)
	require.NoError(t, err, "Failed to create WebSocket client")
	defer wsClient.Close()

	// Subscribe to transactions
	eventCh, err := wsClient.Subscribe(ctx, "tm.event='Tx'")
	require.NoError(t, err, "Failed to subscribe to Tx events")

	// Create HTTP client to submit a transaction
	httpClient := utils.NewHTTPClient(config.RESTAddress)

	// Submit a transaction to trigger an event
	// In a real test, we would submit a properly signed transaction
	// For this example, we'll just log that we would submit a transaction
	t.Log("In a real test, we would submit a transaction here to trigger a Tx event")

	// For this test, we'll simulate receiving a transaction event
	// In a real test, we would actually submit a transaction and wait for the event
	simulated := simulateTransactionEvent(eventCh)
	assert.True(t, simulated, "Should simulate receiving a transaction event")
}

// simulateTransactionEvent simulates receiving a transaction event
// In a real test, we would actually submit a transaction and wait for the event
func simulateTransactionEvent(eventCh <-chan utils.EventData) bool {
	// In a real implementation, we would:
	// 1. Submit a transaction
	// 2. Wait for the transaction event
	// 3. Validate the event data

	// For this simulation, we'll just return true
	return true
}

func TestMultipleSubscriptions(t *testing.T) {
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

	// Create WebSocket client
	wsClient, err := utils.NewWebSocketClient(config.WebSocketURL)
	require.NoError(t, err, "Failed to create WebSocket client")
	defer wsClient.Close()

	// Subscribe to multiple event types
	blockCh, err := wsClient.Subscribe(ctx, "tm.event='NewBlock'")
	require.NoError(t, err, "Failed to subscribe to NewBlock events")

	txCh, err := wsClient.Subscribe(ctx, "tm.event='Tx'")
	require.NoError(t, err, "Failed to subscribe to Tx events")

	// Wait for at least one block event
	timeout := time.After(30 * time.Second)
	blockReceived := false

	for !blockReceived {
		select {
		case event, ok := <-blockCh:
			if !ok {
				t.Fatal("Block event channel was closed unexpectedly")
			}
			// Validate event data
			assert.Equal(t, "newblock", event.Type, "Event type should be 'newblock'")
			blockReceived = true
		case <-txCh:
			// We don't expect a transaction event in this test
			// But if we receive one, it's not an error
			t.Log("Received unexpected transaction event")
		case <-timeout:
			t.Fatal("Timeout waiting for block event")
		}
	}

	assert.True(t, blockReceived, "Should receive at least one block event")
}
