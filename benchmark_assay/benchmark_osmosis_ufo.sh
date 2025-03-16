#!/bin/bash
# Script to benchmark Osmosis with UFO integration

set -e

# Default parameters
OSMOSIS_DIR="/tmp/osmosis-ufo-test"
BLOCK_TIMES=("1000" "100" "10" "1")  # block times in milliseconds
TX_COUNT=1000
DURATION=60  # seconds
TEST_DIR="/tmp/osmosis-benchmark"
VISUALIZE=false
CONCURRENCY=10 # Default concurrency for sending transactions
COMPARISON_FILE="" # For comparing results with a previous run
CHAIN_ID_PREFIX="ufo-bench"

# Parse arguments
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --osmosis-dir DIR     Directory containing osmosisd-ufo binary (default: $OSMOSIS_DIR)"
  echo "  --block-times LIST    Comma-separated list of block times in ms (default: 1000,100,10,1)"
  echo "  --tx-count COUNT      Number of transactions to send (default: $TX_COUNT)"
  echo "  --duration SECONDS    Duration of each test in seconds (default: $DURATION)"
  echo "  --concurrency NUM     Number of concurrent transactions to send (default: $CONCURRENCY)"
  echo "  --test-dir DIR        Directory for test data (default: $TEST_DIR)"
  echo "  --visualize           Generate visualization of results"
  echo "  --compare FILE        Compare results with a previous benchmark CSV file"
  echo "  --help                Display this help message"
}

# Process all arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --osmosis-dir)
      OSMOSIS_DIR="$2"
      shift 2
      ;;
    --block-times)
      IFS=',' read -ra BLOCK_TIMES <<< "$2"
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
    --test-dir)
      TEST_DIR="$2"
      shift 2
      ;;
    --visualize)
      VISUALIZE=true
      shift
      ;;
    --compare)
      COMPARISON_FILE="$2"
      shift 2
      ;;
    --help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Now setup logging
LOG_FILE="$TEST_DIR/benchmark.log"
RESULTS_FILE="$TEST_DIR/results.csv"

# Check if Osmosis binary exists
if [ ! -f "$OSMOSIS_DIR/osmosisd-ufo" ]; then
  echo "Error: Osmosis UFO binary not found at $OSMOSIS_DIR/osmosisd-ufo"
  exit 1
fi

# Check for required tools
for cmd in jq bc; do
  if ! command -v $cmd &> /dev/null; then
    echo "Error: $cmd command not found but required"
    exit 1
  fi
done

# Create test directory
mkdir -p "$TEST_DIR"
echo "Test directory: $TEST_DIR"

# Log configuration
echo "=== Configuration ===" | tee -a "$LOG_FILE"
echo "Osmosis binary: $OSMOSIS_DIR/osmosisd-ufo" | tee -a "$LOG_FILE"
echo "Block times: ${BLOCK_TIMES[*]}" | tee -a "$LOG_FILE" 
echo "Transaction count: $TX_COUNT" | tee -a "$LOG_FILE"
echo "Test duration: $DURATION seconds" | tee -a "$LOG_FILE"
echo "Concurrency: $CONCURRENCY" | tee -a "$LOG_FILE"
echo "Visualize: $VISUALIZE" | tee -a "$LOG_FILE"
echo "Test directory: $TEST_DIR" | tee -a "$LOG_FILE"
echo "Results file: $RESULTS_FILE" | tee -a "$LOG_FILE"
echo "===================" | tee -a "$LOG_FILE"

# Create results file with header
echo "block_time,tps,latency_ms,cpu_usage,memory_usage,blocks_produced,avg_tx_per_block" > "$RESULTS_FILE"

# Log benchmark parameters
echo "=== Benchmark Parameters ===" | tee -a "$LOG_FILE"
echo "Osmosis UFO Directory: $OSMOSIS_DIR" | tee -a "$LOG_FILE"
echo "Block Times (ms): ${BLOCK_TIMES[*]}" | tee -a "$LOG_FILE"
echo "Transaction Count: $TX_COUNT" | tee -a "$LOG_FILE"
echo "Test Duration: $DURATION seconds" | tee -a "$LOG_FILE"
echo "Concurrency: $CONCURRENCY" | tee -a "$LOG_FILE"
echo "Results File: $RESULTS_FILE" | tee -a "$LOG_FILE"
echo "Timestamp: $(date)" | tee -a "$LOG_FILE"
echo "===========================" | tee -a "$LOG_FILE"

