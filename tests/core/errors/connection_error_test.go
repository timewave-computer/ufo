package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/timewave/ufo/tests/utils"
)

// TestConnectionTimeouts tests handling of connection timeouts.
func TestConnectionTimeouts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Test different timeout configurations
	timeoutTests := []struct {
		name           string
		timeoutMS      int
		endpoint       string
		expectTimeout  bool
		endpointParams map[string]string
	}{
		{
			name:          "Very short timeout (1ms)",
			timeoutMS:     1, // Extremely short timeout - should always time out
			endpoint:      "/cosmos/base/tendermint/v1beta1/blocks/latest",
			expectTimeout: true,
		},
		{
			name:          "Medium timeout (100ms)",
			timeoutMS:     100, // Short but might be enough for simple requests
			endpoint:      "/cosmos/base/tendermint/v1beta1/blocks/latest",
			expectTimeout: false, // May or may not time out, depends on system
		},
		{
			name:      "Long query with short timeout",
			timeoutMS: 50,
			endpoint:  "/cosmos/tx/v1beta1/txs",
			endpointParams: map[string]string{
				"events":           "tx.height>=1",
				"pagination.limit": "100",
			},
			expectTimeout: true, // Complex query likely to time out
		},
		{
			name:          "Reasonable timeout (5s)",
			timeoutMS:     5000,
			endpoint:      "/cosmos/base/tendermint/v1beta1/blocks/latest",
			expectTimeout: false, // Should not time out
		},
	}

	for _, tt := range timeoutTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom client with specified timeout
			customClient := utils.NewHTTPClientWithTimeout(node.Config.RESTAddress, time.Duration(tt.timeoutMS)*time.Millisecond)

			// Build the endpoint URL with any parameters
			url := tt.endpoint
			if tt.endpointParams != nil && len(tt.endpointParams) > 0 {
				url += "?"
				for k, v := range tt.endpointParams {
					url += k + "=" + v + "&"
				}
				// Remove trailing &
				url = url[:len(url)-1]
			}

			// Send the request
			_, err := customClient.Get(ctx, url)

			if tt.expectTimeout {
				require.Error(t, err, "Expected timeout error for %s", tt.name)
				require.Contains(t, err.Error(), "timeout", "Expected timeout-related error message")
				t.Logf("Got expected timeout error: %v", err)
			} else {
				// If we don't expect a timeout, the test is informational
				if err != nil {
					t.Logf("Request with %dms timeout resulted in error: %v", tt.timeoutMS, err)
				} else {
					t.Logf("Request with %dms timeout succeeded", tt.timeoutMS)
				}
			}
		})
	}

	// Verify we can still make requests with a reasonable timeout
	normalClient := utils.NewHTTPClientWithTimeout(node.Config.RESTAddress, 10*time.Second)
	resp, err := normalClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err, "Expected successful request with normal timeout")
	require.NotNil(t, resp["block"], "Expected valid response with normal timeout")
	t.Logf("Normal request succeeded after timeout tests")
}

// TestConnectionRefusedHandling tests handling of connection refused errors.
func TestConnectionRefusedHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Define a list of non-existent server endpoints to test
	nonExistentEndpoints := []struct {
		name     string
		address  string
		endpoint string
	}{
		{
			name:     "Non-existent localhost port",
			address:  "http://localhost:65535", // Port out of range or unlikely to be used
			endpoint: "/cosmos/base/tendermint/v1beta1/blocks/latest",
		},
		{
			name:     "Invalid IP address",
			address:  "http://127.0.0.1:65535", // Another unlikely port
			endpoint: "/cosmos/tx/v1beta1/txs",
		},
		{
			name:     "Non-existent domain",
			address:  "http://nonexistent.domain.that.does.not.exist:8080",
			endpoint: "/cosmos/bank/v1beta1/balances/cosmos1...",
		},
	}

	for _, tt := range nonExistentEndpoints {
		t.Run(tt.name, func(t *testing.T) {
			// Create client pointing to non-existent server
			nonExistentClient := utils.NewHTTPClient(tt.address)

			// Set a shorter timeout to make tests run faster
			nonExistentClient.SetTimeout(3 * time.Second)

			// Send request to non-existent server
			_, err := nonExistentClient.Get(ctx, tt.endpoint)

			// Verify error is returned
			require.Error(t, err, "Expected connection error for %s", tt.name)

			// The exact error message varies by platform/language, but should contain
			// some indication of connection failure
			t.Logf("Got expected connection error: %v", err)

			// Try different HTTP methods
			err = nonExistentClient.PostRaw(ctx, tt.endpoint, []byte(`{}`))
			require.Error(t, err, "Expected connection error for POST to %s", tt.name)
			t.Logf("Got expected connection error for POST: %v", err)
		})
	}

	// Test client recovery - after failed connections, a client should still work
	// with a valid server

	// First set up a valid node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)
	defer node.Cleanup()

	// Create a client that first hits invalid endpoints, then a valid one
	client := utils.NewHTTPClient("http://localhost:65535") // Start with invalid

	// Try an invalid endpoint
	_, err = client.Get(ctx, "/some/endpoint")
	require.Error(t, err, "Expected error for invalid endpoint")

	// Now update the client to a valid endpoint and verify it works
	client.SetBaseURL(node.Config.RESTAddress)
	resp, err := client.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err, "Expected successful request after updating base URL")
	require.NotNil(t, resp["block"], "Expected valid response after recovery")
	t.Logf("Client successfully recovered after connection errors")
}

