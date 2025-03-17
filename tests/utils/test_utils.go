package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BinaryType represents the type of binary to use for tests
type BinaryType string

const (
	// BinaryTypeUfo represents the UFO binary
	BinaryTypeUfo BinaryType = "ufo"

	// BinaryTypeFauxmosisUfo represents the Fauxmosis UFO binary
	BinaryTypeFauxmosisUfo BinaryType = "fauxmosis-ufo"

	// BinaryTypeOsmosisUfo represents the Osmosis UFO binary
	BinaryTypeOsmosisUfo BinaryType = "osmosis-ufo"

	// BinaryTypeFauxmosisComet represents the Fauxmosis Comet binary
	BinaryTypeFauxmosisComet BinaryType = "fauxmosis-comet"

	// BinaryTypeOsmosisUfoBridged represents the Osmosis UFO bridged binary
	BinaryTypeOsmosisUfoBridged BinaryType = "osmosis-ufo-bridged"

	// BinaryTypeOsmosisUfoPatched represents the Osmosis UFO patched binary
	BinaryTypeOsmosisUfoPatched BinaryType = "osmosis-ufo-patched"
)

// TestConfig represents the configuration for a test node
type TestConfig struct {
	// HomeDir is the home directory for the node
	HomeDir string

	// RESTAddress is the address for the REST server
	RESTAddress string

	// RPCAddress is the address for the RPC server
	RPCAddress string

	// GRPCAddress is the address for the gRPC server
	GRPCAddress string

	// P2PAddress is the address for the P2P server
	P2PAddress string

	// ChainID is the chain ID for the node
	ChainID string

	// BinaryType is the type of binary to use
	BinaryType BinaryType

	// DebugLevel is the debug level for the node
	DebugLevel string

	// KeysDir is the directory for keys
	KeysDir string

	// BlockTime is the block time for the node
	BlockTime string

	// BlockTimeMS is the block time in milliseconds
	BlockTimeMS int

	// WithCometMock determines whether to use the mock Comet implementation
	WithCometMock bool

	// TrustingPeriod is the trusting period for IBC clients
	TrustingPeriod string

	// ValidatorCount specifies the number of validators for the chain
	ValidatorCount int

	// ValidatorRotationInterval specifies how often (in seconds) the validator set should rotate
	// A value of 0 means no automatic rotation
	ValidatorRotationInterval int

	// EpochLength specifies the number of blocks for each epoch (for Osmosis-based chains)
	// Used to configure epochs for testing IBC client updates across epoch boundaries
	EpochLength int
}

// TestNode represents a test node
type TestNode struct {
	// Config is the configuration for the node
	Config TestConfig

	// Cmd is the command for the node
	Cmd *exec.Cmd

	// Client is the HTTP client for the node
	Client *HTTPClient

	// Cancel is the cancel function for the node context
	Cancel context.CancelFunc
}

// Stop stops the test node
func (n *TestNode) Stop() {
	if n.Cancel != nil {
		n.Cancel()
	}
	if n.Cmd != nil && n.Cmd.Process != nil {
		n.Cmd.Process.Kill()
	}
}

// RotateValidators rotates the validator set by temporarily stopping one validator,
// waiting for consensus to adjust, and then starting it again.
// This simulates validator set changes at regular intervals for testing.
func (n *TestNode) RotateValidators(ctx context.Context) error {
	if n.Config.ValidatorCount < 2 {
		return fmt.Errorf("cannot rotate validators with less than 2 validators")
	}

	client := n.Client
	if client == nil {
		return fmt.Errorf("node HTTP client is not available")
	}

	// Get current validators
	validators, err := client.GetValidators(ctx)
	if err != nil {
		return fmt.Errorf("failed to get validators: %w", err)
	}

	if len(validators) < 2 {
		return fmt.Errorf("need at least 2 active validators for rotation, found %d", len(validators))
	}

	// Select a random validator to rotate (not always the same one)
	// This ensures more interesting validator set changes
	validatorIndex := time.Now().UnixNano() % int64(len(validators))
	validatorMap := validators[validatorIndex]

	// Extract the validator address from the map
	validatorAddr, ok := validatorMap["operator_address"].(string)
	if !ok {
		validatorAddr = fmt.Sprintf("validator-%d", validatorIndex) // Fallback
	}

	// For faster and more frequent validator set changes, we rotate more aggressively
	fmt.Printf("🔄 ROTATING VALIDATOR: %s on chain %s (using index %d)\n", validatorAddr, n.Config.ChainID, validatorIndex)

	// For this test implementation, we'll simulate the effect by manipulating
	// the node directly through its RPC/REST API
	err = client.SimulateValidatorSetChange(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to simulate validator set change: %w", err)
	}

	// No need to wait here - the rotation is simulated and will take effect
	// in the next consensus round
	return nil
}

