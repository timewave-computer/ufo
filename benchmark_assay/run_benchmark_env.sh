#!/bin/bash
# Script to launch a Nix development shell for benchmarking

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "============================================"
echo "= UFO Benchmark Environment"
echo "============================================"
echo "Launching Nix development shell with all required packages..."
echo "This might take a few minutes on first run..."
echo "============================================"

# Launch a development shell for benchmarking
nix develop --impure "$PROJECT_DIR#benchmark"

# If the user doesn't have flakes enabled, fall back to nix-shell
if [ $? -ne 0 ]; then
  echo "Flakes might not be enabled, falling back to nix-shell..."
  nix-shell -p "python3.withPackages(ps: with ps; [ pandas matplotlib numpy ])" \
            -p jq -p bc -p curl \
            --run "bash -c 'cd \"$PROJECT_DIR\" && echo \"UFO Benchmark Environment ready.\" && echo \"Run a benchmark with: ./benchmark_assay/run_performance_tests.sh\" && exec bash'"
fi 