package fauxmosisintegration

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// RPCClient is a mock implementation of the CometBFT RPC client that Osmosis uses
type RPCClient struct {
	adapter *CometBFTAdapter
}

// NewRPCClient creates a new mock RPC client
func NewRPCClient(adapter *CometBFTAdapter) *RPCClient {
	return &RPCClient{
		adapter: adapter,
	}
}

// ABCIInfo returns information about the ABCI application
func (c *RPCClient) ABCIInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"response": map[string]interface{}{
			"version":             "0.1.0",
			"app_version":         "1",
			"last_block_height":   c.adapter.GetConsensusState().GetCurrentHeight(),
			"last_block_app_hash": hex.EncodeToString([]byte("app_hash")),
		},
	}, nil
}

// ABCIQuery performs an ABCI query
func (c *RPCClient) ABCIQuery(ctx context.Context, path string, data []byte, height int64, prove bool) (map[string]interface{}, error) {
	// In a real implementation, we would route this to the appropriate handler
	// For now, we'll return a mock response
	return map[string]interface{}{
		"response": map[string]interface{}{
			"code":   0,
			"log":    "",
			"info":   "",
			"index":  0,
			"key":    "",
			"value":  hex.EncodeToString([]byte(`{"result":"mock abci query result"}`)),
			"proof":  "",
			"height": c.adapter.GetConsensusState().GetCurrentHeight(),
		},
	}, nil
}

// BroadcastTxCommit broadcasts a transaction and waits for it to be committed
func (c *RPCClient) BroadcastTxCommit(ctx context.Context, tx []byte) (map[string]interface{}, error) {
	err := c.adapter.BroadcastTx(tx)
	if err != nil {
		return nil, err
	}

	// Wait for a simulated commit (in real life would wait for block inclusion)
	time.Sleep(100 * time.Millisecond)

	// Get the result of delivering the transaction
	result, err := c.adapter.GetApplication().DeliverTx(tx)
	if err != nil {
		return map[string]interface{}{
			"check_tx": map[string]interface{}{
				"code": 0,
				"log":  "transaction validated",
			},
			"deliver_tx": map[string]interface{}{
				"code": 1,
				"log":  err.Error(),
			},
			"hash":   hex.EncodeToString(getSampleTxHash(tx)),
			"height": c.adapter.GetConsensusState().GetCurrentHeight(),
		}, nil
	}

	return map[string]interface{}{
		"check_tx": map[string]interface{}{
			"code": 0,
			"log":  "transaction validated",
		},
		"deliver_tx": map[string]interface{}{
			"code": 0,
			"log":  result,
		},
		"hash":   hex.EncodeToString(getSampleTxHash(tx)),
		"height": c.adapter.GetConsensusState().GetCurrentHeight(),
	}, nil
}

// BroadcastTxAsync broadcasts a transaction asynchronously
func (c *RPCClient) BroadcastTxAsync(ctx context.Context, tx []byte) (map[string]interface{}, error) {
	err := c.adapter.BroadcastTx(tx)
	if err != nil {
		return map[string]interface{}{
			"code":   1,
			"log":    err.Error(),
			"hash":   hex.EncodeToString(getSampleTxHash(tx)),
			"height": 0,
		}, nil
	}

	return map[string]interface{}{
		"code":   0,
		"log":    "transaction added to mempool",
		"hash":   hex.EncodeToString(getSampleTxHash(tx)),
		"height": 0,
	}, nil
}

// BroadcastTxSync broadcasts a transaction synchronously
func (c *RPCClient) BroadcastTxSync(ctx context.Context, tx []byte) (map[string]interface{}, error) {
	err := c.adapter.BroadcastTx(tx)
	if err != nil {
		return map[string]interface{}{
			"code":   1,
			"log":    err.Error(),
			"hash":   hex.EncodeToString(getSampleTxHash(tx)),
			"height": 0,
		}, nil
	}

	return map[string]interface{}{
		"code":   0,
		"log":    "transaction added to mempool",
		"hash":   hex.EncodeToString(getSampleTxHash(tx)),
		"height": 0,
	}, nil
}

// Block returns information about a block
func (c *RPCClient) Block(ctx context.Context, height *int64) (map[string]interface{}, error) {
	cs := c.adapter.GetConsensusState()

	var h int64
	if height == nil || *height == 0 {
		h = cs.GetCurrentHeight()
	} else {
		h = *height
	}

	// Check if we have the block
	if h <= 0 || h > cs.GetCurrentHeight() {
		return nil, fmt.Errorf("block not found: height %d", h)
	}

	// We need to adjust for 0-indexed vs 1-indexed
	blockIndex := int(h - 1)
	if blockIndex >= len(cs.CommittedBlocks) {
		return nil, fmt.Errorf("block not found: height %d", h)
	}

	block := cs.CommittedBlocks[blockIndex]

	txs := make([]string, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		txs = append(txs, hex.EncodeToString([]byte(tx)))
	}

	return map[string]interface{}{
		"block": map[string]interface{}{
			"header": map[string]interface{}{
				"height":           h,
				"time":             time.Now().Format(time.RFC3339),
				"last_block_id":    map[string]interface{}{"hash": "mock_previous_hash"},
				"last_commit_hash": "mock_last_commit_hash",
				"data_hash":        "mock_data_hash",
				"validators_hash":  "mock_validators_hash",
				"app_hash":         "mock_app_hash",
			},
			"data": map[string]interface{}{
				"txs": txs,
			},
			"last_commit": map[string]interface{}{
				"height":     h - 1,
				"round":      0,
				"block_id":   map[string]interface{}{"hash": "mock_previous_hash"},
				"signatures": []interface{}{},
			},
		},
	}, nil
}