// StartValidatorRotation starts a goroutine that periodically rotates validators
// according to the ValidatorRotationInterval in the node's config.
// Returns a function that can be called to stop the rotation.
func (n *TestNode) StartValidatorRotation(ctx context.Context) (func(), error) {
	if n.Config.ValidatorRotationInterval <= 0 {
		return func() {}, nil // No rotation needed
	}

	// Use millisecond precision for faster validator rotation
	rotationInterval := time.Duration(n.Config.ValidatorRotationInterval) * time.Millisecond
	rotationCtx, cancelRotation := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(rotationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := n.RotateValidators(ctx); err != nil {
					fmt.Printf("Error rotating validators: %v\n", err)
				}
			case <-rotationCtx.Done():
				return
			}
		}
	}()

	return cancelRotation, nil
}

// SetupTestNode sets up a test node with the given configuration
func SetupTestNode(ctx context.Context, config TestConfig) (*TestNode, error) {
	// Get the binary path for the configured binary type
	binaryPath, err := GetBinaryPath(config.BinaryType)
	if err != nil {
		return nil, fmt.Errorf("failed to get binary path: %w", err)
	}

	// Ensure directories exist
	os.MkdirAll(config.HomeDir, 0755)
	if config.KeysDir != "" {
		os.MkdirAll(config.KeysDir, 0755)
	}

	// Initialize the node
	args := []string{
		"init",
		"--home", config.HomeDir,
		"--chain-id", config.ChainID,
	}

	if config.DebugLevel != "" {
		args = append(args, "--log_level", config.DebugLevel)
	}

	if config.WithCometMock {
		args = append(args, "--with-comet-mock")
	}

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize node: %w, output: %s", err, string(output))
	}

	// Modify genesis.json for epoch configuration if needed
	if config.EpochLength > 0 && (config.BinaryType == BinaryTypeOsmosisUfoPatched || config.BinaryType == BinaryTypeOsmosisUfoBridged) {
		genesisPath := filepath.Join(config.HomeDir, "config", "genesis.json")
		err = configureEpochsInGenesis(genesisPath, config.EpochLength)
		if err != nil {
			fmt.Printf("Warning: failed to configure epochs in genesis.json: %v\n", err)
		}
	}

	// Handle multi-validator setup
	if config.ValidatorCount > 1 {
		// Create additional validators
		for i := 1; i < config.ValidatorCount; i++ {
			validatorID := fmt.Sprintf("validator%d", i+1)
			validatorDir := filepath.Join(config.HomeDir, validatorID)
			os.MkdirAll(validatorDir, 0755)

			// Initialize the validator
			validatorArgs := []string{
				"init",
				"--home", validatorDir,
				"--chain-id", config.ChainID,
			}

			if config.DebugLevel != "" {
				validatorArgs = append(validatorArgs, "--log_level", config.DebugLevel)
			}

			validatorCmd := exec.CommandContext(ctx, binaryPath, validatorArgs...)
			validatorOutput, err := validatorCmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize validator %s: %w, output: %s", validatorID, err, string(validatorOutput))
			}

			// Generate validator key
			genValidatorArgs := []string{
				"keys", "add", validatorID,
				"--keyring-backend", "test",
				"--home", validatorDir,
			}
			genValidatorCmd := exec.CommandContext(ctx, binaryPath, genValidatorArgs...)
			genValidatorOutput, err := genValidatorCmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to generate validator key for %s: %w, output: %s", validatorID, err, string(genValidatorOutput))
			}

			// Add the validator to genesis
			addValidatorArgs := []string{
				"add-genesis-account", validatorID,
				"1000000000stake",
				"--keyring-backend", "test",
				"--home", validatorDir,
			}
			addValidatorCmd := exec.CommandContext(ctx, binaryPath, addValidatorArgs...)
			addValidatorOutput, err := addValidatorCmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to add validator account for %s: %w, output: %s", validatorID, err, string(addValidatorOutput))
			}

			// Create validator transaction
			createValidatorArgs := []string{
				"gentx", validatorID,
				"100000000stake",
				"--chain-id", config.ChainID,
				"--keyring-backend", "test",
				"--home", validatorDir,
			}
			createValidatorCmd := exec.CommandContext(ctx, binaryPath, createValidatorArgs...)
			createValidatorOutput, err := createValidatorCmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to create gentx for %s: %w, output: %s", validatorID, err, string(createValidatorOutput))
			}

			// Copy gentx to main node
			gentxPath := filepath.Join(validatorDir, "config", "gentx")
			mainGentxPath := filepath.Join(config.HomeDir, "config", "gentx")
			os.MkdirAll(mainGentxPath, 0755)

			gentxFiles, err := os.ReadDir(gentxPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read gentx directory for %s: %w", validatorID, err)
			}

			for _, file := range gentxFiles {
				if file.IsDir() {
					continue
				}
				sourceFile := filepath.Join(gentxPath, file.Name())
				destFile := filepath.Join(mainGentxPath, file.Name())
				sourceData, err := os.ReadFile(sourceFile)
				if err != nil {
					return nil, fmt.Errorf("failed to read gentx file %s: %w", sourceFile, err)
				}
				err = os.WriteFile(destFile, sourceData, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to write gentx file %s: %w", destFile, err)
				}
			}
		}

		// Collect all gentxs
		collectGentxArgs := []string{
			"collect-gentxs",
			"--home", config.HomeDir,
		}
		collectGentxCmd := exec.CommandContext(ctx, binaryPath, collectGentxArgs...)
		collectGentxOutput, err := collectGentxCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to collect gentxs: %w, output: %s", err, string(collectGentxOutput))
		}
	}

	// Create node context
	nodeCtx, cancel := context.WithCancel(ctx)

	// Start the node
	startArgs := []string{
		"start",
		"--home", config.HomeDir,
		"--rpc.laddr", config.RPCAddress,
		"--grpc.address", config.GRPCAddress,
	}

	if config.DebugLevel != "" {
		startArgs = append(startArgs, "--log_level", config.DebugLevel)
	}

	if config.WithCometMock {
		startArgs = append(startArgs, "--with-comet-mock")
	}

	if config.BlockTime != "" {
		startArgs = append(startArgs, "--block-time", config.BlockTime)
	} else if config.BlockTimeMS > 0 {
		startArgs = append(startArgs, "--block-time", fmt.Sprintf("%dms", config.BlockTimeMS))
	}

	// Add epoch length configuration if specified
	if config.EpochLength > 0 {
		// Use the proper Osmosis epoch configuration flag
		startArgs = append(startArgs, "--epoch-blocks", fmt.Sprintf("%d", config.EpochLength))
		// Keep the generic flag as well for compatibility
		startArgs = append(startArgs, "--epoch-length", fmt.Sprintf("%d", config.EpochLength))

		// Also create/modify app.toml with the epoch settings
		appConfigPath := filepath.Join(config.HomeDir, "config", "app.toml")
		err := configureEpochsInAppConfig(appConfigPath, config.EpochLength)
		if err != nil {
			fmt.Printf("Warning: failed to configure epochs in app.toml: %v\n", err)
		}
	}

	startCmd := exec.CommandContext(nodeCtx, binaryPath, startArgs...)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	err = startCmd.Start()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start node: %w", err)
	}

	// Create HTTP client
	client := NewHTTPClient(config.RESTAddress)

	// Return the test node
	return &TestNode{
		Config: config,
		Cmd:    startCmd,
		Client: client,
		Cancel: cancel,
	}, nil
}

