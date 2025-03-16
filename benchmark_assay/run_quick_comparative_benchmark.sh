#!/bin/bash
# Script to run a quick comparative benchmark with all configurations
# This is useful for testing and demonstration purposes

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Default quick benchmark settings
BLOCK_TIMES="1000,100,10"  # Only test 3 representative block times
TX_COUNT=100               # Use a small number of transactions
DURATION=30                # Short test duration
CONCURRENCY=5              # Low concurrency
RUN_NAME="quick_comparison_$(date +%Y%m%d_%H%M%S)"

echo "============================================"
echo "= UFO Quick Comparative Benchmark"
echo "============================================"
echo "This script runs a quick benchmark with all 10 configurations:"
echo "  1. Fauxmosis with CometBFT (1 validator)"
echo "  2. Fauxmosis with CometBFT (4 validators)"
echo "  3. Fauxmosis with UFO (1 validator)"
echo "  4. Fauxmosis with UFO (4 validators)"
echo "  5. Osmosis with UFO Bridged (1 validator)"
echo "  6. Osmosis with UFO Bridged (4 validators)"
echo "  7. Osmosis with UFO Patched (1 validator)"
echo "  8. Osmosis with UFO Patched (4 validators)"
echo "  9. Osmosis with CometBFT (1 validator)"
echo " 10. Osmosis with CometBFT (4 validators)"
echo ""
echo "Using minimal settings for quick testing:"
echo "  - Block times: $BLOCK_TIMES ms"
echo "  - Transactions: $TX_COUNT"
echo "  - Duration: $DURATION seconds"
echo "  - Concurrency: $CONCURRENCY"
echo "============================================"
echo

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --block-times)
      BLOCK_TIMES="$2"
      shift 2
      ;;
    --tx-count)
      TX_COUNT="$2"
      shift 2
      ;;
    --duration)
      DURATION="$2"
      shift 2
      ;;
    --concurrency)
      CONCURRENCY="$2"
      shift 2
      ;;
    --run-name)
      RUN_NAME="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Available options:"
      echo "  --block-times LIST    Comma-separated list of block times in ms (default: $BLOCK_TIMES)"
      echo "  --tx-count COUNT      Number of transactions (default: $TX_COUNT)"
      echo "  --duration SECONDS    Test duration in seconds (default: $DURATION)"
      echo "  --concurrency NUM     Number of concurrent transactions (default: $CONCURRENCY)"
      echo "  --run-name NAME       Custom name for the benchmark run (default: $RUN_NAME)"
      exit 1
      ;;
  esac
done

# Run the performance tests script with quick settings and all configurations
echo "Starting quick comparative benchmark..."
"$SCRIPT_DIR/run_performance_tests.sh" \
  --block-times "$BLOCK_TIMES" \
  --tx-count "$TX_COUNT" \
  --duration "$DURATION" \
  --concurrency "$CONCURRENCY" \
  --run-name "$RUN_NAME" \
  --configurations "all"

echo
echo "Quick comparative benchmark completed!"
echo "Results are available in: $PROJECT_DIR/benchmark_results/$RUN_NAME"
echo
echo "Jupyter notebook generated at:"
echo "$PROJECT_DIR/benchmark_results/$RUN_NAME/benchmark_analysis.ipynb"
echo
echo "To open the notebook using Nix:"
echo "nix run .#notebook \"$PROJECT_DIR/benchmark_results/$RUN_NAME/benchmark_analysis.ipynb\""
echo
echo "Or for a more interactive experience:"
echo "nix develop .#jupyter"
echo
echo "To consolidate these results as the primary benchmark:"
echo "$SCRIPT_DIR/cleanup_benchmark_results.sh --keep-run $RUN_NAME" 