# Function to run a single benchmark
function run_benchmark {
  local block_time=$1
  local test_home="$TEST_DIR/test_$block_time"
  local chain_id="$CHAIN_ID_PREFIX-$block_time"
  
  echo "=== Running benchmark with $block_time ms block time ===" | tee -a "$LOG_FILE"
  echo "Setting up test environment..." | tee -a "$LOG_FILE"
  
  # Initialize Osmosis node
  rm -rf "$test_home"
  mkdir -p "$test_home"
  
  echo "Initializing Osmosis node..." | tee -a "$LOG_FILE"
  "$OSMOSIS_DIR/osmosisd-ufo" init benchmark --home "$test_home" --chain-id "$chain_id" > /dev/null 2>&1
  
  # Set block time in config
  block_time_seconds=$(echo "scale=3; $block_time/1000" | bc)
  sed -i.bak "s/timeout_commit = \"5s\"/timeout_commit = \"${block_time_seconds}s\"/" "$test_home/config/config.toml"
  sed -i.bak "s/timeout_propose = \"3s\"/timeout_propose = \"${block_time_seconds}s\"/" "$test_home/config/config.toml"
  
  # Enable Prometheus metrics
  sed -i.bak 's/prometheus = false/prometheus = true/' "$test_home/config/config.toml"
  
  # Add test account
  "$OSMOSIS_DIR/osmosisd-ufo" keys add test-account --keyring-backend test --home "$test_home" > /dev/null 2>&1
  address=$("$OSMOSIS_DIR/osmosisd-ufo" keys show test-account -a --keyring-backend test --home "$test_home")
  
  # Add account to genesis
  "$OSMOSIS_DIR/osmosisd-ufo" add-genesis-account $address 10000000000uosmo,10000000000uatom --home "$test_home" > /dev/null 2>&1
  
  # Create validator
  "$OSMOSIS_DIR/osmosisd-ufo" gentx test-account 1000000uosmo --chain-id "$chain_id" --keyring-backend test --home "$test_home" > /dev/null 2>&1
  
  # Collect genesis transactions
  "$OSMOSIS_DIR/osmosisd-ufo" collect-gentxs --home "$test_home" > /dev/null 2>&1
  
  # Start the node
  echo "Starting node with block time $block_time ms..." | tee -a "$LOG_FILE"
  "$OSMOSIS_DIR/osmosisd-ufo" start --home "$test_home" > "$test_home/node.log" 2>&1 &
  node_pid=$!
  
  # Wait for node to start
  echo "Waiting for node to start..." | tee -a "$LOG_FILE"
  sleep 10
  
  # Get initial block height
  initial_height=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
  echo "Initial block height: $initial_height" | tee -a "$LOG_FILE"
  
  # Run test transactions
  echo "Sending $TX_COUNT test transactions with concurrency $CONCURRENCY..." | tee -a "$LOG_FILE"
  start_time=$(date +%s.%N)
  
  # Start monitoring
  cpu_usage_start=$(ps -p $node_pid -o %cpu | tail -1 | tr -d ' ')
  memory_usage_start=$(ps -p $node_pid -o %mem | tail -1 | tr -d ' ')
  
  # Send transactions in batches with concurrency
  for ((i=1; i<=$TX_COUNT; i+=$CONCURRENCY)); do
    batch_size=$CONCURRENCY
    if [ $(($i + $CONCURRENCY)) -gt $TX_COUNT ]; then
      batch_size=$(($TX_COUNT - $i + 1))
    fi
    
    for ((j=0; j<$batch_size; j++)); do
      # Send a simple bank transaction
      "$OSMOSIS_DIR/osmosisd-ufo" tx bank send test-account $address 1uosmo \
        --keyring-backend test --chain-id "$chain_id" --home "$test_home" \
        --broadcast-mode async --gas-prices 0.1uosmo --gas auto --gas-adjustment 1.5 -y > /dev/null 2>&1 &
    done
    
    # Wait for all transactions in this batch
    wait
    
    # Small delay between batches to prevent overwhelming the node
    sleep 0.01
  done
  
  # Wait for transactions to be processed
  echo "Waiting for transactions to be processed (${DURATION}s)..." | tee -a "$LOG_FILE"
  sleep $DURATION
  
  # End monitoring
  cpu_usage_end=$(ps -p $node_pid -o %cpu | tail -1 | tr -d ' ')
  memory_usage_end=$(ps -p $node_pid -o %mem | tail -1 | tr -d ' ')
  end_time=$(date +%s.%N)
  
  # Get final block height
  final_height=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
  blocks_produced=$((final_height - initial_height))
  echo "Final block height: $final_height" | tee -a "$LOG_FILE"
  echo "Blocks produced during test: $blocks_produced" | tee -a "$LOG_FILE"
  
  # Get unconfirmed transaction count
  tx_count=$(curl -s http://localhost:26657/num_unconfirmed_txs | jq -r '.result.n_txs')
  completed_tx=$(echo "$TX_COUNT - $tx_count" | bc)
  
  # Calculate metrics
  total_time=$(echo "$end_time - $start_time" | bc)
  tps=$(echo "scale=2; $completed_tx / $total_time" | bc)
  latency=$(echo "scale=2; 1000 * $total_time / $completed_tx" | bc)
  
  # Calculate CPU and memory usage
  cpu_usage=$(echo "$cpu_usage_end - $cpu_usage_start" | bc)
  memory_usage=$(echo "$memory_usage_end - $memory_usage_start" | bc)
  
  # Calculate average transactions per block
  avg_tx_per_block=0
  if [ $blocks_produced -gt 0 ]; then
    avg_tx_per_block=$(echo "scale=2; $completed_tx / $blocks_produced" | bc)
  fi
  
  # Log results
  echo "Results for block time: $block_time ms" | tee -a "$LOG_FILE"
  echo "TPS: $tps" | tee -a "$LOG_FILE"
  echo "Latency: $latency ms" | tee -a "$LOG_FILE"
  echo "CPU usage: $cpu_usage%" | tee -a "$LOG_FILE"
  echo "Memory usage: $memory_usage%" | tee -a "$LOG_FILE"
  echo "Blocks produced: $blocks_produced" | tee -a "$LOG_FILE"
  echo "Avg transactions per block: $avg_tx_per_block" | tee -a "$LOG_FILE"
  
  # Save results to CSV
  echo "$block_time,$tps,$latency,$cpu_usage,$memory_usage,$blocks_produced,$avg_tx_per_block" >> "$RESULTS_FILE"
  
  # Kill the node
  echo "Stopping node..." | tee -a "$LOG_FILE"
  kill $node_pid
  wait $node_pid 2>/dev/null || true
  sleep 2
}

# Run benchmarks for each block time
for block_time in "${BLOCK_TIMES[@]}"; do
  run_benchmark "$block_time"
done

# Generate performance report
echo -e "\n=== Performance Report ===" | tee -a "$LOG_FILE"
echo "Block Time (ms) | TPS | Latency (ms) | CPU Usage (%) | Memory Usage (%) | Blocks | Avg Tx/Block" | tee -a "$LOG_FILE"
echo "--------------- | --- | ------------ | ------------- | --------------- | ------ | ------------" | tee -a "$LOG_FILE"
while IFS="," read -r block_time tps latency cpu_usage memory_usage blocks_produced avg_tx_per_block; do
  if [ "$block_time" != "block_time" ]; then  # Skip header
    printf "%-15s | %-3s | %-12s | %-13s | %-15s | %-6s | %-12s\n" "$block_time" "$tps" "$latency" "$cpu_usage" "$memory_usage" "$blocks_produced" "$avg_tx_per_block" | tee -a "$LOG_FILE"
  fi
done < "$RESULTS_FILE"

echo -e "\nBenchmark completed. Results saved to $RESULTS_FILE" | tee -a "$LOG_FILE"

# Compare with previous results if requested
if [ -n "$COMPARISON_FILE" ] && [ -f "$COMPARISON_FILE" ]; then
  echo -e "\n=== Comparing with previous results ===" | tee -a "$LOG_FILE"
  echo "Current results vs results from $COMPARISON_FILE" | tee -a "$LOG_FILE"
  echo "Block Time (ms) | TPS Change (%) | Latency Change (%)" | tee -a "$LOG_FILE"
  echo "--------------- | -------------- | -----------------" | tee -a "$LOG_FILE"
  
  # Skip headers
  tail -n +2 "$RESULTS_FILE" > "$TEST_DIR/current.tmp"
  tail -n +2 "$COMPARISON_FILE" > "$TEST_DIR/previous.tmp"
  
  # Process each line in current results
  while IFS="," read -r block_time tps latency cpu_usage memory_usage blocks_produced avg_tx_per_block; do
    # Find matching block time in previous results
    prev_line=$(grep "^$block_time," "$TEST_DIR/previous.tmp")
    if [ -n "$prev_line" ]; then
      prev_tps=$(echo "$prev_line" | cut -d',' -f2)
      prev_latency=$(echo "$prev_line" | cut -d',' -f3)
      
      # Calculate changes
      tps_change=$(echo "scale=2; 100 * ($tps - $prev_tps) / $prev_tps" | bc)
      latency_change=$(echo "scale=2; 100 * ($latency - $prev_latency) / $prev_latency" | bc)
      
      # Print comparison
      printf "%-15s | %+14s | %+17s\n" "$block_time" "$tps_change%" "$latency_change%" | tee -a "$LOG_FILE"
    fi
  done < "$TEST_DIR/current.tmp"
  
  # Clean up temporary files
  rm -f "$TEST_DIR/current.tmp" "$TEST_DIR/previous.tmp"
fi

# Visualize results if requested
if [ "$VISUALIZE" = true ]; then
  echo -e "\nGenerating visualizations..." | tee -a "$LOG_FILE"
  # Check if visualization script exists
  if [ -f "$(dirname "$0")/visualize_benchmark.py" ]; then
    "$(dirname "$0")/visualize_benchmark.py" "$RESULTS_FILE"
  else
    echo "Warning: Visualization script not found at $(dirname "$0")/visualize_benchmark.py" | tee -a "$LOG_FILE"
    echo "Visualizations will not be generated." | tee -a "$LOG_FILE"
  fi
fi

echo -e "\nAll tests completed successfully!" | tee -a "$LOG_FILE"
echo "Log file: $LOG_FILE" | tee -a "$LOG_FILE"
echo "Results file: $RESULTS_FILE" | tee -a "$LOG_FILE" 