// CleanupTestNode cleans up a test node with the given configuration
func CleanupTestNode(ctx context.Context, config TestConfig) error {
	// Get the binary path for the configured binary type
	binaryPath, err := GetBinaryPath(config.BinaryType)
	if err != nil {
		return fmt.Errorf("failed to get binary path: %w", err)
	}

	// Stop the node
	stopCmd := exec.CommandContext(ctx, binaryPath, "stop", "--home", config.HomeDir)
	if err := stopCmd.Run(); err != nil {
		fmt.Printf("Warning: failed to stop node: %v\n", err)
	}

	// Remove the home directory
	if err := os.RemoveAll(config.HomeDir); err != nil {
		return fmt.Errorf("failed to remove home directory: %w", err)
	}

	return nil
}

// StartTestNode starts a test node with the given configuration
func StartTestNode(ctx context.Context, config TestConfig) (*TestNode, error) {
	// Print more diagnostic information
	fmt.Printf("Starting test node with configuration:\n")
	fmt.Printf("  Chain ID: %s\n", config.ChainID)
	fmt.Printf("  Binary Type: %s\n", config.BinaryType)
	fmt.Printf("  HomeDir: %s\n", config.HomeDir)
	fmt.Printf("  RPC Address: %s\n", config.RPCAddress)
	fmt.Printf("  Epoch Length: %d\n", config.EpochLength)
	fmt.Printf("  Validator Count: %d\n", config.ValidatorCount)

	// Check if the binary exists
	binaryPath, err := GetBinaryPath(config.BinaryType)
	if err != nil {
		return nil, fmt.Errorf("binary for %s not found: %w", config.BinaryType, err)
	}
	fmt.Printf("  Using binary at: %s\n", binaryPath)

	// Set up the node
	node, err := SetupTestNode(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to set up node: %w", err)
	}

	// Wait for the node to start producing blocks
	waitTime := 5 * time.Second
	fmt.Printf("  Waiting %v for node to start producing blocks...\n", waitTime)
	time.Sleep(waitTime)

	// Verify the node is running by checking its status
	fmt.Printf("  Verifying node is running...\n")
	_, err = node.Client.GetNodeStatus(ctx)
	if err != nil {
		node.Stop()
		return nil, fmt.Errorf("failed to verify node is running: %w", err)
	}
	fmt.Printf("  Node started and verified successfully\n")

	return node, nil
}

