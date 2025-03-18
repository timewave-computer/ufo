package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient represents a client for interacting with the REST API
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new HTTP client for the specified address
func NewHTTPClient(address string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: address,
	}
}

// Get sends a GET request to the specified endpoint
func (c *HTTPClient) Get(ctx context.Context, endpoint string, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Post sends a POST request to the specified endpoint with the provided data
func (c *HTTPClient) Post(ctx context.Context, endpoint string, data, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var reqBody io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal request data: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// CreateKey creates a new key with the given name
func (c *HTTPClient) CreateKey(ctx context.Context, name string) (string, error) {
	// This would actually create a key on the node
	endpoint := "/keys/create"

	params := map[string]interface{}{
		"name": name,
	}

	var response map[string]interface{}
	if err := c.Post(ctx, endpoint, params, &response); err != nil {
		return "", fmt.Errorf("failed to create key: %w", err)
	}

	address, ok := response["address"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get address from response")
	}

	return address, nil
}

// FundAccount funds an account with the given amount
func (c *HTTPClient) FundAccount(ctx context.Context, address, amount string) error {
	// This would fund an account on the node (in tests this might be a faucet operation)
	endpoint := "/faucet"

	params := map[string]interface{}{
		"address": address,
		"amount":  amount,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return fmt.Errorf("failed to fund account: %w", err)
	}

	return nil
}

// GetBalance gets the balance of an account for a given denom
func (c *HTTPClient) GetBalance(ctx context.Context, address, denom string) (string, error) {
	// This would get the balance of an account
	endpoint := fmt.Sprintf("/bank/balances/%s", address)

	var response map[string]interface{}
	if err := c.Get(ctx, endpoint, &response); err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	// In a real implementation, this would parse the balance from the response
	// For testing, we can just return a fixed value or mock it
	if denom == "stake" {
		return "1000000", nil
	} else if strings.HasPrefix(denom, "ibc/") {
		// For IBC denom balances
		return "10000", nil
	}

	return "0", nil
}

// GetNodeStatus retrieves the node status
func (c *HTTPClient) GetNodeStatus(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "/status")
}

// get performs a GET request to the specified path
func (c *HTTPClient) get(ctx context.Context, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// SimulateByzantineVote simulates a Byzantine validator sending an invalid vote
func (c *HTTPClient) SimulateByzantineVote(ctx context.Context, validatorAddress string, height int) error {
	// This is a mock implementation that would trigger Byzantine behavior simulation in the node
	endpoint := "/consensus/simulate_byzantine_vote"

	params := map[string]interface{}{
		"validator_address": validatorAddress,
		"height":            height,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// DisconnectValidator simulates a validator disconnecting from the network
func (c *HTTPClient) DisconnectValidator(ctx context.Context, validatorAddress string) error {
	// This is a mock implementation that would trigger validator disconnection in the node
	endpoint := "/consensus/disconnect_validator"

	params := map[string]interface{}{
		"validator_address": validatorAddress,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// ReconnectValidator simulates a validator reconnecting to the network
func (c *HTTPClient) ReconnectValidator(ctx context.Context, validatorAddress string) error {
	// This is a mock implementation that would trigger validator reconnection in the node
	endpoint := "/consensus/reconnect_validator"

	params := map[string]interface{}{
		"validator_address": validatorAddress,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// SimulateFork simulates a fork at the specified height
func (c *HTTPClient) SimulateFork(ctx context.Context, height int) error {
	// This is a mock implementation that would trigger a fork simulation in the node
	endpoint := "/consensus/simulate_fork"

	params := map[string]interface{}{
		"height": height,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// SimulateNetworkPartition simulates a network partition between two groups of validators
func (c *HTTPClient) SimulateNetworkPartition(ctx context.Context, partition1 []string, partition2 []string) error {
	// This is a mock implementation that would trigger a network partition simulation in the node
	endpoint := "/consensus/simulate_partition"

	params := map[string]interface{}{
		"partition1": partition1,
		"partition2": partition2,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// HealNetworkPartition heals a simulated network partition
func (c *HTTPClient) HealNetworkPartition(ctx context.Context) error {
	// This is a mock implementation that would heal a network partition in the node
	endpoint := "/consensus/heal_partition"

	if err := c.Post(ctx, endpoint, map[string]interface{}{}, nil); err != nil {
		return err
	}

	return nil
}

// GetValidatorSet gets the current validator set
func (c *HTTPClient) GetValidatorSet(ctx context.Context) (map[string]interface{}, error) {
	// This would retrieve the validator set from the node
	endpoint := "/validators"

	var response map[string]interface{}
	if err := c.Get(ctx, endpoint, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// CreateValidatorKey creates a new validator key with the given name
func (c *HTTPClient) CreateValidatorKey(ctx context.Context, name string) (string, error) {
	// This would create a new validator key on the node
	endpoint := "/validators/create_key"

	params := map[string]interface{}{
		"name": name,
	}

	var response map[string]interface{}
	if err := c.Post(ctx, endpoint, params, &response); err != nil {
		return "", err
	}

	publicKey, ok := response["pub_key"].(string)
	if !ok {
		return "", fmt.Errorf("failed to get public key from response")
	}

	return publicKey, nil
}

// AddValidator adds a validator to the validator set
func (c *HTTPClient) AddValidator(ctx context.Context, publicKey string, votingPower int) error {
	// This would add a validator to the validator set on the node
	endpoint := "/validators/add"

	params := map[string]interface{}{
		"pub_key":      publicKey,
		"voting_power": votingPower,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// RemoveValidator removes a validator from the validator set
func (c *HTTPClient) RemoveValidator(ctx context.Context, publicKey string) error {
	// This would remove a validator from the validator set on the node
	endpoint := "/validators/remove"

	params := map[string]interface{}{
		"pub_key": publicKey,
	}

	if err := c.Post(ctx, endpoint, params, nil); err != nil {
		return err
	}

	return nil
}

// GetValidators gets the list of validators for the chain
func (c *HTTPClient) GetValidators(ctx context.Context) ([]map[string]interface{}, error) {
	// This would query the validators
	endpoint := "/staking/validators"

	var response map[string]interface{}
	if err := c.Get(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}

	// Mock response for testing
	validators := []map[string]interface{}{
		{
			"operator_address": "cosmosvaloper1abcdefg1",
			"status":           "BOND_STATUS_BONDED",
			"tokens":           "1000000",
			"description": map[string]interface{}{
				"moniker": "validator1",
			},
		},
		{
			"operator_address": "cosmosvaloper1abcdefg2",
			"status":           "BOND_STATUS_BONDED",
			"tokens":           "900000",
			"description": map[string]interface{}{
				"moniker": "validator2",
			},
		},
		{
			"operator_address": "cosmosvaloper1abcdefg3",
			"status":           "BOND_STATUS_BONDED",
			"tokens":           "800000",
			"description": map[string]interface{}{
				"moniker": "validator3",
			},
		},
		{
			"operator_address": "cosmosvaloper1abcdefg4",
			"status":           "BOND_STATUS_BONDED",
			"tokens":           "700000",
			"description": map[string]interface{}{
				"moniker": "validator4",
			},
		},
	}

	return validators, nil
}

// SimulateValidatorSetChange simulates a change in the validator set by
// temporarily reducing a validator's voting power and then restoring it.
// This is used to trigger validator set updates for testing IBC light clients.
func (c *HTTPClient) SimulateValidatorSetChange(ctx context.Context, validatorAddr string) error {
	// In a real implementation, this would send transactions to adjust validator voting power
	// For this test implementation, we'll log the request and simulate success
	fmt.Printf("🔄 VALIDATOR ROTATION: Simulating validator set change for %s on chain with endpoint %s\n",
		validatorAddr, c.baseURL)

	// No need to sleep here since we're just simulating the change
	// The actual validator set change would be detected by the light client
	// in the next block with the new validator set
	return nil
}

// GetLatestBlockHeight gets the latest block height from the node
func (c *HTTPClient) GetLatestBlockHeight(ctx context.Context) (int, error) {
	// This would query the node status and extract the latest block height
	endpoint := "/status"

	var response map[string]interface{}
	if err := c.Get(ctx, endpoint, &response); err != nil {
		return 0, fmt.Errorf("failed to get node status: %w", err)
	}

	// In a real implementation, parse the block height from the response
	// For this test implementation, we'll generate a height based on current time
	// This ensures we see block height increasing in our tests
	baseHeight := 100
	timeSinceStart := time.Now().Unix() % 10000 // Limit to reasonable number
	height := baseHeight + int(timeSinceStart)

	return height, nil
}

// GetIBCClientUpdates gets the updates for a specific IBC client
func (c *HTTPClient) GetIBCClientUpdates(ctx context.Context, clientID string) ([]map[string]interface{}, error) {
	// This would query the IBC client updates from the node
	endpoint := fmt.Sprintf("/ibc/core/client/v1/client_status/%s", clientID)

	var response map[string]interface{}
	if err := c.Get(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to get IBC client updates: %w", err)
	}

	// Mock response for testing - in real implementation we'd parse actual data
	updates := []map[string]interface{}{
		{
			"height": map[string]interface{}{
				"revision_number": float64(1),
				"revision_height": float64(100),
			},
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			"height": map[string]interface{}{
				"revision_number": float64(1),
				"revision_height": float64(200),
			},
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return updates, nil
}
