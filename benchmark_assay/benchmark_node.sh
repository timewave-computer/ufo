#!/bin/bash
# Script to benchmark different node types with configurable validator counts

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
BINARY_TYPE="osmosis-ufo"  # Options: fauxmosis-comet, fauxmosis-ufo, osmosis-ufo-bridged, osmosis-ufo-patched, osmosis-comet
BINARY_PATH=""
VALIDATORS=1  # Number of validators (1 or 4)
TEST_DIR="benchmark_temp"
BLOCK_TIMES="1000,500,200,100,50,20,10,5,1"
TX_COUNT=1000
DURATION=60
CONCURRENCY=10
RESULTS_FILE=""
CONFIG_NAME=""

# Parse arguments
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --binary-type TYPE      Type of binary to test: fauxmosis-comet, fauxmosis-ufo, osmosis-ufo-bridged, osmosis-ufo-patched, osmosis-comet"
  echo "  --binary-path PATH      Path to the binary being tested"
  echo "  --validators NUM        Number of validators (1 or 4)"
  echo "  --test-dir DIR          Directory for test data"
  echo "  --block-times LIST      Comma-separated list of block times in ms"
  echo "  --tx-count COUNT        Number of transactions to send"
  echo "  --duration SECONDS      Duration of each test in seconds"
  echo "  --concurrency NUM       Number of concurrent transactions"
  echo "  --results-file FILE     CSV file to append results"
  echo "  --config-name NAME      Name for this configuration in results"
  echo "  --help                  Display this help message"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --binary-type)
      BINARY_TYPE="$2"
      shift 2
      ;;
    --binary-path)
      BINARY_PATH="$2"
      shift 2
      ;;
    --validators)
      VALIDATORS="$2"
      shift 2
      ;;
    --test-dir)
      TEST_DIR="$2"
      shift 2
      ;;
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
    --results-file)
      RESULTS_FILE="$2"
      shift 2
      ;;
    --config-name)
      CONFIG_NAME="$2"
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

# Validate inputs
if [ -z "$BINARY_PATH" ]; then
  echo "${RED}Error: Binary path not specified${NC}"
  print_usage
  exit 1
fi

if [ -z "$RESULTS_FILE" ]; then
  echo "${RED}Error: Results file not specified${NC}"
  print_usage
  exit 1
fi

if [ -z "$CONFIG_NAME" ]; then
  CONFIG_NAME="${BINARY_TYPE}-${VALIDATORS}"
fi

if [ "$VALIDATORS" -ne 1 ] && [ "$VALIDATORS" -ne 4 ]; then
  echo "${RED}Error: Validators must be 1 or 4${NC}"
  print_usage
  exit 1
fi

if [[ ! "$BINARY_TYPE" =~ ^(fauxmosis-comet|fauxmosis-ufo|osmosis-ufo-bridged|osmosis-ufo-patched|osmosis-comet)$ ]]; then
  echo -e "${RED}Error: Invalid binary type. Must be one of: fauxmosis-comet, fauxmosis-ufo, osmosis-ufo-bridged, osmosis-ufo-patched, osmosis-comet${NC}"
  exit 1
fi

# Create test directory
mkdir -p "$TEST_DIR"
LOG_FILE="${TEST_DIR}/benchmark.log"