// TestServerDisconnectionRecovery tests recovery from server disconnection.
func TestServerDisconnectionRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Create a new test node
	node, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err)

	// Don't defer cleanup yet, as we'll manually stop and restart

	// Create HTTP client
	httpClient := utils.NewHTTPClient(node.Config.RESTAddress)

	// Wait for the node to be ready
	time.Sleep(5 * time.Second)

	// Step 1: Verify initial connectivity
	initialResp, err := httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.NoError(t, err, "Expected successful initial request")
	require.NotNil(t, initialResp["block"], "Expected valid initial response")
	t.Logf("Initial connectivity verified")

	// Step 2: Stop the server
	t.Logf("Stopping the node...")
	node.Cleanup()

	// Give the node time to fully stop
	time.Sleep(5 * time.Second)

	// Step 3: Send requests and verify errors
	_, err = httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
	require.Error(t, err, "Expected error after server shutdown")
	t.Logf("Got expected error after server shutdown: %v", err)

	// Try a few more endpoints
	_, err = httpClient.Get(ctx, "/cosmos/bank/v1beta1/supply")
	require.Error(t, err, "Expected error after server shutdown")

	err = httpClient.PostRaw(ctx, "/cosmos/tx/v1beta1/simulate", []byte(`{}`))
	require.Error(t, err, "Expected error after server shutdown")

	// Step 4: Restart the server
	t.Logf("Restarting the node...")
	newNode, err := utils.NewTestNode(ctx, utils.DefaultTestConfig())
	require.NoError(t, err, "Failed to restart test node")
	defer newNode.Cleanup()

	// Give the node time to fully start
	time.Sleep(10 * time.Second)

	// Step 5: Test client recovery
	// Update the client to use the new node's address (which might be the same)
	httpClient.SetBaseURL(newNode.Config.RESTAddress)

	// Create a retry mechanism
	maxRetries := 5
	retryDelay := 2 * time.Second
	var recoveryResp map[string]interface{}

	// Try to reconnect with retries
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		recoveryResp, err = httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/1")
		if err == nil {
			break
		}
		lastErr = err
		t.Logf("Reconnection attempt %d failed: %v, retrying in %v", i+1, err, retryDelay)
		time.Sleep(retryDelay)
	}

	require.NoError(t, lastErr, "Failed to reconnect after server restart")
	require.NotNil(t, recoveryResp["block"], "Expected valid response after reconnection")
	t.Logf("Successfully reconnected after server restart")

	// Step 6: Verify continued connectivity with multiple requests
	_, err = httpClient.Get(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest")
	require.NoError(t, err, "Expected successful request after reconnection")

	_, err = httpClient.Get(ctx, "/cosmos/auth/v1beta1/params")
	require.NoError(t, err, "Expected successful request after reconnection")

	t.Logf("Connectivity fully restored after server restart")
}

// TestPartialResponseHandling tests handling of partial responses.
func TestPartialResponseHandling(t *testing.T) {
	// This test requires a custom mock server that can send partial responses
	t.Skip("This test requires a custom mock server and is deferred to a future implementation phase")

	/*
		Implementation would require:
		1. Setting up a custom HTTP server that can be controlled to send partial responses
		2. Configuring the server to deliberately send incomplete responses (cut off mid-stream)
		3. Testing client behavior under these conditions
		4. Verifying appropriate error handling and recovery

		Since this requires significant custom infrastructure beyond the scope of the
		current implementation phase, we're skipping this test for now.
	*/
}
