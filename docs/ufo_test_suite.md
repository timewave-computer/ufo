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

## IBC Tests

The IBC tests verify the functionality of UFO's IBC implementation by setting up test chains, creating connections and channels, and performing token transfers.

### IBC Test Categories

1. **Basic Setup Tests**
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

To run the IBC tests with all binary types:

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