// GetBinaryPath returns the path to the binary for the specified type
func GetBinaryPath(binaryType BinaryType) (string, error) {
	// First, check if the binary is available in the PATH
	path, err := exec.LookPath(string(binaryType))
	if err == nil {
		return path, nil
	}

	// Try to find the binary in common locations
	// First, get the project root
	projectRoot := getProjectRoot()

	// Define possible locations
	possibleLocations := []string{
		filepath.Join(projectRoot, "bin", string(binaryType)),
		filepath.Join(projectRoot, "build", string(binaryType)),
		filepath.Join(projectRoot, string(binaryType)),
		filepath.Join(projectRoot, "cmd", string(binaryType), string(binaryType)),
		filepath.Join(projectRoot, "tests", "bin", string(binaryType)),
	}

	// Check each location
	for _, location := range possibleLocations {
		if _, err := os.Stat(location); err == nil {
			return location, nil
		}
	}

	// If binary type contains a hyphen, try to split and check each part
	if strings.Contains(string(binaryType), "-") {
		parts := strings.Split(string(binaryType), "-")
		for _, part := range parts {
			for _, location := range possibleLocations {
				modLocation := strings.Replace(location, string(binaryType), part, 1)
				if _, err := os.Stat(modLocation); err == nil {
					return modLocation, nil
				}
			}

			// Also try just the part in PATH
			path, err := exec.LookPath(part)
			if err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("binary not found: %s", binaryType)
}

// getProjectRoot attempts to find the root directory of the project
func getProjectRoot() string {
	// First try to use git to find the repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	// If git command fails, try to use the working directory
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	// Check if we're in a subdirectory of the project
	for {
		// Check if this directory contains tests/bin or bin
		if _, err := os.Stat(filepath.Join(dir, "tests", "bin")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "bin")); err == nil {
			return dir
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// If all else fails, return the current directory
	return "."
}

// TestChain represents a test chain for IBC testing
type TestChain struct {
	ID             string
	Name           string
	BinaryName     string
	HomeDir        string
	ConfigDir      string
	RPCAddress     string
	P2PAddress     string
	APIAddress     string
	GRPCAddress    string
	ValidatorKey   string
	AccountKey     string
	AccountAddress string
	Command        *exec.Cmd
	HTTPClient     *HTTPClient
}

// setBashCommand executes a bash command
func setBashCommand(cmd string) error {
	c := exec.Command("bash", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// NewCustomCommand creates a new custom command for test purposes
func NewCustomCommand(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	return cmd
}

// SetupChain initializes a chain for testing
func SetupChain(ctx context.Context, id, name, binaryName, homeDir string) (*TestChain, error) {
	// Create test chain
	chain := &TestChain{
		ID:          id,
		Name:        name,
		BinaryName:  binaryName,
		HomeDir:     homeDir,
		ConfigDir:   filepath.Join(homeDir, "config"),
		RPCAddress:  fmt.Sprintf("tcp://127.0.0.1:%d", 26656+getChainIDOffset(id)),
		P2PAddress:  fmt.Sprintf("tcp://127.0.0.1:%d", 26658+getChainIDOffset(id)),
		APIAddress:  fmt.Sprintf("tcp://127.0.0.1:%d", 1317+getChainIDOffset(id)),
		GRPCAddress: fmt.Sprintf("127.0.0.1:%d", 9090+getChainIDOffset(id)),
	}

	// Create HTTP client
	chain.HTTPClient = NewHTTPClient(fmt.Sprintf("http://127.0.0.1:%d", 1317+getChainIDOffset(id)))

	// Create chain home directory
	if err := os.MkdirAll(chain.HomeDir, 0755); err != nil {
		return nil, err
	}

	// Initialize chain
	if err := chain.Init(ctx); err != nil {
		return nil, err
	}

	return chain, nil
}

// Init initializes the chain
func (c *TestChain) Init(ctx context.Context) error {
	// Initialize chain
	initCmd := exec.CommandContext(ctx, c.BinaryName, "init", c.Name, "--chain-id", c.ID, "--home", c.HomeDir)
	output, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize chain: %w, output: %s", err, output)
	}

	// Create validator key
	addValidatorKeyCmd := exec.CommandContext(ctx, c.BinaryName, "keys", "add", "validator", "--keyring-backend", "test", "--home", c.HomeDir)
	output, err = addValidatorKeyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add validator key: %w, output: %s", err, output)
	}
	c.ValidatorKey = "validator"

	// Create account key
	addAccountKeyCmd := exec.CommandContext(ctx, c.BinaryName, "keys", "add", "account", "--keyring-backend", "test", "--home", c.HomeDir)
	output, err = addAccountKeyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add account key: %w, output: %s", err, output)
	}
	c.AccountKey = "account"

	// Get account address
	getAddrCmd := exec.CommandContext(ctx, c.BinaryName, "keys", "show", "account", "-a", "--keyring-backend", "test", "--home", c.HomeDir)
	output, err = getAddrCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get account address: %w, output: %s", err, output)
	}
	c.AccountAddress = string(output)

	// Add genesis account for validator
	addValidatorGenesisAccountCmd := exec.CommandContext(ctx,
		c.BinaryName, "add-genesis-account", "validator", "1000000000000stake",
		"--keyring-backend", "test", "--home", c.HomeDir)
	output, err = addValidatorGenesisAccountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add validator genesis account: %w, output: %s", err, output)
	}

	// Add genesis account for account
	addAccountGenesisAccountCmd := exec.CommandContext(ctx,
		c.BinaryName, "add-genesis-account", "account", "1000000000000stake",
		"--keyring-backend", "test", "--home", c.HomeDir)
	output, err = addAccountGenesisAccountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add account genesis account: %w, output: %s", err, output)
	}

	// Create validator gentx
	createGentxCmd := exec.CommandContext(ctx,
		c.BinaryName, "gentx", "validator", "100000000stake",
		"--chain-id", c.ID, "--keyring-backend", "test", "--home", c.HomeDir)
	output, err = createGentxCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create gentx: %w, output: %s", err, output)
	}

	// Collect gentxs
	collectGentxsCmd := exec.CommandContext(ctx, c.BinaryName, "collect-gentxs", "--home", c.HomeDir)
	output, err = collectGentxsCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to collect gentxs: %w, output: %s", err, output)
	}

	return nil
}

