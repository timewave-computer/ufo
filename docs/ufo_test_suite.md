# UFO Test Suite

This document describes the test suite for the UFO (Universal Fast Orderer) project, including how to run tests, the test directory structure, and specific information about IBC tests.

## Overview

The UFO test suite includes various categories of tests:

- **Core Tests**: Tests for core functionality
- **Consensus Tests**: Tests for the consensus algorithm
- **IBC Tests**: Tests for Inter-Blockchain Communication
- **Integration Tests**: Tests for integrating various components
- **Stress Tests**: Performance and stability tests

## Running Tests with Nix

All test functionality is integrated with our Nix flake system, providing a consistent and reproducible workflow.

### Available Test Commands

Run these commands from the project root using `nix run`:

```bash
# List all available test commands
nix run .#test-help

# Run all tests
nix run .#test-all

# Run specific test categories
nix run .#test-core
nix run .#test-consensus
nix run .#test-ibc
nix run .#test-integration
nix run .#test-stress

# Run tests with special options
nix run .#test-verbose    # Verbose output
nix run .#test-race       # Race detection
nix run .#test-cover      # Coverage report
nix run .#test-cover-html # HTML coverage report

# Other commands
nix run .#test-mockgen    # Generate mocks
nix run .#test-lint       # Run linter
nix run .#test-clean      # Clean test artifacts
```

### Building Test Binaries

Before running tests, you may need to build all the required binaries:

```bash
# Build all UFO binaries including test binaries
nix run .#build-all
```

This will build and place all binaries in the `tests/bin` directory.

## Test Directory Structure

- `core/` - Core functionality tests
- `consensus/` - Consensus algorithm tests
- `ibc/` - Inter-Blockchain Communication tests
- `integration/` - Integration tests
- `stress/` - Stress and performance tests
- `utils/` - Testing utilities
- `bin/` - Test binaries (generated during build)
- `scripts/` - Test scripts for running tests in various configurations

## IBC Tests

The IBC tests verify the functionality of UFO's IBC implementation by setting up test chains, creating connections and channels, and performing token transfers.

### IBC Test Categories

1. **Basic Setup Tests**
   - `TestIBCBasicTransfer`: Simplified test for verifying basic IBC functionality with a single validator per chain
   - `TestIBCEnvironmentSetup`: Tests the basic setup of an IBC-enabled environment with two chains
   - `TestHermesRelayerConfiguration`: Tests the configuration options of the Hermes relayer
   - `TestMultiChainIBCSetup`: Tests the setup of an IBC environment with more than two chains

2. **Transfer Tests**
   - `TestBasicIBCTransfer`: Tests a basic IBC transfer between two chains
   - `TestBidirectionalIBCTransfer`: Tests transferring tokens back and forth between two chains
   - `TestIBCTimeouts`: Tests that IBC transfers respect timeout parameters
   - `TestCustomIBCChannelConfig`: Tests creating IBC channels with custom configuration options

3. **Security Tests**
   - `TestIBCClientTrustingPeriodSecurity`: Tests the security implications of the trusting period
   - `TestIBCDoubleSpendPrevention`: Tests the prevention of double spending via IBC
   - `TestIBCChannelSecurityValidation`: Tests the security validation of IBC channels
   - `TestIBCPacketDataValidation`: Tests the validation of IBC packet data

4. **Byzantine Tests**
   - `TestIBCByzantineValidators`: Tests IBC behavior with Byzantine validators

5. **Fork Tests**
   - `TestIBCFork`: Tests IBC behavior during a chain fork

6. **Light Client Tests**
   - `TestIBCLightClientUpdates`: Tests IBC light client updates with 4 validators per chain to verify that IBC connections remain functional when validator sets change

### Binary Types

UFO tests support multiple binary types:

1. **Patched Binary (osmosis-ufo-patched)**
   - Uses the patched approach where CometBFT imports are directly replaced with UFO adapters
   - More tightly integrated with the Osmosis codebase

2. **Bridged Binary (osmosis-ufo-bridged)**
   - Uses the bridge approach where UFO connects to Osmosis as a separate service
   - More loosely coupled with the Osmosis codebase

### Running IBC Tests

All test scripts are located in the `scripts/` directory. To run the IBC tests with all binary types:

```bash
./scripts/run_ibc_tests.sh all
```

To run the tests with a specific binary type:

```bash
# Run with patched binary
./scripts/run_ibc_tests.sh patched

# Run with bridged binary
./scripts/run_ibc_tests.sh bridged
```

To run a specific test with a specific binary type:

```bash
# Run light client test with patched binary
./scripts/run_ibc_tests.sh patched -run TestIBCLightClientUpdates

# Run light client test with bridged binary
./scripts/run_ibc_tests.sh bridged -run TestIBCLightClientUpdates
```

To run a simplified IBC test with a single validator per chain:

```bash
# Run the simplified IBC test
./scripts/run_simple_ibc_test.sh
```

