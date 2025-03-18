#!/usr/bin/env bash
set -eo pipefail

# Script to build binaries and place them in the bin directory
# This is the central build script used by other scripts and the Nix environment

PROJECT_ROOT=$(git rev-parse --show-toplevel)
BIN_DIR="${PROJECT_ROOT}/bin"

# Create bin directory if it doesn't exist
mkdir -p "${BIN_DIR}"

echo "Building UFO binaries..."

# Set platform-specific environment variables
if [[ "$(uname)" == "Darwin" ]]; then
    echo "Detected Darwin (macOS) platform"
    export CGO_ENABLED=1
elif [[ "$(uname)" == "Linux" ]]; then
    echo "Detected Linux platform"
    export CGO_ENABLED=1
    # These will be overridden in nix shells, but helpful when running directly
    if command -v pkg-config &>/dev/null; then
        echo "Using pkg-config for library detection"
    else
        echo "pkg-config not found, setting manual library paths"
        # Set default paths that will be overridden in nix environment
        export CGO_CFLAGS="-I/usr/include"
        export CGO_LDFLAGS="-L/usr/lib -L/usr/lib/x86_64-linux-gnu"
    fi
else
    echo "Unknown platform: $(uname)"
fi

# Check if we're in a nix IBC shell specifically
if [[ "$IN_NIX_SHELL" != "" && "$(go version 2>/dev/null | grep "go1.22")" != "" ]]; then
    echo "✅ Building in compatible nix environment with Go 1.22"
    
    # Build main binaries based on available cmd directories
    echo "Building main binaries..."
    go build -o "${BIN_DIR}/osmosis-ufo-patched" ./cmd/osmosis-ufo-patched
    go build -o "${BIN_DIR}/osmosis-ufo-bridged" ./cmd/osmosis-ufo-bridged
    go build -o "${BIN_DIR}/osmosis-comet" ./cmd/osmosis-comet
    go build -o "${BIN_DIR}/fauxmosis-ufo" ./cmd/fauxmosis-ufo
    go build -o "${BIN_DIR}/fauxmosis-comet" ./cmd/fauxmosis-comet
else
    echo "⚠️ Not in compatible nix environment, using nix develop .#ibc to build"
    
    # Use nix develop .#ibc to ensure correct build environment
    # Build each binary separately to ensure proper error handling
    nix develop .#ibc -c bash -c "cd ${PROJECT_ROOT} && go build -o ${BIN_DIR}/osmosis-ufo-patched ./cmd/osmosis-ufo-patched"
    nix develop .#ibc -c bash -c "cd ${PROJECT_ROOT} && go build -o ${BIN_DIR}/osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged"
    nix develop .#ibc -c bash -c "cd ${PROJECT_ROOT} && go build -o ${BIN_DIR}/osmosis-comet ./cmd/osmosis-comet"
    nix develop .#ibc -c bash -c "cd ${PROJECT_ROOT} && go build -o ${BIN_DIR}/fauxmosis-ufo ./cmd/fauxmosis-ufo"
    nix develop .#ibc -c bash -c "cd ${PROJECT_ROOT} && go build -o ${BIN_DIR}/fauxmosis-comet ./cmd/fauxmosis-comet"
fi

# Mark all binaries as executable
chmod +x "${BIN_DIR}"/*

# Create symlink from result to bin
rm -f "${PROJECT_ROOT}/result"
ln -sf "bin" "${PROJECT_ROOT}/result"

echo "All binaries built successfully in: ${BIN_DIR}"
ls -la "${BIN_DIR}"

# Output success message
echo "✅ Build complete! The binaries are accessible at: ${BIN_DIR}/"
echo "  You can also access them via the 'result/' symlink" 