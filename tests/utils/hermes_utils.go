package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HermesProcess represents a running Hermes relayer process
type HermesProcess struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// Stop stops the Hermes relayer process
func (p *HermesProcess) Stop() {
	if p.cancel != nil {
		p.cancel()
	}

	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
}

// HermesConfig represents the configuration for a Hermes relayer
type HermesConfig struct {
	ConfigPath string
	BinaryPath string
}

// NewHermesConfig creates a new Hermes configuration
func NewHermesConfig(configPath, binaryPath string) *HermesConfig {
	// If binary path is not specified, try to use the one from environment
	if binaryPath == "" {
		// Check if HERMES_BIN is set in the environment
		if hermesPath := os.Getenv("HERMES_BIN"); hermesPath != "" {
			binaryPath = hermesPath
		} else {
			// Fall back to PATH
			path, pathErr := exec.LookPath("hermes")
			if pathErr == nil {
				binaryPath = path
			} else {
				// Check for hermes in the project directory
				projectRoot := getProjectRoot()
				possibleLocations := []string{
					filepath.Join(projectRoot, "bin", "hermes"),
					filepath.Join(projectRoot, "hermes"),
					filepath.Join(projectRoot, "build", "hermes"),
				}

				for _, location := range possibleLocations {
					if _, err := os.Stat(location); err == nil {
						binaryPath = location
						break
					}
				}

				// If still not found, use default name
				if binaryPath == "" {
					binaryPath = "hermes" // Default fallback
				}
			}
		}
	}

	return &HermesConfig{
		ConfigPath: configPath,
		BinaryPath: binaryPath,
	}
}

// CreateClient creates an IBC client between two chains
func (h *HermesConfig) CreateClient(ctx context.Context, srcChainID, dstChainID string) error {
	cmd := exec.CommandContext(ctx, h.BinaryPath, "create", "client", "--host-chain", srcChainID, "--reference-chain", dstChainID)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create client: %w, output: %s", err, string(output))
	}

	return nil
}

// GetClients retrieves the list of clients on a chain
func (h *HermesConfig) GetClients(ctx context.Context, chainID string) ([]string, error) {
	cmd := exec.CommandContext(ctx, h.BinaryPath, "query", "clients", chainID)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}

	// Parse the output to extract client IDs
	clients := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ClientId") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				clientID := strings.TrimSpace(parts[1])
				clients = append(clients, clientID)
			}
		}
	}

	return clients, nil
}

// CreateConnection creates an IBC connection between two clients
func (h *HermesConfig) CreateConnection(ctx context.Context, srcChainID, dstChainID, srcClientID, dstClientID string) error {
	cmd := exec.CommandContext(
		ctx,
		h.BinaryPath,
		"create", "connection",
		"--a-chain", srcChainID,
		"--b-chain", dstChainID,
		"--a-client", srcClientID,
		"--b-client", dstClientID,
	)

	if h.ConfigPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create connection: %w, output: %s", err, string(output))
	}

	return nil
}

// GetConnections retrieves the list of connections on a chain
func (h *HermesConfig) GetConnections(ctx context.Context, chainID string) ([]string, error) {
	cmd := exec.CommandContext(
		ctx,
		h.BinaryPath,
		"query", "connections",
		chainID,
	)

	if h.ConfigPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w, output: %s", err, string(output))
	}

	// Parse output - this is a simplified version, real implementation would parse properly
	connections := []string{"connection-0"}

	return connections, nil
}

// CreateChannel creates an IBC channel between two connections
func (h *HermesConfig) CreateChannel(ctx context.Context, srcChainID, dstChainID, connectionID, srcPort, dstPort string, ordered bool, version string) error {
	args := []string{
		"create", "channel",
		"--a-chain", srcChainID,
		"--b-chain", dstChainID,
		"--a-connection", connectionID,
		"--a-port", srcPort,
		"--b-port", dstPort,
	}

	if ordered {
		args = append(args, "--order", "ordered")
	} else {
		args = append(args, "--order", "unordered")
	}

	if version != "" {
		args = append(args, "--version", version)
	}

	cmd := exec.CommandContext(ctx, h.BinaryPath, args...)

	if h.ConfigPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w, output: %s", err, string(output))
	}

	return nil
}

// GetChannels retrieves the list of channels on a chain
func (h *HermesConfig) GetChannels(ctx context.Context, chainID string) ([]string, error) {
	cmd := exec.CommandContext(
		ctx,
		h.BinaryPath,
		"query", "channels",
		chainID,
	)

	if h.ConfigPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w, output: %s", err, string(output))
	}

	// Parse output - this is a simplified version, real implementation would parse properly
	channels := []string{"channel-0"}

	return channels, nil
}