// Start starts the chain
func (c *TestChain) Start(ctx context.Context) error {
	args := []string{
		"start",
		"--home", c.HomeDir,
		"--rpc.laddr", c.RPCAddress,
		"--p2p.laddr", c.P2PAddress,
		"--grpc.address", c.GRPCAddress,
		"--grpc-web.enable=false",
		"--api.enable",
		"--api.address", c.APIAddress,
	}

	cmd := exec.CommandContext(ctx, c.BinaryName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start chain: %w", err)
	}

	c.Command = cmd

	// Wait for the chain to start
	time.Sleep(5 * time.Second)

	return nil
}

// Stop stops the chain
func (c *TestChain) Stop() error {
	if c.Command == nil || c.Command.Process == nil {
		return nil
	}
	return c.Command.Process.Kill()
}

// CleanupChain cleans up the chain's home directory
func CleanupChain(homeDir string) error {
	if homeDir == "" {
		return nil
	}
	return os.RemoveAll(homeDir)
}

// getChainIDOffset returns an offset based on the chain ID to differentiate ports
func getChainIDOffset(id string) int {
	switch id {
	case "chain-a":
		return 0
	case "chain-b":
		return 100
	case "chain-c":
		return 200
	default:
		return 300
	}
}

// WaitForChainStart waits for a chain to start by polling its status endpoint
func WaitForChainStart(ctx context.Context, httpClient *HTTPClient, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for chain to start")
		case <-ticker.C:
			_, err := httpClient.GetNodeStatus(ctx)
			if err == nil {
				return nil
			}
		}
	}
}