// BlockResults returns the results of transactions in a block
func (c *RPCClient) BlockResults(ctx context.Context, height *int64) (map[string]interface{}, error) {
	cs := c.adapter.GetConsensusState()

	var h int64
	if height == nil || *height == 0 {
		h = cs.GetCurrentHeight()
	} else {
		h = *height
	}

	// Check if we have the block
	if h <= 0 || h > cs.GetCurrentHeight() {
		return nil, fmt.Errorf("block not found: height %d", h)
	}

	// We need to adjust for 0-indexed vs 1-indexed
	blockIndex := int(h - 1)
	if blockIndex >= len(cs.CommittedBlocks) {
		return nil, fmt.Errorf("block not found: height %d", h)
	}

	block := cs.CommittedBlocks[blockIndex]

	// Create mock transaction results
	txResults := make([]map[string]interface{}, 0, len(block.Transactions))
	for range block.Transactions {
		txResults = append(txResults, map[string]interface{}{
			"code":       0,
			"data":       "",
			"log":        "transaction executed successfully",
			"info":       "",
			"gas_wanted": "200000",
			"gas_used":   "50000",
			"events":     []interface{}{},
		})
	}

	return map[string]interface{}{
		"height":      h,
		"txs_results": txResults,
	}, nil
}

// Commit returns the commit information for a block
func (c *RPCClient) Commit(ctx context.Context, height *int64) (map[string]interface{}, error) {
	cs := c.adapter.GetConsensusState()

	var h int64
	if height == nil || *height == 0 {
		h = cs.GetCurrentHeight()
	} else {
		h = *height
	}

	return map[string]interface{}{
		"signed_header": map[string]interface{}{
			"header": map[string]interface{}{
				"height":   h,
				"time":     time.Now().Format(time.RFC3339),
				"chain_id": "osmosis-mock",
				"app_hash": "mock_app_hash",
			},
			"commit": map[string]interface{}{
				"height":     h,
				"signatures": []interface{}{},
			},
		},
	}, nil
}

// Genesis returns the genesis state
func (c *RPCClient) Genesis(ctx context.Context) (map[string]interface{}, error) {
	// Return a minimal mock genesis response
	return map[string]interface{}{
		"genesis": map[string]interface{}{
			"chain_id":     "osmosis-mock",
			"genesis_time": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"app_hash":     "",
		},
	}, nil
}

// NetInfo returns network information
func (c *RPCClient) NetInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"listening": true,
		"listeners": []string{"tcp://0.0.0.0:26656"},
		"n_peers":   0,
		"peers":     []interface{}{},
	}, nil
}

// Status returns the node status
func (c *RPCClient) Status(ctx context.Context) (map[string]interface{}, error) {
	return c.adapter.Status(), nil
}

// Validators returns the validator set at a specified height
func (c *RPCClient) Validators(ctx context.Context, height *int64, page, perPage *int) (map[string]interface{}, error) {
	cs := c.adapter.GetConsensusState()
	validatorSet := c.adapter.GetValidatorSet()

	validators := make([]map[string]interface{}, 0, validatorSet.Size())
	for _, validator := range validatorSet.Validators {
		validators = append(validators, map[string]interface{}{
			"address":           validator.Address,
			"pub_key":           map[string]interface{}{"type": "ed25519", "value": "mock_pubkey"},
			"voting_power":      validator.VotingPower,
			"proposer_priority": 0,
		})
	}

	return map[string]interface{}{
		"block_height": cs.GetCurrentHeight(),
		"validators":   validators,
		"total":        validatorSet.Size(),
	}, nil
}

// Simulate returns a simulated response for a transaction
func (c *RPCClient) Simulate(ctx context.Context, tx []byte) (map[string]interface{}, error) {
	err := c.adapter.GetApplication().CheckTx(tx)
	if err != nil {
		return nil, err
	}

	// If check passes, simulate success
	return map[string]interface{}{
		"gas_info": map[string]interface{}{
			"gas_wanted": 200000,
			"gas_used":   50000,
		},
		"result": map[string]interface{}{
			"data":   "",
			"log":    "transaction would be executed successfully",
			"events": []interface{}{},
		},
	}, nil
}

// Health returns the health status of the node
func (c *RPCClient) Health(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// getSampleTxHash returns a simple hash for a transaction
func getSampleTxHash(tx []byte) []byte {
	if len(tx) < 8 {
		return tx
	}
	return tx[:8]
}

// RPCResponse represents a generic JSON-RPC response
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}