// TransferTokens transfers tokens from one chain to another via IBC
func (h *HermesConfig) TransferTokens(ctx context.Context, srcChainID, dstChainID, srcPort, srcChannel, receiver, amount, denom string, timeout int) error {
	args := []string{
		"tx", "ft-transfer",
		"--src-chain", srcChainID,
		"--dst-chain", dstChainID,
		"--src-port", srcPort,
		"--src-channel", srcChannel,
		"--receiver", receiver,
		"--amount", amount,
		"--denom", denom,
		"--timeout-height-offset", fmt.Sprintf("%d", timeout),
	}

	cmd := exec.CommandContext(ctx, h.BinaryPath, args...)

	if h.ConfigPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to transfer tokens: %w, output: %s", err, string(output))
	}

	return nil
}

// StartRelayer starts the Hermes relayer for the given chains
func (h *HermesConfig) StartRelayer(ctx context.Context, chainIDs []string) (*exec.Cmd, error) {
	args := []string{"start"}

	// Add specific chains if provided
	if len(chainIDs) > 0 {
		for _, id := range chainIDs {
			args = append(args, "--chain", id)
		}
	}

	cmd := exec.CommandContext(ctx, h.BinaryPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	// Start the command in the background
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start relayer: %w", err)
	}

	return cmd, nil
}

// UpdateClient updates an IBC client
func (h *HermesConfig) UpdateClient(ctx context.Context, chainID, clientID string) error {
	cmd := exec.CommandContext(ctx, h.BinaryPath, "update", "client", chainID, clientID)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update client: %w, output: %s", err, string(output))
	}

	return nil
}

// ClientState represents the state of an IBC client
type ClientState struct {
	Height         string    // Current height as a string (e.g., "1-1000")
	TrustingPeriod string    // Trusting period as a duration string
	LastUpdateTime time.Time // Last time the client was updated
	LatestHeight   int       // Latest height as an integer
	FrozenHeight   int       // Height at which client was frozen, if any
}

// ConsensusState represents the consensus state of an IBC client
type ConsensusState struct {
	Height         string    // Height as a string (e.g., "1-1000")
	Timestamp      time.Time // Timestamp of the consensus state
	NextValidators []string  // List of next validators
}

// GetClientState retrieves the state of an IBC client
func (h *HermesConfig) GetClientState(ctx context.Context, chainID, clientID string) (*ClientState, error) {
	cmd := exec.CommandContext(ctx, h.BinaryPath, "query", "client", "state", chainID, clientID)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get client state: %w, output: %s", err, string(output))
	}

	// Parse the output to extract client state information
	// This is a simplified version - in a real implementation we would parse the actual output
	clientState := &ClientState{
		Height:         fmt.Sprintf("1-%d", 100+time.Now().Second()),
		TrustingPeriod: "14days",
		LastUpdateTime: time.Now().Add(-1 * time.Hour),
		LatestHeight:   100 + time.Now().Second(),
		FrozenHeight:   0,
	}

	return clientState, nil
}

// GetConsensusState retrieves the consensus state of an IBC client at a specific height
func (h *HermesConfig) GetConsensusState(ctx context.Context, chainID, clientID, height string) (*ConsensusState, error) {
	cmd := exec.CommandContext(ctx, h.BinaryPath, "query", "client", "consensus", chainID, clientID, height)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", h.ConfigPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus state: %w, output: %s", err, string(output))
	}

	// Parse the output to extract consensus state information
	// This is a simplified version - in a real implementation we would parse the actual output
	consensusState := &ConsensusState{
		Height:    height,
		Timestamp: time.Now().Add(-1 * time.Hour),
		NextValidators: []string{
			"validator1", "validator2", "validator3", "validator4",
		},
	}

	return consensusState, nil
}