// WaitForBlockHeight waits for the chain to reach a specific block height
func WaitForBlockHeight(ctx context.Context, httpClient *HTTPClient, height int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for block height %d", height)
		case <-ticker.C:
			status, err := httpClient.GetNodeStatus(ctx)
			if err != nil {
				continue
			}

			syncInfo, ok := status["sync_info"].(map[string]interface{})
			if !ok || syncInfo == nil {
				continue
			}

			latestHeight, ok := syncInfo["latest_block_height"].(string)
			if !ok {
				continue
			}

			var currentHeight int
			if _, err := fmt.Sscanf(latestHeight, "%d", &currentHeight); err != nil {
				continue
			}

			if currentHeight >= height {
				return nil
			}
		}
	}
}

// SetupIBCChains sets up two chains with IBC connection and channel
func SetupIBCChains(ctx context.Context, binaryA, binaryB, hermesConfigPath string) (*TestChain, *TestChain, *HermesConfig, error) {
	// Setup Chain A
	chainA, err := SetupChain(ctx, "chain-a", "Chain A", binaryA, "testdata/chain-a")
	if err != nil {
		return nil, nil, nil, err
	}

	// Setup Chain B
	chainB, err := SetupChain(ctx, "chain-b", "Chain B", binaryB, "testdata/chain-b")
	if err != nil {
		return nil, nil, nil, err
	}

	// Start chains
	if err := chainA.Start(ctx); err != nil {
		return nil, nil, nil, err
	}

	if err := chainB.Start(ctx); err != nil {
		return nil, nil, nil, err
	}

	// Wait for chains to start
	if err := WaitForChainStart(ctx, chainA.HTTPClient, 30*time.Second); err != nil {
		return nil, nil, nil, err
	}

	if err := WaitForChainStart(ctx, chainB.HTTPClient, 30*time.Second); err != nil {
		return nil, nil, nil, err
	}

	// Setup Hermes
	hermes := NewHermesConfig(hermesConfigPath, "")

	return chainA, chainB, hermes, nil
}

// SetupIBCConnection sets up IBC connection between two chains
func SetupIBCConnection(ctx context.Context, hermes *HermesConfig, chainA, chainB *TestChain) error {
	// Create clients
	if err := hermes.CreateClient(ctx, chainA.ID, chainB.ID); err != nil {
		return err
	}

	if err := hermes.CreateClient(ctx, chainB.ID, chainA.ID); err != nil {
		return err
	}

	// Get clients
	clientsA, err := hermes.GetClients(ctx, chainA.ID)
	if err != nil || len(clientsA) == 0 {
		return fmt.Errorf("failed to get clients for chain A: %w", err)
	}

	clientsB, err := hermes.GetClients(ctx, chainB.ID)
	if err != nil || len(clientsB) == 0 {
		return fmt.Errorf("failed to get clients for chain B: %w", err)
	}

	// Create connection
	if err := hermes.CreateConnection(ctx, chainA.ID, chainB.ID, clientsA[0], clientsB[0]); err != nil {
		return err
	}

	// Wait for connection to be established
	time.Sleep(5 * time.Second)

	return nil
}

