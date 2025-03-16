// Package abci provides ABCI++ application implementation
package abci

import (
	"sync"

	"github.com/timewave/ufo/src/chains/osmosis"
	"github.com/timewave/ufo/src/memstate"
	"github.com/timewave/ufo/src/transactions"
)

// ApplicationInterface defines the ABCI++ application interface
type ApplicationInterface interface {
	CheckTx(tx []byte) error
	PrepareProposal() [][]byte
	ProcessProposal(txs [][]byte) bool
	DeliverTx(tx []byte) (string, error)
	FinalizeBlock(txs [][]byte) (string, error)
	Commit() []byte
}

// Application implements the ABCI++ application
type Application struct {
	// State is the in-memory state
	State *memstate.State
	// TxPool is the transaction pool
	TxPool [][]byte
	// Mutex for thread safety
	Mutex sync.Mutex
	// TxRegistry holds registered transaction processors
	TxRegistry *transactions.TxProcessorRegistry
}

// Make sure Application implements ApplicationInterface
var _ ApplicationInterface = (*Application)(nil)

// NewApplication creates a new ABCI++ application
func NewApplication() *Application {
	state := memstate.NewState()
	registry := transactions.NewTxProcessorRegistry()

	// Register KV store processor
	kvProcessor := transactions.NewKVStoreTxProcessor(state)
	registry.RegisterProcessor("kv", kvProcessor)

	// Register Osmosis processor
	osmosisProcessor := osmosis.NewTxProcessor("osmosis-1")
	registry.RegisterProcessor("osmosis", osmosisProcessor)

	// Add some test validators for Osmosis
	osmosisProcessor.AddValidator("osmovaloper1", 10, osmosis.NewCoin("uosmo", 1000000))
	osmosisProcessor.AddValidator("osmovaloper2", 5, osmosis.NewCoin("uosmo", 2000000))

	// Add a test liquidity pool
	osmosisProcessor.AddLiquidityPool(1,
		[]osmosis.Coin{
			osmosis.NewCoin("uosmo", 5000000),
			osmosis.NewCoin("uatom", 1000000),
		},
		10000000,
		0.003)

	// Create the application
	return &Application{
		State:      state,
		TxRegistry: registry,
		TxPool:     make([][]byte, 0),
	}
}

// CheckTx validates a transaction before including it in the mempool
func (app *Application) CheckTx(tx []byte) error {
	// Simplified implementation
	return nil
}

// PrepareProposal prepares a block proposal
func (app *Application) PrepareProposal() [][]byte {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	return app.TxPool
}

// ProcessProposal validates a block proposal
func (app *Application) ProcessProposal(txs [][]byte) bool {
	return true
}

// deliverTxInternal processes a transaction
func (app *Application) deliverTxInternal(tx []byte) (string, error) {
	// Simplified implementation
	return "", nil
}

// DeliverTx processes a transaction during block execution
func (app *Application) DeliverTx(tx []byte) (string, error) {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	return app.deliverTxInternal(tx)
}

// FinalizeBlock processes all transactions in a block
func (app *Application) FinalizeBlock(txs [][]byte) (string, error) {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	result := ""
	for _, tx := range txs {
		r, err := app.deliverTxInternal(tx)
		if err != nil {
			return "", err
		}
		result += r
	}

	return result, nil
}

// Commit commits the current state
func (app *Application) Commit() []byte {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	// Reset tx pool after commit
	app.TxPool = make([][]byte, 0)

	// Commit state changes
	return app.State.Commit()
}