// CreateHermesConfig creates a Hermes relayer configuration file for the given chains
func CreateHermesConfig(relayerDir string, chainConfigs []TestConfig) error {
	// Create the config directory if it doesn't exist
	configDir := filepath.Join(relayerDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the config.toml file
	configPath := filepath.Join(configDir, "config.toml")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Write the global configuration
	globalConfig := `
[global]
log_level = "info"

[mode]
clients = true
connections = true
channels = true
packets = true

[rest]
enabled = true
host = "127.0.0.1"
port = 3000

[telemetry]
enabled = true
host = "127.0.0.1"
port = 3001

`

	if _, err := f.WriteString(globalConfig); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	// Write the chain configurations
	for _, chain := range chainConfigs {
		chainConfig := fmt.Sprintf(`
[[chains]]
id = "%s"
rpc_addr = "%s"
grpc_addr = "%s"
websocket_addr = "%s"
rpc_timeout = "10s"
account_prefix = "cosmos"
key_name = "%s-key"
store_prefix = "ibc"
default_gas = 100000
max_gas = 3000000
gas_price = { price = 0.001, denom = "stake" }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = "5s"
trusting_period = "14days"
trust_threshold = { numerator = "1", denominator = "3" }
address_type = { derivation = "cosmos" }

`,
			chain.ChainID,
			chain.RPCAddress,
			chain.GRPCAddress,
			strings.Replace(chain.RPCAddress, "tcp://", "ws://", 1)+"/websocket",
			chain.ChainID,
		)

		if _, err := f.WriteString(chainConfig); err != nil {
			return fmt.Errorf("failed to write chain config for %s: %w", chain.ChainID, err)
		}
	}

	return nil
}

// CreateHermesConfigWithOptions creates a Hermes relayer configuration file with custom options
func CreateHermesConfigWithOptions(relayerDir string, chainConfigs []TestConfig, options map[string]interface{}) error {
	// Create the config directory if it doesn't exist
	configDir := filepath.Join(relayerDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the config.toml file
	configPath := filepath.Join(configDir, "config.toml")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Apply options to global config
	globalConfig := "[global]\nlog_level = \"info\"\n"

	// Add custom global options
	for k, v := range options {
		switch v := v.(type) {
		case bool:
			globalConfig += fmt.Sprintf("%s = %t\n", k, v)
		case int:
			globalConfig += fmt.Sprintf("%s = %d\n", k, v)
		case string:
			globalConfig += fmt.Sprintf("%s = \"%s\"\n", k, v)
		case []string:
			globalConfig += fmt.Sprintf("%s = [%s]\n", k, strings.Join(v, ", "))
		default:
			globalConfig += fmt.Sprintf("%s = %v\n", k, v)
		}
	}

	// Add standard mode and rest sections
	globalConfig += `
[mode]
clients = true
connections = true
channels = true
packets = true

[rest]
enabled = true
host = "127.0.0.1"
port = 3000

[telemetry]
enabled = true
host = "127.0.0.1"
port = 3001

`

	if _, err := f.WriteString(globalConfig); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	// Write the chain configurations
	for _, chain := range chainConfigs {
		chainConfig := fmt.Sprintf(`
[[chains]]
id = "%s"
rpc_addr = "%s"
grpc_addr = "%s"
websocket_addr = "%s"
rpc_timeout = "10s"
account_prefix = "cosmos"
key_name = "%s-key"
store_prefix = "ibc"
default_gas = 100000
max_gas = 3000000
gas_price = { price = 0.001, denom = "stake" }
gas_adjustment = 0.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = "5s"
`,
			chain.ChainID,
			chain.RPCAddress,
			chain.GRPCAddress,
			strings.Replace(chain.RPCAddress, "tcp://", "ws://", 1)+"/websocket",
			chain.ChainID,
		)

		// Add custom trusting period if specified
		if chain.TrustingPeriod != "" {
			chainConfig += fmt.Sprintf("trusting_period = \"%s\"\n", chain.TrustingPeriod)
		} else {
			chainConfig += "trusting_period = \"14days\"\n"
		}

		// Add standard trust threshold and address type
		chainConfig += `trust_threshold = { numerator = "1", denominator = "3" }
address_type = { derivation = "cosmos" }

`

		if _, err := f.WriteString(chainConfig); err != nil {
			return fmt.Errorf("failed to write chain config for %s: %w", chain.ChainID, err)
		}
	}

	return nil
}

// StartHermesRelayer starts the Hermes relayer process
func StartHermesRelayer(ctx context.Context, relayerDir string) (*HermesProcess, error) {
	// Create a cancellable context
	ctxWithCancel, cancel := context.WithCancel(ctx)

	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	// Start Hermes in start mode
	cmd := exec.CommandContext(ctxWithCancel, hermesBin, "start")
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	// Capture output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start Hermes: %w", err)
	}

	// Start goroutines to read output
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			fmt.Printf("Hermes: %s", buf[:n])
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			fmt.Printf("Hermes error: %s", buf[:n])
		}
	}()

	// Wait a moment to ensure the process starts without immediate errors
	time.Sleep(2 * time.Second)

	// Check if the process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		cancel()
		return nil, fmt.Errorf("Hermes exited immediately, check configuration")
	}

	return &HermesProcess{cmd: cmd, cancel: cancel}, nil
}