// SetupIBCChannel sets up IBC channel between two chains
func SetupIBCChannel(ctx context.Context, hermes *HermesConfig, chainA, chainB *TestChain) error {
	// Get connections
	connectionsA, err := hermes.GetConnections(ctx, chainA.ID)
	if err != nil || len(connectionsA) == 0 {
		return fmt.Errorf("failed to get connections for chain A: %w", err)
	}

	// Create channel
	if err := hermes.CreateChannel(ctx, chainA.ID, chainB.ID, connectionsA[0], "transfer", "transfer", false, ""); err != nil {
		return err
	}

	// Wait for channel to be established
	time.Sleep(5 * time.Second)

	return nil
}

// SetupRelayer starts the Hermes relayer for the two chains
func SetupRelayer(ctx context.Context, hermes *HermesConfig, chainA, chainB *TestChain) (*exec.Cmd, error) {
	return hermes.StartRelayer(ctx, []string{chainA.ID, chainB.ID})
}

// PerformIBCTransfer performs an IBC token transfer from chainA to chainB
func PerformIBCTransfer(ctx context.Context, hermes *HermesConfig, chainA, chainB *TestChain, amount, denom string) error {
	// Get channels
	channelsA, err := hermes.GetChannels(ctx, chainA.ID)
	if err != nil || len(channelsA) == 0 {
		return fmt.Errorf("failed to get channels for chain A: %w", err)
	}

	// Perform transfer
	return hermes.TransferTokens(ctx, chainA.ID, chainB.ID, "transfer", channelsA[0],
		chainB.AccountAddress, amount, denom, 1000)
}

// VerifyIBCTransfer verifies that an IBC token transfer was successful
func VerifyIBCTransfer(ctx context.Context, chainB *TestChain, amount string, sourceChainID string, denom string) error {
	// Wait for the transfer to be processed
	time.Sleep(10 * time.Second)

	// Construct IBC denom
	ibcDenom := fmt.Sprintf("ibc/HASH-%s-%s", sourceChainID, denom)

	// Get balance
	balance, err := chainB.HTTPClient.GetBalance(ctx, chainB.AccountAddress, ibcDenom)
	if err != nil {
		return err
	}

	if balance != amount {
		return fmt.Errorf("expected balance %s, got %s", amount, balance)
	}

	return nil
}

// CleanupTestEnvironment cleans up the test environment
func CleanupTestEnvironment(chainA, chainB *TestChain, relayerCmd *exec.Cmd) {
	if relayerCmd != nil && relayerCmd.Process != nil {
		relayerCmd.Process.Kill()
	}

	if chainA != nil {
		chainA.Stop()
		CleanupChain(chainA.HomeDir)
	}

	if chainB != nil {
		chainB.Stop()
		CleanupChain(chainB.HomeDir)
	}
}

// configureEpochsInAppConfig creates or modifies the app.toml file to configure epochs
func configureEpochsInAppConfig(configPath string, epochLength int) error {
	// Check if the file exists
	var content []byte
	var err error
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create a new file with default content
		content = []byte("[epochs]\n")
	} else {
		// Read existing file
		content, err = os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read app.toml: %w", err)
		}
	}

	// Check if the epochs section exists
	epochsSection := "[epochs]"
	if !strings.Contains(string(content), epochsSection) {
		content = append(content, []byte(epochsSection+"\n")...)
	}

	// Add or update epoch settings
	lines := strings.Split(string(content), "\n")
	epochsFound := false
	epochsConfigured := false
	for i, line := range lines {
		if line == epochsSection {
			epochsFound = true
		}
		if epochsFound && strings.Contains(line, "epoch_length") {
			lines[i] = fmt.Sprintf("epoch_length = %d", epochLength)
			epochsConfigured = true
		}
	}

	// If epoch_length wasn't found, add it
	if !epochsConfigured {
		// Find the end of the epochs section
		for i, line := range lines {
			if epochsFound && (i == len(lines)-1 || (len(line) > 0 && line[0] == '[' && line != epochsSection)) {
				// Insert epoch_length before this section or at the end
				newLines := append(
					lines[:i],
					fmt.Sprintf("epoch_length = %d", epochLength),
				)
				if i < len(lines) {
					newLines = append(newLines, lines[i:]...)
				}
				lines = newLines
				break
			}
		}
	}

	// Add day and week epochs if they don't exist
	epochsLines := []string{
		fmt.Sprintf("epoch_length = %d", epochLength),
		fmt.Sprintf("day_epoch_blocks = %d", epochLength),
		fmt.Sprintf("week_epoch_blocks = %d", epochLength*7),
	}

	// Find the epochs section and update it
	epochsFound = false
	epochsSectionIndex := -1
	for i, line := range lines {
		if line == epochsSection {
			epochsFound = true
			epochsSectionIndex = i
			break
		}
	}

	if epochsFound {
		// Replace everything after the epochs section until the next section
		nextSectionIndex := len(lines)
		for i := epochsSectionIndex + 1; i < len(lines); i++ {
			if len(lines[i]) > 0 && lines[i][0] == '[' {
				nextSectionIndex = i
				break
			}
		}

		newLines := append(
			lines[:epochsSectionIndex+1],
			epochsLines...,
		)
		if nextSectionIndex < len(lines) {
			newLines = append(newLines, lines[nextSectionIndex:]...)
		}
		lines = newLines
	} else {
		// Append the epochs section at the end
		lines = append(lines, epochsSection)
		lines = append(lines, epochsLines...)
	}

	// Write the updated content back
	err = os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write app.toml: %w", err)
	}

	fmt.Printf("Successfully configured epochs in %s with epoch_length=%d\n", configPath, epochLength)
	return nil
}

