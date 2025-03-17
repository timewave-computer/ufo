#!/bin/bash
set -e

# Default binary type is "patched" if not specified
BINARY_TYPE=${1:-"patched"}
shift 2>/dev/null || true

# Check if binary type is valid
if [[ "$BINARY_TYPE" != "patched" && "$BINARY_TYPE" != "bridged" && "$BINARY_TYPE" != "all" ]]; then
    echo "Error: Invalid binary type '$BINARY_TYPE'"
    echo "Usage: $0 [patched|bridged|all] [test flags]"
    echo "  patched - Use osmosis-ufo-patched binary"
    echo "  bridged - Use osmosis-ufo-bridged binary"
    echo "  all     - Run tests with both binaries"
    exit 1
fi

# Get the project root directory
PROJECT_ROOT=$(git rev-parse --show-toplevel)
if [ -z "$PROJECT_ROOT" ]; then
    PROJECT_ROOT=$(pwd)
fi

# Check if we're in a Nix environment with Hermes
if [[ -n "$IN_NIX_SHELL" ]] && command -v hermes &> /dev/null; then
    echo "Using Hermes from Nix environment: $(hermes version)"
    export HERMES_BIN=$(which hermes)
# If not in Nix or Hermes isn't available, check if Hermes is installed
elif ! command -v hermes &> /dev/null; then
    echo "Hermes not found, installing..."
    
    # Create bin directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/bin"
    
    # Download Hermes binary for macOS
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "Downloading Hermes for macOS..."
        curl -L https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-apple-darwin.tar.gz -o hermes.tar.gz
        tar -xzf hermes.tar.gz
        mv hermes "$PROJECT_ROOT/bin/"
        rm hermes.tar.gz
    else
        echo "Downloading Hermes for Linux..."
        curl -L https://github.com/informalsystems/hermes/releases/download/v1.7.4/hermes-v1.7.4-x86_64-unknown-linux-gnu.tar.gz -o hermes.tar.gz
        tar -xzf hermes.tar.gz
        mv hermes "$PROJECT_ROOT/bin/"
        rm hermes.tar.gz
    fi
    
    chmod +x "$PROJECT_ROOT/bin/hermes"
    export PATH="$PROJECT_ROOT/bin:$PATH"
    export HERMES_BIN="$PROJECT_ROOT/bin/hermes"
    
    echo "Hermes installed at $PROJECT_ROOT/bin/hermes"
    hermes version
fi

# Function to run tests with a specific binary
run_tests_with_binary() {
    local binary_type=$1
    local rest_args=("${@:2}")
    local binary_path
    
    if [[ "$binary_type" == "patched" ]]; then
        binary_path="./cmd/osmosis-ufo-patched"
        binary_name="osmosis-ufo-patched"
        echo "===== Running tests with PATCHED UFO binary ====="
    else
        binary_path="./cmd/osmosis-ufo-bridged"
        binary_name="osmosis-ufo-bridged"
        echo "===== Running tests with BRIDGED UFO binary ====="
    fi
    
    # Create a temporary directory for the test binaries
    TEMP_BIN_DIR="$PROJECT_ROOT/bin"
    mkdir -p "$TEMP_BIN_DIR"
    
    # Build the binary directly without Nix
    echo "Building $binary_name binary..."
    cd $PROJECT_ROOT
    go build -o $TEMP_BIN_DIR/$binary_name $binary_path
    
    # Create symlinks to the required binaries in the project root
    echo "Using $binary_name binary for IBC tests"
    ln -sf $TEMP_BIN_DIR/$binary_name $PROJECT_ROOT/$binary_name
    ln -sf $TEMP_BIN_DIR/$binary_name $TEMP_BIN_DIR/fauxmosis-ufo
    ln -sf $TEMP_BIN_DIR/$binary_name $TEMP_BIN_DIR/ufo
    
    # Add the bin directory to the PATH
    export PATH=$TEMP_BIN_DIR:$PATH
    export UFO_BINARY_TYPE="$binary_type"
    
    # Run the tests
    cd $PROJECT_ROOT
    
    # Use more timeouts for long-running tests like the light client test
    if [[ "${rest_args[*]}" == *TestIBCLightClient* ]]; then
        echo "Running light client test with extended timeout..."
        go test -v -timeout 20m ./tests/ibc -run TestIBCLightClientUpdates "${rest_args[@]}"
    else
        go test -v ./tests/ibc "${rest_args[@]}"
    fi
    
    # Clean up
    echo "Cleaning up $binary_type binaries..."
    rm -f "$PROJECT_ROOT/$binary_name"
}

# Run tests based on the binary type
if [[ "$BINARY_TYPE" == "all" ]]; then
    # Run tests with both binaries
    echo "Running tests with both PATCHED and BRIDGED binaries"
    run_tests_with_binary "patched" "$@"
    echo ""
    echo "=============================================="
    echo ""
    run_tests_with_binary "bridged" "$@"
else
    # Run tests with the specified binary
    run_tests_with_binary "$BINARY_TYPE" "$@"
fi

echo "Done!" 