package fauxmosisintegration

import (
	"fmt"
	"sync"
	"time"

	"github.com/timewave/ufo/src/abci"
	"github.com/timewave/ufo/src/consensus"
)

// CometBFTAdapter is an adapter that makes UFO look like CometBFT to Osmosis
type CometBFTAdapter struct {
	app            *abci.Application
	consensusState *consensus.ConsensusState
	txCh           chan []byte
	stopCh         chan struct{}
	mutex          sync.Mutex
	isRunning      bool
}

// NewCometBFTAdapter creates a new adapter
func NewCometBFTAdapter(app *abci.Application, cs *consensus.ConsensusState) *CometBFTAdapter {
	return &CometBFTAdapter{
		app:            app,
		consensusState: cs,
		txCh:           make(chan []byte, 1000),
		stopCh:         make(chan struct{}),
		isRunning:      false,
	}
}

// Start starts the adapter's consensus process
func (a *CometBFTAdapter) Start() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("adapter already running")
	}

	a.isRunning = true
	go a.consensusLoop()
	return nil
}

// Stop stops the adapter's consensus process
func (a *CometBFTAdapter) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return fmt.Errorf("adapter not running")
	}

	close(a.stopCh)
	a.isRunning = false
	return nil
}

// BroadcastTx adds a transaction to the transaction pool
func (a *CometBFTAdapter) BroadcastTx(tx []byte) error {
	if !a.isRunning {
		return fmt.Errorf("adapter not running")
	}

	// First, check if the transaction is valid
	err := a.app.CheckTx(tx)
	if err != nil {
		return err
	}

	// We don't actually need to add to the channel since CheckTx
	// already added it to the app's transaction pool
	return nil
}

// consensusLoop runs the consensus process
func (a *CometBFTAdapter) consensusLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.runConsensusRound()
		}
	}
}

// runConsensusRound runs a single consensus round
func (a *CometBFTAdapter) runConsensusRound() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// Prepare the proposal
	txs := a.app.PrepareProposal()
	if len(txs) == 0 {
		// No transactions, nothing to do
		return
	}

	// Process the proposal
	accepted := a.app.ProcessProposal(txs)
	if !accepted {
		// Proposal not accepted, try again later
		return
	}

	// Run consensus round
	consensusTxs := make([]consensus.Transaction, 0, len(txs))
	for _, tx := range txs {
		consensusTxs = append(consensusTxs, consensus.Transaction(tx))
	}

	block, err := a.consensusState.RunConsensusRound(consensusTxs)
	if err != nil {
		fmt.Printf("Error running consensus round: %v\n", err)
		return
	}

	// Finalize the block
	events, err := a.app.FinalizeBlock(txs)
	if err != nil {
		fmt.Printf("Error finalizing block: %v\n", err)
		return
	}

	// Commit the block
	appHash := a.app.Commit()

	fmt.Printf("Committed block at height %d, hash: %s, app hash: %x, events: %s\n",
		block.Height, block.Hash, appHash, events)
}

// Status returns the node status (including sync status, validator info, etc.)
func (a *CometBFTAdapter) Status() map[string]interface{} {
	// Mock status information that Osmosis might expect
	latestHeight := a.consensusState.GetCurrentHeight()

	var latestBlockHash string
	if len(a.consensusState.CommittedBlocks) > 0 {
		latestBlock := a.consensusState.CommittedBlocks[len(a.consensusState.CommittedBlocks)-1]
		latestBlockHash = latestBlock.Hash
	}

	return map[string]interface{}{
		"node_info": map[string]interface{}{
			"network": "osmosis",
			"version": "UFO-v0.1.0",
		},
		"sync_info": map[string]interface{}{
			"latest_block_height": latestHeight,
			"latest_block_hash":   latestBlockHash,
			"catching_up":         false,
		},
		"validator_info": map[string]interface{}{
			"address":      a.consensusState.CurrentProposer.Address,
			"voting_power": a.consensusState.CurrentProposer.VotingPower,
		},
	}
}

// GetConsensusState returns the consensus state for direct access
func (a *CometBFTAdapter) GetConsensusState() *consensus.ConsensusState {
	return a.consensusState
}

// GetApplication returns the ABCI application for direct access
func (a *CometBFTAdapter) GetApplication() *abci.Application {
	return a.app
}

// GetValidatorSet returns the validator set
func (a *CometBFTAdapter) GetValidatorSet() *consensus.ValidatorSet {
	return a.consensusState.ValidatorSet
}
