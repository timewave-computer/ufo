package fauxmosisintegration

import (
	"fmt"

	"github.com/timewave/ufo/src/abci"
	"github.com/timewave/ufo/src/consensus"
)

// FauxmosisUFOIntegration represents the integration between Fauxmosis and UFO
type FauxmosisUFOIntegration struct {
	App            *abci.Application
	ConsensusState *consensus.ConsensusState
	Adapter        *CometBFTAdapter
	RPCClient      *RPCClient
	HTTPServer     *RPCHTTPServer
	IsRunning      bool
}

// NewFauxmosisUFOIntegration creates a new integration instance
func NewFauxmosisUFOIntegration() *FauxmosisUFOIntegration {
	// Create validators
	validators := []*consensus.Validator{
		consensus.NewValidator("val1", "address1", 10),
		consensus.NewValidator("val2", "address2", 10),
		consensus.NewValidator("val3", "address3", 10),
		consensus.NewValidator("val4", "address4", 10),
	}

	// Create validator set
	validatorSet := consensus.NewValidatorSet(validators)
	fmt.Println("Created validator set with", validatorSet.Size(), "validators")

	// Create proposer selector
	proposerSelector := consensus.NewRoundRobinProposerSelector()

	// Create consensus state
	cs := consensus.NewConsensusState(validatorSet, proposerSelector)
	fmt.Println("Initial height:", cs.GetCurrentHeight())
	fmt.Println("Initial proposer:", cs.CurrentProposer.ID)

	// Create ABCI++ application
	app := abci.NewApplication()
	fmt.Println("Created ABCI++ application")

	// Create adapter
	adapter := NewCometBFTAdapter(app, cs)

	// Create RPC client
	rpcClient := NewRPCClient(adapter)

	// Create HTTP server
	httpServer := NewRPCHTTPServer(rpcClient, ":26657")

	return &FauxmosisUFOIntegration{
		App:            app,
		ConsensusState: cs,
		Adapter:        adapter,
		RPCClient:      rpcClient,
		HTTPServer:     httpServer,
		IsRunning:      false,
	}
}

// Start initializes and starts the integration
func (i *FauxmosisUFOIntegration) Start() error {
	if i.IsRunning {
		return fmt.Errorf("integration already running")
	}

	// Start the adapter
	err := i.Adapter.Start()
	if err != nil {
		return fmt.Errorf("failed to start adapter: %v", err)
	}

	// Start the HTTP server
	err = i.HTTPServer.Start()
	if err != nil {
		i.Adapter.Stop()
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}

	i.IsRunning = true
	fmt.Println("Fauxmosis with UFO integration started")
	return nil
}

// Stop stops the integration
func (i *FauxmosisUFOIntegration) Stop() error {
	if !i.IsRunning {
		return fmt.Errorf("integration not running")
	}

	// Stop the HTTP server
	err := i.HTTPServer.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop HTTP server: %v", err)
	}

	// Stop the adapter
	err = i.Adapter.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop adapter: %v", err)
	}

	i.IsRunning = false
	fmt.Println("Fauxmosis with UFO integration stopped")
	return nil
}

// GetApp returns the ABCI application
func (i *FauxmosisUFOIntegration) GetApp() *abci.Application {
	return i.App
}

// GetConsensusState returns the consensus state
func (i *FauxmosisUFOIntegration) GetConsensusState() *consensus.ConsensusState {
	return i.ConsensusState
}

// GetAdapter returns the CometBFT adapter
func (i *FauxmosisUFOIntegration) GetAdapter() *CometBFTAdapter {
	return i.Adapter
}

// GetRPCClient returns the RPC client
func (i *FauxmosisUFOIntegration) GetRPCClient() *RPCClient {
	return i.RPCClient
}

// GetHTTPServer returns the HTTP server
func (i *FauxmosisUFOIntegration) GetHTTPServer() *RPCHTTPServer {
	return i.HTTPServer
}
