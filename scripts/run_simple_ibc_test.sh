#!/usr/bin/env bash

echo "========================================================"
echo "🧪 Running simplified IBC Basic Transfer Test..."
echo "========================================================"
echo "Current directory: $(pwd)"

# Find the project root directory (directory containing flake.nix)
ROOT_DIR=$(pwd)
while [[ ! -f "$ROOT_DIR/flake.nix" && "$ROOT_DIR" != "/" ]]; do
    ROOT_DIR=$(dirname "$ROOT_DIR")
done

if [[ ! -f "$ROOT_DIR/flake.nix" ]]; then
    echo "❌ Error: Could not find flake.nix in parent directories."
    exit 1
fi

cd "$ROOT_DIR"
echo "Project root directory: $ROOT_DIR"

# Set binary type if not already set
if [ -z "$UFO_BINARY_TYPE" ]; then
    export UFO_BINARY_TYPE="patched"
fi

echo "Using binary type: $UFO_BINARY_TYPE"

# Clean up any previous output
rm -f test_output.log

MAX_ATTEMPTS=3
ATTEMPT=1
TIMEOUT_BASE=2  # Base timeout in minutes

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
    TIMEOUT=$((TIMEOUT_BASE * ATTEMPT))  # Increase timeout with each attempt
    
    echo "========================================================"
    echo "🔄 Test attempt $ATTEMPT of $MAX_ATTEMPTS (Timeout: ${TIMEOUT}m)"
    echo "========================================================"

    # Build the test command with increasing timeout
    TEST_CMD="timeout ${TIMEOUT}m go test -v ./tests/ibc -run TestIBCBasicTransfer"
    
    # Run the test with current timeout
    echo "Running: $TEST_CMD"
    eval "$TEST_CMD" 2>&1 | tee test_output.log
    EXIT_CODE=${PIPESTATUS[0]}
    
    if [ $EXIT_CODE -eq 0 ]; then
        echo "========================================================"
        echo "✅ IBC Basic Transfer Test PASSED!"
        echo "========================================================"
        exit 0
    elif [ $EXIT_CODE -eq 124 ]; then
        echo "⏱️ Test timed out after ${TIMEOUT} minutes."
    else
        echo "❌ Test failed with exit code $EXIT_CODE."
    fi
    
    # Check if it was killed due to resource limits
    if grep -q "signal: killed" test_output.log; then
        echo "⚠️ Test was killed due to resource constraints."
    fi
    
    # Increase resources for next attempt
    if [ $ATTEMPT -lt $MAX_ATTEMPTS ]; then
        echo "Waiting before retry..."
        sleep 5
        echo "Retrying with increased resources..."
    else
        echo "========================================================"
        echo "❌ IBC Basic Transfer Test FAILED after $MAX_ATTEMPTS attempts."
        echo "Check test_output.log for details."
        echo "========================================================"
        exit 1
    fi
    
    ((ATTEMPT++))
done 