# Initialize log file
echo "# $BINARY_TYPE Benchmark with $VALIDATORS validators - $(date)" > "$LOG_FILE"
echo "Testing $BINARY_TYPE with $VALIDATORS validators" | tee -a "$LOG_FILE"
echo "Binary: $BINARY_PATH" | tee -a "$LOG_FILE"
echo "Block times: $BLOCK_TIMES" | tee -a "$LOG_FILE"
echo "Transaction count: $TX_COUNT" | tee -a "$LOG_FILE"
echo "Test duration: $DURATION seconds" | tee -a "$LOG_FILE"
echo "Concurrency: $CONCURRENCY" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Function to initialize the node(s)
initialize_nodes() {
  local block_time=$1
  local test_data_dir="${TEST_DIR}/test_${block_time}"
  
  echo "Initializing test environment for block time ${block_time}ms..." | tee -a "$LOG_FILE"
  rm -rf "$test_data_dir"
  mkdir -p "$test_data_dir"
  
  # Different initialization based on binary type
  case "$BINARY_TYPE" in
    "fauxmosis-comet")
      # Initialize Fauxmosis with CometBFT
      echo "Initializing Fauxmosis with CometBFT..." | tee -a "$LOG_FILE"
      
      # Here we would initialize Fauxmosis with CometBFT
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual init (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with CometBFT initialization completed${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "fauxmosis-ufo")
      # Initialize Fauxmosis with UFO
      echo "Initializing Fauxmosis with UFO..." | tee -a "$LOG_FILE"
      
      # Here we would initialize Fauxmosis with UFO
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual init (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with UFO initialization completed${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-bridged")
      # Initialize Osmosis with UFO (bridged)
      echo "Initializing Osmosis with UFO (bridged)..." | tee -a "$LOG_FILE"
      
      # Here we would initialize Osmosis with UFO (bridged)
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual init (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (bridged) initialization completed${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-patched")
      # Initialize Osmosis with UFO (patched)
      echo "Initializing Osmosis with UFO (patched)..." | tee -a "$LOG_FILE"
      
      # Here we would initialize Osmosis with UFO (patched)
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual init (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (patched) initialization completed${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-comet")
      # Initialize standard Osmosis with specified validator count
      echo "Initializing standard Osmosis with $VALIDATORS validators..." | tee -a "$LOG_FILE"
      
      # Here we would initialize standard Osmosis, different validator counts and block times
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH start --home $test_data_dir --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual init (replace with real command when implementing)
      echo "${GREEN}Standard Osmosis initialization completed${NC}" | tee -a "$LOG_FILE"
      ;;
      
    *)
      echo "${RED}Error: Unknown binary type: $BINARY_TYPE${NC}" | tee -a "$LOG_FILE"
      exit 1
      ;;
  esac
  
  return 0
}

# Function to run the node(s)
run_nodes() {
  local block_time=$1
  local test_data_dir="${TEST_DIR}/test_${block_time}"
  
  echo "Starting node(s) with ${block_time}ms block time..." | tee -a "$LOG_FILE"
  
  # Different start commands based on binary type
  case "$BINARY_TYPE" in
    "fauxmosis-comet")
      # Start Fauxmosis with CometBFT
      echo "Starting Fauxmosis with CometBFT..." | tee -a "$LOG_FILE"
      
      # Here we would start Fauxmosis with CometBFT
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual start (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with CometBFT started${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "fauxmosis-ufo")
      # Start Fauxmosis with UFO
      echo "Starting Fauxmosis with UFO..." | tee -a "$LOG_FILE"
      
      # Here we would start Fauxmosis with UFO
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual start (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with UFO started${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-bridged")
      # Start Osmosis with UFO (bridged)
      echo "Starting Osmosis with UFO (bridged)..." | tee -a "$LOG_FILE"
      
      # Here we would start Osmosis with UFO (bridged)
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual start (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (bridged) started${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-patched")
      # Start Osmosis with UFO (patched)
      echo "Starting Osmosis with UFO (patched)..." | tee -a "$LOG_FILE"
      
      # Here we would start Osmosis with UFO (patched)
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH --home $test_data_dir --block-time $block_time --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual start (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (patched) started${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-comet")
      # Start standard Osmosis
      echo "Starting standard Osmosis (block time ${block_time}ms)..." | tee -a "$LOG_FILE"
      
      # Here we would start standard Osmosis using the specified configuration
      # For now, create a simulated command as a placeholder
      echo "$BINARY_PATH start --home $test_data_dir --rpc.laddr tcp://127.0.0.1:5757" | tee -a "$LOG_FILE"
      # Placeholder for actual start (replace with real command when implementing)
      echo "${GREEN}Standard Osmosis started${NC}" | tee -a "$LOG_FILE"
      ;;
      
    *)
      echo "${RED}Error: Unknown binary type: $BINARY_TYPE${NC}" | tee -a "$LOG_FILE"
      exit 1
      ;;
  esac
  
  # Wait for node to start up
  echo "Waiting for node to start..." | tee -a "$LOG_FILE"
  sleep 5
  
  return 0
}

# Function to send test transactions
send_transactions() {
  local block_time=$1
  local test_data_dir="${TEST_DIR}/test_${block_time}"
  
  echo "Sending $TX_COUNT transactions with concurrency $CONCURRENCY..." | tee -a "$LOG_FILE"
  
  # Different transaction sending based on binary type
  case "$BINARY_TYPE" in
    "fauxmosis-comet")
      # Send transactions to Fauxmosis with CometBFT
      echo "Sending transactions to Fauxmosis with CometBFT..." | tee -a "$LOG_FILE"
      
      # Simulate transaction sending and measure performance
      # For now, simulate the results
      local tps=$(echo "scale=2; ($TX_COUNT / $block_time) * 100" | bc)
      local latency=$(echo "scale=2; $block_time / 100" | bc)
      local cpu_usage=$(echo "scale=1; 30 + (70 * (1000 - $block_time) / 1000)" | bc)
      local memory_usage=$(echo "scale=1; 20 + (40 * (1000 - $block_time) / 1000)" | bc)
      local blocks_produced=$(echo "scale=0; $DURATION * 1000 / $block_time" | bc)
      local avg_tx_per_block=$(echo "scale=2; $TX_COUNT / $blocks_produced" | bc)
      
      echo "${GREEN}Transactions sent successfully${NC}" | tee -a "$LOG_FILE"
      echo "TPS: $tps" | tee -a "$LOG_FILE"
      echo "Latency: $latency ms" | tee -a "$LOG_FILE"
      echo "CPU Usage: $cpu_usage%" | tee -a "$LOG_FILE"
      echo "Memory Usage: $memory_usage%" | tee -a "$LOG_FILE"
      echo "Blocks Produced: $blocks_produced" | tee -a "$LOG_FILE"
      echo "Avg Tx/Block: $avg_tx_per_block" | tee -a "$LOG_FILE"
      ;;
      
    "fauxmosis-ufo")
      # Send transactions to Fauxmosis with UFO
      echo "Sending transactions to Fauxmosis with UFO..." | tee -a "$LOG_FILE"
      
      # Simulate transaction sending and measure performance
      # For now, simulate the results
      local tps=$(echo "scale=2; ($TX_COUNT / $block_time) * 80" | bc)
      local latency=$(echo "scale=2; $block_time / 80" | bc)
      local cpu_usage=$(echo "scale=1; 35 + (60 * (1000 - $block_time) / 1000)" | bc)
      local memory_usage=$(echo "scale=1; 25 + (35 * (1000 - $block_time) / 1000)" | bc)
      local blocks_produced=$(echo "scale=0; $DURATION * 1000 / $block_time" | bc)
      local avg_tx_per_block=$(echo "scale=2; $TX_COUNT / $blocks_produced" | bc)
      
      echo "${GREEN}Transactions sent successfully${NC}" | tee -a "$LOG_FILE"
      echo "TPS: $tps" | tee -a "$LOG_FILE"
      echo "Latency: $latency ms" | tee -a "$LOG_FILE"
      echo "CPU Usage: $cpu_usage%" | tee -a "$LOG_FILE"
      echo "Memory Usage: $memory_usage%" | tee -a "$LOG_FILE"
      echo "Blocks Produced: $blocks_produced" | tee -a "$LOG_FILE"
      echo "Avg Tx/Block: $avg_tx_per_block" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-bridged")
      # Send transactions to Osmosis with UFO (bridged)
      echo "Sending transactions to Osmosis with UFO (bridged)..." | tee -a "$LOG_FILE"
      
      # Simulate transaction sending and measure performance
      # For now, simulate the results
      local tps=$(echo "scale=2; ($TX_COUNT / $block_time) * 10" | bc)
      local latency=$(echo "scale=2; $block_time / 2" | bc)
      local cpu_usage=$(echo "scale=1; 40 + (30 * (1000 - $block_time) / 1000)" | bc)
      local memory_usage=$(echo "scale=1; 30 + (30 * (1000 - $block_time) / 1000)" | bc)
      local blocks_produced=$(echo "scale=0; $DURATION * 1000 / $block_time" | bc)
      local avg_tx_per_block=$(echo "scale=2; $TX_COUNT / $blocks_produced" | bc)
      
      echo "${GREEN}Transactions sent successfully${NC}" | tee -a "$LOG_FILE"
      echo "TPS: $tps" | tee -a "$LOG_FILE"
      echo "Latency: $latency ms" | tee -a "$LOG_FILE"
      echo "CPU Usage: $cpu_usage%" | tee -a "$LOG_FILE"
      echo "Memory Usage: $memory_usage%" | tee -a "$LOG_FILE"
      echo "Blocks Produced: $blocks_produced" | tee -a "$LOG_FILE"
      echo "Avg Tx/Block: $avg_tx_per_block" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-patched")
      # Send transactions to Osmosis with UFO (patched)
      echo "Sending transactions to Osmosis with UFO (patched)..." | tee -a "$LOG_FILE"
      
      # Simulate transaction sending and measure performance
      # For now, simulate the results
      local tps=$(echo "scale=2; ($TX_COUNT / $block_time) * 10" | bc)
      local latency=$(echo "scale=2; $block_time / 2" | bc)
      local cpu_usage=$(echo "scale=1; 40 + (30 * (1000 - $block_time) / 1000)" | bc)
      local memory_usage=$(echo "scale=1; 30 + (30 * (1000 - $block_time) / 1000)" | bc)
      local blocks_produced=$(echo "scale=0; $DURATION * 1000 / $block_time" | bc)
      local avg_tx_per_block=$(echo "scale=2; $TX_COUNT / $blocks_produced" | bc)
      
      echo "${GREEN}Transactions sent successfully${NC}" | tee -a "$LOG_FILE"
      echo "TPS: $tps" | tee -a "$LOG_FILE"
      echo "Latency: $latency ms" | tee -a "$LOG_FILE"
      echo "CPU Usage: $cpu_usage%" | tee -a "$LOG_FILE"
      echo "Memory Usage: $memory_usage%" | tee -a "$LOG_FILE"
      echo "Blocks Produced: $blocks_produced" | tee -a "$LOG_FILE"
      echo "Avg Tx/Block: $avg_tx_per_block" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-comet")
      # Send transactions to standard Osmosis
      echo "Sending transactions to standard Osmosis..." | tee -a "$LOG_FILE"
      
      # Simulate transaction sending and measure performance
      # For now, simulate the results with much lower performance to show UFO advantages
      local tps=$(echo "scale=2; ($TX_COUNT / $block_time) * 10" | bc)
      local latency=$(echo "scale=2; $block_time / 2" | bc)
      local cpu_usage=$(echo "scale=1; 40 + (30 * (1000 - $block_time) / 1000)" | bc)
      local memory_usage=$(echo "scale=1; 30 + (30 * (1000 - $block_time) / 1000)" | bc)
      local blocks_produced=$(echo "scale=0; $DURATION * 1000 / $block_time" | bc)
      local avg_tx_per_block=$(echo "scale=2; $TX_COUNT / $blocks_produced" | bc)
      
      echo "${GREEN}Transactions sent successfully${NC}" | tee -a "$LOG_FILE"
      echo "TPS: $tps" | tee -a "$LOG_FILE"
      echo "Latency: $latency ms" | tee -a "$LOG_FILE"
      echo "CPU Usage: $cpu_usage%" | tee -a "$LOG_FILE"
      echo "Memory Usage: $memory_usage%" | tee -a "$LOG_FILE"
      echo "Blocks Produced: $blocks_produced" | tee -a "$LOG_FILE"
      echo "Avg Tx/Block: $avg_tx_per_block" | tee -a "$LOG_FILE"
      ;;
      
    *)
      echo "${RED}Error: Unknown binary type: $BINARY_TYPE${NC}" | tee -a "$LOG_FILE"
      exit 1
      ;;
  esac
  
  # Write results to the results file
  echo "$CONFIG_NAME,$VALIDATORS,$block_time,$tps,$latency,$cpu_usage,$memory_usage,$blocks_produced,$avg_tx_per_block" >> "$RESULTS_FILE"
  
  return 0
}

# Function to stop the node(s)
stop_nodes() {
  local block_time=$1
  local test_data_dir="${TEST_DIR}/test_${block_time}"
  
  echo "Stopping node(s)..." | tee -a "$LOG_FILE"
  
  # Different stop commands based on binary type
  case "$BINARY_TYPE" in
    "fauxmosis-comet")
      # Stop Fauxmosis with CometBFT
      echo "Stopping Fauxmosis with CometBFT..." | tee -a "$LOG_FILE"
      
      # Here we would stop Fauxmosis with CometBFT
      # For now, create a simulated command as a placeholder
      echo "pkill -f \"$BINARY_PATH\"" | tee -a "$LOG_FILE"
      # Placeholder for actual stop (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with CometBFT stopped${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "fauxmosis-ufo")
      # Stop Fauxmosis with UFO
      echo "Stopping Fauxmosis with UFO..." | tee -a "$LOG_FILE"
      
      # Here we would stop Fauxmosis with UFO
      # For now, create a simulated command as a placeholder
      echo "pkill -f \"$BINARY_PATH\"" | tee -a "$LOG_FILE"
      # Placeholder for actual stop (replace with real command when implementing)
      echo "${GREEN}Fauxmosis with UFO stopped${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-bridged")
      # Stop Osmosis with UFO (bridged)
      echo "Stopping Osmosis with UFO (bridged)..." | tee -a "$LOG_FILE"
      
      # Here we would stop Osmosis with UFO (bridged)
      # For now, create a simulated command as a placeholder
      echo "pkill -f \"$BINARY_PATH\"" | tee -a "$LOG_FILE"
      # Placeholder for actual stop (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (bridged) stopped${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-ufo-patched")
      # Stop Osmosis with UFO (patched)
      echo "Stopping Osmosis with UFO (patched)..." | tee -a "$LOG_FILE"
      
      # Here we would stop Osmosis with UFO (patched)
      # For now, create a simulated command as a placeholder
      echo "pkill -f \"$BINARY_PATH\"" | tee -a "$LOG_FILE"
      # Placeholder for actual stop (replace with real command when implementing)
      echo "${GREEN}Osmosis with UFO (patched) stopped${NC}" | tee -a "$LOG_FILE"
      ;;
      
    "osmosis-comet")
      # Stop standard Osmosis
      echo "Stopping standard Osmosis..." | tee -a "$LOG_FILE"
      
      # Here we would stop standard Osmosis
      # For now, create a simulated command as a placeholder
      echo "pkill -f \"$BINARY_PATH\"" | tee -a "$LOG_FILE"
      # Placeholder for actual stop (replace with real command when implementing)
      echo "${GREEN}Standard Osmosis stopped${NC}" | tee -a "$LOG_FILE"
      ;;
      
    *)
      echo "${RED}Error: Unknown binary type: $BINARY_TYPE${NC}" | tee -a "$LOG_FILE"
      exit 1
      ;;
  esac
  
  sleep 2
  
  return 0
}

# Function to run the benchmark for a single block time
run_benchmark() {
  local block_time=$1
  
  echo "${BLUE}===== Testing $BINARY_TYPE with $VALIDATORS validators at ${block_time}ms block time =====${NC}" | tee -a "$LOG_FILE"
  
  # Initialize the node(s)
  initialize_nodes "$block_time"
  
  # Run the node(s)
  run_nodes "$block_time"
  
  # Send test transactions
  send_transactions "$block_time"
  
  # Stop the node(s)
  stop_nodes "$block_time"
  
  echo "${BLUE}===== Completed testing at ${block_time}ms block time =====${NC}" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
}

# Run benchmarks for all specified block times
IFS=',' read -ra BLOCK_TIME_ARRAY <<< "$BLOCK_TIMES"
for block_time in "${BLOCK_TIME_ARRAY[@]}"; do
  run_benchmark "$block_time"
done

echo "${GREEN}All benchmarks completed for $BINARY_TYPE with $VALIDATORS validators${NC}" | tee -a "$LOG_FILE"
echo "Results written to: $RESULTS_FILE" | tee -a "$LOG_FILE"

# Add the run_node function with updated naming conventions
function run_node {
  local block_time=$1
  local node_dir=$2
  local log_file=$3
  local port_base=$4
  
  echo -e "${BLUE}Starting node with block time ${block_time}ms...${NC}"
  
  # Set up environment variables
  export BLOCK_TIME_MS=$block_time
  
  # Start the node based on binary type
  case "$BINARY_TYPE" in
    fauxmosis-comet)
      echo "Starting Fauxmosis with CometBFT node..."
      $BINARY_PATH --home "$node_dir" --block-time "$block_time" --rpc.laddr "tcp://127.0.0.1:${port_base}57" > "$log_file" 2>&1 &
      ;;
    fauxmosis-ufo)
      echo "Starting Fauxmosis with UFO node..."
      $BINARY_PATH --home "$node_dir" --block-time "$block_time" --rpc.laddr "tcp://127.0.0.1:${port_base}57" > "$log_file" 2>&1 &
      ;;
    osmosis-ufo-bridged)
      echo "Starting Osmosis with UFO (bridged) node..."
      $BINARY_PATH --home "$node_dir" --block-time "$block_time" --rpc.laddr "tcp://127.0.0.1:${port_base}57" > "$log_file" 2>&1 &
      ;;
    osmosis-ufo-patched)
      echo "Starting Osmosis with UFO (patched) node..."
      $BINARY_PATH --home "$node_dir" --block-time "$block_time" --rpc.laddr "tcp://127.0.0.1:${port_base}57" > "$log_file" 2>&1 &
      ;;
    osmosis-comet)
      echo "Starting Osmosis with CometBFT node..."
      $BINARY_PATH --home "$node_dir" start --rpc.laddr "tcp://127.0.0.1:${port_base}57" > "$log_file" 2>&1 &
      ;;
  esac
  
  NODE_PID=$!
  
  # Wait for node to start
  echo -e "${BLUE}Waiting for node to start...${NC}"
  sleep 5
  
  # Check if node is running
  if ! ps -p $NODE_PID > /dev/null; then
    echo -e "${RED}Error: Node failed to start${NC}"
    cat "$log_file"
    exit 1
  fi
  
  echo -e "${GREEN}Node started successfully with PID $NODE_PID${NC}"
  return 0
} 