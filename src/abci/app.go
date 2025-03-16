package abci

import (
	"sync"

	"github.com/timewave/ufo/src/chains/osmosis"
	"github.com/timewave/ufo/src/memstate"
	"github.com/timewave/ufo/src/transactions"
)

// ApplicationInterface defines the interface for ABCI++ applications
type ApplicationInterface interface {
	CheckTx(tx []byte) error
	PrepareProposal() [][]byte
	ProcessProposal(txs [][]byte) bool
	DeliverTx(tx []byte) (string, error)
	FinalizeBlock(txs [][]byte) (string, error)
	Commit() []byte
}

// Application is an ABCI++ application
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
	registry.RegisterProcessor(kvProcessor)
	
	// Register Osmosis processor
	osmosisProcessor := osmosis.NewTxProcessor()
	registry.RegisterProcessor(osmosisProcessor)
	
	// Add some test validators for Osmosis
	osmosisProcessor.AddValidator("osmovaloper1", 0.1, osmosis.Coin{Denom: "uosmo", Amount: "1000000"})
	osmosisProcessor.AddValidator("osmovaloper2", 0.05, osmosis.Coin{Denom: "uosmo", Amount: "2000000"})
	
	// Add a test liquidity pool
	osmosisProcessor.AddLiquidityPool(1, 
		[]osmosis.Coin{
			{Denom: "uosmo", Amount: "1000000"},
			{Denom: "uatom", Amount: "500000"},
		},
		osmosis.Coin{Denom: "gamm/1", Amount: "1500000"},
		0.003)
	
	return &Application{
		State:      state,
		TxPool:     make([][]byte, 0),
		TxRegistry: registry,
	}
}

// CheckTx validates a transaction
func (app *Application) CheckTx(tx []byte) error {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	// Use the registry to check the transaction
	err := app.TxRegistry.CheckTx(tx)
	if err != nil {
		return err
	}
	
	// Add to transaction pool
	app.TxPool = append(app.TxPool, tx)
	return nil
}

// PrepareProposal prepares a proposal
func (app *Application) PrepareProposal() [][]byte {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	return app.TxPool
}

// ProcessProposal processes a proposal
func (app *Application) ProcessProposal(txs [][]byte) bool {
	return true
}

// deliverTxInternal processes a transaction without acquiring locks
func (app *Application) deliverTxInternal(tx []byte) (string, error) {
	// Use the registry to deliver the transaction
	return app.TxRegistry.DeliverTx(tx)
}

// DeliverTx processes a transaction
func (app *Application) DeliverTx(tx []byte) (string, error) {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	return app.deliverTxInternal(tx)
}

// FinalizeBlock executes all transactions in a block
func (app *Application) FinalizeBlock(txs [][]byte) (string, error) {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	events := ""
	for _, tx := range txs {
		event, err := app.deliverTxInternal(tx)
		if err != nil {
			return "", err
		}
		events += event + "\n"
	}

	// Clear the transaction pool
	app.TxPool = make([][]byte, 0)

	return events, nil
}

// Commit commits the state and returns the app hash
func (app *Application) Commit() []byte {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()

	return app.State.Commit()
} 