To execute the light client test with detailed output logging:

```bash
# Run the light client test with output capture
./scripts/run_and_capture.sh
```

For diagnosing issues with chain startup:

```bash
# Run the chain diagnostic test
./scripts/diagnose_chain_start.sh
```

### IBC Test Configuration

The tests use the following configuration:

- Two or more UFO chains running in separate directories
- Hermes relayer for IBC communication
- Fast block times (typically 100ms-1s) for quicker test execution
- Temporary directories for chain data and relayer configuration
- Multi-validator setups (typically 4 validators per chain for light client tests)
- Configurable epoch settings for testing epoch boundary behaviors

### Test Environment

The tests are run in a Nix shell environment that provides all the necessary dependencies, including:
- Go compiler
- Hermes IBC relayer
- UFO binary (built during test setup)

The environment is set up automatically by the test script, so you don't need to install these dependencies manually.

## Resource Considerations and Troubleshooting

### Resource Requirements

IBC tests, especially those with multiple validators, can be resource-intensive. Here are some guidelines:

- **Memory:** IBC tests with 4 validators per chain typically require 8-12GB RAM
- **CPU:** At least 4 CPU cores recommended for full test suite
- **Disk Space:** At least 10GB free space for chain data and logs
- **Network:** Tests use various ports (26656-26665, 9090-9099, etc.), ensure they are available

### Resource-Limited Environments

If you're running tests on resource-constrained machines, consider:

1. **Minimal Test Example**

For CI or machines with limited resources, use the simplified IBC test:

```go
// Minimal IBC test example for CI or resource-constrained environments
func TestIBCMinimal(t *testing.T) {
    // Skip if CI flag is set or in a resource-constrained environment
    if os.Getenv("CI") != "" || os.Getenv("MINIMAL_TEST") != "" {
        t.Skip("Skipping full IBC test in CI or minimal test environment")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Use a single validator per chain and slower block times
    chain1Config := utils.TestConfig{
        ChainID:        "minimal-a",
        RPCAddress:     "tcp://localhost:26657",
        HomeDir:        t.TempDir(),
        BlockTime:      "1s",          // Slower block time
        ValidatorCount: 1,             // Single validator
        EpochLength:    5,             // Shorter epochs
    }
    chain1Config.BinaryType = utils.BinaryTypeOsmosisUfoPatched

    // Start chain in non-blocking mode - implement actual test logic
}
```

2. **Sequential Test Execution**

For running multiple tests, run them sequentially rather than in parallel:

```bash
# Run tests one at a time
go test -p 1 ./tests/ibc -run TestIBCBasic
```

3. **Docker or VM with Resource Limits**

Set explicit resource limits when running in containerized environments:

```bash
docker run --memory=4g --cpus=2 ufo-test:latest ./scripts/run_simple_ibc_test.sh
```

### Troubleshooting Common Issues

If you encounter "signal: killed" or test timeouts:

- **Process Termination:** If you see "signal: killed", the OS is terminating the process due to excessive memory use
  - **Solution:** Increase available memory or reduce test requirements (validators, block time)

- **Slow Test Execution:** If tests are timing out but processes aren't killed
  - **Solution:** Increase timeout values in the test or run with fewer validators

- **Port Conflicts:** If you see connection refused errors
  - **Solution:** Ensure no other processes are using the required ports or customize port configuration

- **Binary Not Found:** If tests fail with "binary not found"
  - **Solution:** Run in the Nix environment with `nix develop .#ibc` or ensure binaries are built with `nix run .#build-all`

## Running Tests Manually

If you need more control over the test execution, you can:

1. Enter the Nix shell:
   ```bash
   nix develop .#ibc
   ```

2. Build the required binary:
   ```bash
   go build -o ./bin/osmosis-ufo-patched ./cmd/osmosis-ufo-patched
   ```

3. Run the test directly:
   ```bash
   UFO_BINARY_TYPE=patched PATH=$PWD/bin:$PATH go test -v ./tests/ibc -run TestIBCLightClientUpdates
   ```

## Troubleshooting

- If you encounter issues with the Nix environment, try running `nix develop --pure` to ensure a clean environment.
- Check the test logs for detailed error messages.
- Increase test timeouts for slow machines using the `-timeout` flag: `go test -timeout 30m ./tests/ibc`.
- Make sure the required ports (26657, 26658, etc.) are available and not used by other processes.
- Check the logs in the temporary directories for detailed error information.
- For IBC tests, ensure that the Hermes binary is properly installed and configured.
- For epoch-related tests, verify that the epoch configuration is properly set in both app.toml and genesis.json.
- If tests are being terminated with "signal: killed", your system may be running out of resources. Try:
  - Reducing the number of validators in the test configuration
  - Increasing the block time to reduce resource usage
  - Closing other resource-intensive applications while running tests
  - Allocating more memory/CPU to the test process if running in a virtualized environment
