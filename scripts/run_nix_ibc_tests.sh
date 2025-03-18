#!/usr/bin/env bash
set -eo pipefail

# Script to run nix-compatible IBC tests

echo "==== Running Nix-Compatible IBC Tests ===="
echo "This script runs IBC tests that are compatible with the nix environment"
echo "These tests work with auto-starting binaries in the nix environment"
echo ""
echo "Available tests:"
echo "  - TestIBCBasicTransfer          - Basic chain startup and IBC transfer test"
echo "  - TestIBCLightClientUpdates     - Light client functionality test"
echo "  - TestIBCClientTrustingPeriod   - Security aspects test"
echo "  - TestDoubleSpendPrevention     - Double spend prevention test"
echo "  - TestIBCChannelSecurity        - Channel security validation test"
echo "  - TestIBCPacketDataValidation   - Packet data validation test"
echo "  - TestIBCFork                   - IBC fork test"
echo "  - TestIBCByzantineValidators    - Byzantine validator test"
echo "  - TestIBCEnvironmentSetup       - Environment setup test"
echo "  - TestHermesRelayerConfig       - Hermes configuration test"
echo "  - TestMultiChainIBCSetup        - Multi-chain setup test"
echo "  - TestHermesConfigCreation      - Hermes config creation test"
echo ""

# Check if we're in a nix environment
if [[ "$IN_NIX_SHELL" != "" || "$PATH" == *"/nix/store"* ]]; then
    echo "✅ Nix environment detected"
else
    echo "⚠️ Warning: Not running in a nix environment. These tests are designed for nix."
    echo "You should run this script within 'nix develop .#ibc'"
    if [[ "$1" != "--force" ]]; then
        echo "To run anyway, use the --force flag"
        exit 1
    fi
    echo "Running anyway due to --force flag..."
fi

# Ensure we're in the project root
cd "$(git rev-parse --show-toplevel)"

# Build test binaries
echo ""
echo "Building fresh test binaries..."
if type build_test_binaries &>/dev/null; then
    build_test_binaries
else
    echo "Building manually..."
    # Use the central build script if it exists
    if [ -f "./scripts/build_binaries.sh" ]; then
        ./scripts/build_binaries.sh
    else
        # Fallback to direct builds
        mkdir -p bin
        go build -o bin/osmosis-ufo-patched ./cmd/osmosis-ufo-patched
        go build -o bin/osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged
        # Create symlink for backward compatibility
        rm -f result
        ln -sf bin result
    fi
    echo "IBC test binaries built successfully at: ./bin/"
    ls -la ./bin/
fi

# Set required environment variables
export UFO_BINARY_TYPE=patched
export UFO_BIN=$(pwd)/bin/osmosis-ufo-patched
chmod +x $UFO_BIN

echo ""
echo "Running nix-compatible tests..."
echo ""

# Run tests with a timeout (adjust as needed)
timeout_duration="10m"

# Run the nix-compatible tests using nix develop .#ibc
if [[ "$1" == "--specific" && "$2" != "" ]]; then
    test_pattern="$2"
    echo "Running specific test: $test_pattern"
    nix develop .#ibc --command bash -c "cd $(pwd) && go test -v ./tests/ibc/... -run $test_pattern"
else
    echo "Running all nix-compatible tests"
    nix develop .#ibc --command bash -c "cd $(pwd) && go test -v ./tests/ibc/..."
fi

# Check exit code
exit_code=$?
if [[ $exit_code -eq 0 ]]; then
    echo ""
    echo "✅ All nix-compatible tests passed!"
    exit 0
elif [[ $exit_code -eq 124 ]]; then
    echo ""
    echo "❌ Tests timed out after $timeout_duration"
    exit 1
else
    echo ""
    echo "❌ Some tests failed with exit code $exit_code"
    exit $exit_code
fi 