// configureEpochsInGenesis modifies the genesis.json file to configure the Osmosis epoch module
func configureEpochsInGenesis(genesisPath string, epochLength int) error {
	// Read the genesis file
	content, err := os.ReadFile(genesisPath)
	if err != nil {
		return fmt.Errorf("failed to read genesis.json: %w", err)
	}

	// Parse the JSON
	var genesis map[string]interface{}
	if err := json.Unmarshal(content, &genesis); err != nil {
		return fmt.Errorf("failed to parse genesis.json: %w", err)
	}

	// Get or create app_state
	appState, ok := genesis["app_state"].(map[string]interface{})
	if !ok {
		appState = make(map[string]interface{})
		genesis["app_state"] = appState
	}

	// Configure the epochs module
	epochsConfig := map[string]interface{}{
		"epochs": []map[string]interface{}{
			{
				"identifier":                 "minute",
				"start_time":                 "0001-01-01T00:00:00Z",
				"duration":                   fmt.Sprintf("%ds", epochLength*1), // Each block is ~1s
				"current_epoch":              float64(0),
				"current_epoch_start_time":   "0001-01-01T00:00:00Z",
				"epoch_counting_started":     false,
				"current_epoch_start_height": float64(0),
			},
			{
				"identifier":                 "hour",
				"start_time":                 "0001-01-01T00:00:00Z",
				"duration":                   fmt.Sprintf("%ds", epochLength*60), // 60 minutes
				"current_epoch":              float64(0),
				"current_epoch_start_time":   "0001-01-01T00:00:00Z",
				"epoch_counting_started":     false,
				"current_epoch_start_height": float64(0),
			},
			{
				"identifier":                 "day",
				"start_time":                 "0001-01-01T00:00:00Z",
				"duration":                   fmt.Sprintf("%ds", epochLength*60*24), // 24 hours
				"current_epoch":              float64(0),
				"current_epoch_start_time":   "0001-01-01T00:00:00Z",
				"epoch_counting_started":     false,
				"current_epoch_start_height": float64(0),
			},
			{
				"identifier":                 "week",
				"start_time":                 "0001-01-01T00:00:00Z",
				"duration":                   fmt.Sprintf("%ds", epochLength*60*24*7), // 7 days
				"current_epoch":              float64(0),
				"current_epoch_start_time":   "0001-01-01T00:00:00Z",
				"epoch_counting_started":     false,
				"current_epoch_start_height": float64(0),
			},
		},
	}

	// Update the app_state with the epochs configuration
	appState["epochs"] = epochsConfig

	// Write the updated genesis back to the file
	updatedContent, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated genesis.json: %w", err)
	}

	err = os.WriteFile(genesisPath, updatedContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated genesis.json: %w", err)
	}

	fmt.Printf("Successfully configured epochs in genesis.json with epoch_length=%d blocks\n", epochLength)
	return nil
}