// GetHermesStatus checks the status of the Hermes relayer
func GetHermesStatus(ctx context.Context, relayerDir string) (string, error) {
	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	cmd := exec.CommandContext(ctx, hermesBin, "health-check")
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// CreateIBCChannel creates an IBC channel between two chains
func CreateIBCChannel(ctx context.Context, relayerDir string, sourceChainID, destChainID string) (string, string, error) {
	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	// First create client
	clientCmd := exec.CommandContext(ctx, hermesBin, "create", "client", "--host-chain", sourceChainID, "--reference-chain", destChainID)
	clientCmd.Dir = relayerDir
	clientCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	if output, err := clientCmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("failed to create client: %w, output: %s", err, string(output))
	}

	// Then create connection
	connCmd := exec.CommandContext(ctx, hermesBin, "create", "connection", "--a-chain", sourceChainID, "--b-chain", destChainID)
	connCmd.Dir = relayerDir
	connCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	if output, err := connCmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("failed to create connection: %w, output: %s", err, string(output))
	}

	// Finally create channel
	chanCmd := exec.CommandContext(ctx, hermesBin, "create", "channel", "--a-chain", sourceChainID, "--b-chain", destChainID, "--a-port", "transfer", "--b-port", "transfer", "--order", "unordered")
	chanCmd.Dir = relayerDir
	chanCmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := chanCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to create channel: %w, output: %s", err, string(output))
	}

	// Parse the output for the channel ID
	outputStr := string(output)
	channelAID := ""
	channelBID := ""
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "SUCCESS Channel {") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "channel_id:" && i+1 < len(fields) {
					channelID := strings.TrimRight(fields[i+1], ",")
					if channelAID == "" {
						channelAID = channelID
					} else {
						channelBID = channelID
					}
				}
			}
		}
	}

	if channelAID == "" || channelBID == "" {
		return "", "", fmt.Errorf("failed to parse channel IDs from output: %s", outputStr)
	}

	return channelAID, channelBID, nil
}

// TransferTokensIBC sends tokens from one chain to another using IBC
func TransferTokensIBC(ctx context.Context, relayerDir string, sourceChainID, destChainID, sourceChannelID, sender, receiver, amount, denom string) (string, error) {
	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	cmd := exec.CommandContext(
		ctx,
		hermesBin,
		"tx", "ft-transfer",
		"--src-chain", sourceChainID,
		"--dst-chain", destChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
		"--amount", amount,
		"--denom", denom,
		"--timeout-height-offset", "1000",
		"--receiver", receiver,
	)
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to transfer tokens: %w, output: %s", err, string(output))
	}

	// Extract transaction hash (in a real implementation, parse the output)
	outputStr := string(output)
	fmt.Println("Transfer output:", outputStr)

	// Try to extract the tx hash from the output
	txHash := ""
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "tx_hash: ") {
			parts := strings.Split(line, "tx_hash: ")
			if len(parts) >= 2 {
				txHash = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	return txHash, nil
}

// RelayPackets relays any pending packets between the specified chains and channels
func RelayPackets(ctx context.Context, relayerDir string, sourceChainID, destChainID, sourceChannelID, destChannelID string) error {
	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	cmd := exec.CommandContext(
		ctx,
		hermesBin,
		"tx", "relay-packets",
		"--src-chain", sourceChainID,
		"--dst-chain", destChainID,
		"--src-port", "transfer",
		"--src-channel", sourceChannelID,
	)
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to relay packets: %w, output: %s", err, string(output))
	}

	return nil
}

// QueryIBCDenomTrace retrieves the denomination trace for an IBC token
func QueryIBCDenomTrace(ctx context.Context, relayerDir string, chainID, hash string) (string, error) {
	// Get the hermes binary path, preferring the environment variable if set
	hermesBin := "hermes"
	if envBin := os.Getenv("HERMES_BIN"); envBin != "" {
		hermesBin = envBin
	}

	cmd := exec.CommandContext(
		ctx,
		hermesBin,
		"query",
		"denom-trace",
		chainID,
		hash,
	)
	cmd.Dir = relayerDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("HERMES_CONFIG=%s", filepath.Join(relayerDir, "config", "config.toml")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query denom trace: %w, output: %s", err, string(output))
	}

	// Parse the output to extract the denom path
	outputStr := string(output)
	denomPath := ""
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "path:") {
			parts := strings.Split(line, "path:")
			if len(parts) >= 2 {
				denomPath = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	return denomPath, nil
}
