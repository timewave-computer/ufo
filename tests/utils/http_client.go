package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// WaitForHeight waits for the node to reach the specified block height
func (c *HTTPClient) WaitForHeight(ctx context.Context, height int64, timeout time.Duration) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type BlockResponse struct {
		Result struct {
			Block struct {
				Header struct {
					Height string `json:"height"`
				} `json:"header"`
			} `json:"block"`
		} `json:"result"`
	}

	for {
		select {
		case <-ticker.C:
			var resp BlockResponse
			if err := c.Get(ctx, "/blocks/latest", &resp); err == nil {
				// Parse height from response
				var currentHeight int64
				fmt.Sscanf(resp.Result.Block.Header.Height, "%d", &currentHeight)

				if currentHeight >= height {
					return nil
				}
			}
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for block height %d", height)
		}
	}
}
