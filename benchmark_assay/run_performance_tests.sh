#!/bin/bash
# Script to build and run performance tests for Osmosis with UFO integration

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Default configuration
BUILD_DIR="/tmp/osmosis-ufo-build"
OSMOSIS_UFO_PATCHED_DIR="/tmp/osmosis-ufo-patched-build"
OSMOSIS_UFO_PATCHED_BINARY="$OSMOSIS_UFO_PATCHED_DIR/osmosisd-ufo"
OSMOSIS_UFO_BRIDGED_DIR="/tmp/osmosis-ufo-bridged-build"
OSMOSIS_UFO_BRIDGED_BINARY="$OSMOSIS_UFO_BRIDGED_DIR/osmosis-ufo-bridged"
OSMOSIS_COMET_DIR="/tmp/osmosis-comet-build"
OSMOSIS_COMET_BINARY="$OSMOSIS_COMET_DIR/osmosisd"
FAUXMOSIS_COMET_DIR="/tmp/fauxmosis-comet-build"
FAUXMOSIS_COMET_BINARY="$FAUXMOSIS_COMET_DIR/fauxmosis-comet"
FAUXMOSIS_UFO_DIR="/tmp/fauxmosis-ufo-build"
FAUXMOSIS_UFO_BINARY="$FAUXMOSIS_UFO_DIR/fauxmosis-ufo"
TEST_DIR="$PROJECT_DIR/benchmark_results"
BLOCK_TIMES="1000,500,200,100,50,20,10,5,1"
TX_COUNT=2000
DURATION=120
CONCURRENCY=20
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RUN_NAME="benchmark_run_${TIMESTAMP}"
CONFIGURATIONS="all"  # Default to run all configurations

# Parse arguments
function print_usage {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --build-dir DIR       Directory for building binaries (default: $BUILD_DIR)"
  echo "  --osmosis-comet-dir DIR Directory for Osmosis with CometBFT (default: $OSMOSIS_COMET_DIR)"
  echo "  --osmosis-ufo-patched-dir DIR Directory for patched Osmosis with UFO (default: $OSMOSIS_UFO_PATCHED_DIR)"
  echo "  --osmosis-ufo-bridged-dir DIR Directory for bridged Osmosis with UFO (default: $OSMOSIS_UFO_BRIDGED_DIR)"
  echo "  --fauxmosis-comet-dir DIR     Directory for Fauxmosis with CometBFT (default: $FAUXMOSIS_COMET_DIR)"
  echo "  --fauxmosis-ufo-dir DIR Directory for Fauxmosis with UFO (default: $FAUXMOSIS_UFO_DIR)"
  echo "  --test-dir DIR        Directory for benchmark results (default: $TEST_DIR)"
  echo "  --block-times LIST    Comma-separated list of block times in ms (default: $BLOCK_TIMES)"
  echo "  --tx-count COUNT      Number of transactions to send (default: $TX_COUNT)"
  echo "  --duration SECONDS    Duration of each test in seconds (default: $DURATION)"
  echo "  --concurrency NUM     Number of concurrent transactions (default: $CONCURRENCY)"
  echo "  --run-name NAME       Name for this benchmark run (default: $RUN_NAME)"
  echo "  --configurations LIST Comma-separated list of configurations to test:"
  echo "                         fauxmosis-comet-1, fauxmosis-comet-4, fauxmosis-ufo-1, fauxmosis-ufo-4,"
  echo "                         osmosis-ufo-bridged-1, osmosis-ufo-bridged-4,"
  echo "                         osmosis-ufo-patched-1, osmosis-ufo-patched-4,"
  echo "                         osmosis-comet-1, osmosis-comet-4, or 'all' (default: all)"
  echo "  --help                Display this help message"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build-dir)
      BUILD_DIR="$2"
      shift 2
      ;;
    --osmosis-comet-dir)
      OSMOSIS_COMET_DIR="$2"
      shift 2
      ;;
    --osmosis-ufo-patched-dir)
      OSMOSIS_UFO_PATCHED_DIR="$2"
      shift 2
      ;;
    --osmosis-ufo-bridged-dir)
      OSMOSIS_UFO_BRIDGED_DIR="$2"
      shift 2
      ;;
    --fauxmosis-comet-dir)
      FAUXMOSIS_COMET_DIR="$2"
      shift 2
      ;;
    --fauxmosis-ufo-dir)
      FAUXMOSIS_UFO_DIR="$2"
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
    --run-name)
      RUN_NAME="$2"
      shift 2
      ;;
    --configurations)
      CONFIGURATIONS="$2"
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

# Set up directories and files
RUN_DIR="${TEST_DIR}/${RUN_NAME}"
LOG_FILE="${RUN_DIR}/benchmark.log"
RESULTS_FILE="${RUN_DIR}/results.csv"

# Print configuration
echo "============================================"
echo "= UFO Performance Benchmark"
echo "============================================"
echo "UFO-patched Osmosis directory: $BUILD_DIR"
echo "Standard Osmosis directory: $OSMOSIS_COMET_DIR"
echo "UFO CLI directory: $FAUXMOSIS_COMET_DIR"
echo "Test directory: $RUN_DIR"
echo "Block times (ms): $BLOCK_TIMES"
echo "Transaction count: $TX_COUNT"
echo "Test duration: $DURATION seconds"
echo "Concurrency: $CONCURRENCY"
echo "Configurations: $CONFIGURATIONS"
echo "Benchmark run name: $RUN_NAME"
echo "============================================"

# Check if necessary tools exist
echo "Checking prerequisites..."
for cmd in jq bc curl; do
  if ! command -v $cmd &> /dev/null; then
    echo "Error: Required command '$cmd' not found"
    exit 1
  fi
done

# Clean up existing files and create new directories
echo "Setting up benchmark environment..."
mkdir -p "$RUN_DIR"

# Remove any existing test data for this run
rm -rf "${RUN_DIR}/test_"*

# Initialize log file
echo "# UFO Benchmark - $(date)" > "$LOG_FILE"
echo "Starting benchmark run: $RUN_NAME" | tee -a "$LOG_FILE"
echo "Results will be saved to: $RUN_DIR" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Initialize results CSV with headers
echo "configuration,validators,block_time,tps,latency_ms,cpu_usage,memory_usage,blocks_produced,avg_tx_per_block" > "$RESULTS_FILE"

# Step 1: Build all necessary binaries
echo "Step 1: Building necessary binaries" | tee -a "$LOG_FILE"
echo "--------------------------------------------" | tee -a "$LOG_FILE"

# Build UFO-patched Osmosis
if [[ "$CONFIGURATIONS" == "all" || "$CONFIGURATIONS" == *"osmosis-ufo-patched"* || "$CONFIGURATIONS" == *"osmosis-ufo-bridged"* ]]; then
  echo "Building Osmosis with UFO integration..." | tee -a "$LOG_FILE"
  
  if [ -f "$PROJECT_DIR/result/bin/build-osmosis-ufo" ]; then
    echo "Using pre-built build-osmosis-ufo script" | tee -a "$LOG_FILE"
    BUILD_SCRIPT="$PROJECT_DIR/result/bin/build-osmosis-ufo"
  else
    echo "Building build-osmosis-ufo script with Nix" | tee -a "$LOG_FILE"
    nix build -L .#build-osmosis-ufo
    BUILD_SCRIPT="$PROJECT_DIR/result/bin/build-osmosis-ufo"
  fi

  if [ ! -f "$BUILD_SCRIPT" ]; then
    echo "Error: Could not locate build-osmosis-ufo script" | tee -a "$LOG_FILE"
    exit 1
  fi

  echo "Building Osmosis with UFO integration at $BUILD_DIR" | tee -a "$LOG_FILE"
  rm -rf "$BUILD_DIR"
  $BUILD_SCRIPT "$BUILD_DIR" | tee -a "$LOG_FILE"

  if [ ! -f "$OSMOSIS_UFO_PATCHED_BINARY" ]; then
    echo "Error: Failed to build Osmosis UFO binary at $OSMOSIS_UFO_PATCHED_BINARY" | tee -a "$LOG_FILE"
    exit 1
  fi

  echo "UFO-patched Osmosis build successful! Binary located at: $OSMOSIS_UFO_PATCHED_BINARY" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
fi

# Build standard Osmosis
if [[ "$CONFIGURATIONS" == "all" || "$CONFIGURATIONS" == *"osmosis-comet"* ]]; then
  echo "Building standard Osmosis..." | tee -a "$LOG_FILE"
  
  # Clone and build standard Osmosis
  mkdir -p "$OSMOSIS_COMET_DIR"
  if [ ! -d "$OSMOSIS_COMET_DIR/.git" ]; then
    echo "Cloning Osmosis repository..." | tee -a "$LOG_FILE"
    git clone --depth 1 https://github.com/osmosis-labs/osmosis.git "$OSMOSIS_COMET_DIR" | tee -a "$LOG_FILE"
  fi
  
  cd "$OSMOSIS_COMET_DIR"
  echo "Building standard Osmosis binary..." | tee -a "$LOG_FILE"
  
  # Use Nix to build if available
  if command -v nix &> /dev/null; then
    echo "Building using Nix environment..." | tee -a "$LOG_FILE"
    # Create a temporary build script
    TEMP_BUILD_SCRIPT=$(mktemp)
    cat > "$TEMP_BUILD_SCRIPT" << 'EOF'
#!/bin/bash
set -e
cd $1
make install
cp "$(which osmosisd)" "$1/osmosisd"
EOF
    chmod +x "$TEMP_BUILD_SCRIPT"
    
    # Run the build script in the Nix environment
    nix develop -c "$TEMP_BUILD_SCRIPT" "$OSMOSIS_COMET_DIR"
    rm "$TEMP_BUILD_SCRIPT"
  else
    # Fallback to direct make
    make install | tee -a "$LOG_FILE"
    cp "$(which osmosisd)" "$OSMOSIS_COMET_BINARY"
  fi
  
  if [ ! -f "$OSMOSIS_COMET_BINARY" ]; then
    echo "Error: Failed to build standard Osmosis binary at $OSMOSIS_COMET_BINARY" | tee -a "$LOG_FILE"
    exit 1
  fi
  
  echo "Standard Osmosis build successful! Binary located at: $OSMOSIS_COMET_BINARY" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
  cd "$PROJECT_DIR"
fi

# Build Fauxmosis with CometBFT
if [[ "$CONFIGURATIONS" == "all" || "$CONFIGURATIONS" == *"fauxmosis-comet"* ]]; then
  echo "Building Fauxmosis with CometBFT..." | tee -a "$LOG_FILE"
  
  mkdir -p "$FAUXMOSIS_COMET_DIR"
  
  # Build Fauxmosis using Nix or Go directly
  if command -v nix &> /dev/null; then
    echo "Building Fauxmosis with Nix..." | tee -a "$LOG_FILE"
    nix build -L .#fauxmosis-comet
    cp "$PROJECT_DIR/result/bin/fauxmosis-comet" "$FAUXMOSIS_COMET_BINARY"
  else
    echo "Building Fauxmosis with Go..." | tee -a "$LOG_FILE"
    go build -o "$FAUXMOSIS_COMET_BINARY" "$PROJECT_DIR/cmd/fauxmosis-comet/main.go"
  fi
  
  if [ ! -f "$FAUXMOSIS_COMET_BINARY" ]; then
    echo "Error: Failed to build Fauxmosis with CometBFT binary at $FAUXMOSIS_COMET_BINARY" | tee -a "$LOG_FILE"
    exit 1
  fi
  
  echo "Fauxmosis with CometBFT build successful! Binary located at: $FAUXMOSIS_COMET_BINARY" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
fi

# Build Fauxmosis with UFO
if [[ "$CONFIGURATIONS" == "all" || "$CONFIGURATIONS" == *"fauxmosis-ufo"* ]]; then
  echo "Building Fauxmosis with UFO..." | tee -a "$LOG_FILE"
  
  mkdir -p "$FAUXMOSIS_UFO_DIR"
  
  # Build Fauxmosis-UFO using Nix or Go directly
  if command -v nix &> /dev/null; then
    echo "Building Fauxmosis-UFO with Nix..." | tee -a "$LOG_FILE"
    nix build -L .#fauxmosis-ufo
    cp "$PROJECT_DIR/result/bin/fauxmosis-ufo" "$FAUXMOSIS_UFO_BINARY"
  else
    echo "Building Fauxmosis-UFO with Go..." | tee -a "$LOG_FILE"
    go build -o "$FAUXMOSIS_UFO_BINARY" "$PROJECT_DIR/cmd/fauxmosis-ufo/main.go"
  fi
  
  if [ ! -f "$FAUXMOSIS_UFO_BINARY" ]; then
    echo "Error: Failed to build Fauxmosis-UFO binary at $FAUXMOSIS_UFO_BINARY" | tee -a "$LOG_FILE"
    exit 1
  fi
  
  echo "Fauxmosis-UFO build successful! Binary located at: $FAUXMOSIS_UFO_BINARY" | tee -a "$LOG_FILE"
  echo "" | tee -a "$LOG_FILE"
fi

# Step 2: Run performance tests for each configuration
echo "Step 2: Running performance tests" | tee -a "$LOG_FILE"
echo "--------------------------------" | tee -a "$LOG_FILE"

# Run the benchmark for a specific configuration
function run_benchmark {
  BINARY_TYPE=$1
  BINARY_PATH=$2
  VALIDATOR_COUNT=$3
  CONFIG_NAME="${BINARY_TYPE}-${VALIDATOR_COUNT}"

  echo "Running benchmark for configuration: $CONFIG_NAME" | tee -a "$LOG_FILE"
  
  # Create directory for this configuration's results
  CONFIG_DIR="$RESULTS_DIR/$CONFIG_NAME"
  mkdir -p "$CONFIG_DIR"
  
  # Run benchmark for this configuration
  "$SCRIPT_DIR/benchmark_node.sh" \
    --binary-type "$BINARY_TYPE" \
    --binary-path "$BINARY_PATH" \
    --validators "$VALIDATOR_COUNT" \
    --block-times "$BLOCK_TIMES" \
    --tx-count "$TX_COUNT" \
    --duration "$DURATION" \
    --concurrency "$CONCURRENCY" \
    --output-dir "$CONFIG_DIR" \
    --run-name "$RUN_NAME"
  
  # Add results to consolidated CSV  
  if [ -f "$CONFIG_DIR/results.csv" ]; then
    # Skip header if consolidated.csv already exists
    if [ -f "$RESULTS_DIR/results.csv" ]; then
      tail -n +2 "$CONFIG_DIR/results.csv" >> "$RESULTS_DIR/results.csv"
    else
      cp "$CONFIG_DIR/results.csv" "$RESULTS_DIR/results.csv"
    fi
  fi
  
  echo "Benchmark for $CONFIG_NAME completed!" | tee -a "$LOG_FILE"
  echo "----------------------------------------" | tee -a "$LOG_FILE"
}

# Function to determine if a specific configuration should be run
function should_run_config {
  local config="$1"
  
  if [ "$CONFIGURATIONS" == "all" ]; then
    return 0
  else
    IFS="," read -ra CONFIG_LIST <<< "$CONFIGURATIONS"
    for c in "${CONFIG_LIST[@]}"; do
      if [ "$c" == "$config" ]; then
        return 0
      fi
    done
    return 1
  fi
}

# Run benchmarks for each selected configuration
echo "" | tee -a "$LOG_FILE"
echo "Step 3: Running benchmarks" | tee -a "$LOG_FILE"
echo "----------------------------------------" | tee -a "$LOG_FILE"

# Fauxmosis with CometBFT
if should_run_config "fauxmosis-comet-1"; then
  run_benchmark "fauxmosis-comet" "$FAUXMOSIS_COMET_BINARY" 1
fi

# Fauxmosis with CometBFT
if should_run_config "fauxmosis-comet-4"; then
  run_benchmark "fauxmosis-comet" "$FAUXMOSIS_COMET_BINARY" 4
fi

# Fauxmosis with UFO
if should_run_config "fauxmosis-ufo-1"; then
  run_benchmark "fauxmosis-ufo" "$FAUXMOSIS_UFO_BINARY" 1
fi

# Fauxmosis with UFO
if should_run_config "fauxmosis-ufo-4"; then
  run_benchmark "fauxmosis-ufo" "$FAUXMOSIS_UFO_BINARY" 4
fi

# Osmosis with UFO (patched) with 1 validator
if should_run_config "osmosis-ufo-patched-1"; then
  run_benchmark "osmosis-ufo-patched" "$OSMOSIS_UFO_PATCHED_BINARY" 1
fi

# Osmosis with UFO (patched) with 4 validators
if should_run_config "osmosis-ufo-patched-4"; then
  run_benchmark "osmosis-ufo-patched" "$OSMOSIS_UFO_PATCHED_BINARY" 4
fi

# Osmosis with UFO (bridged) with 1 validator
if should_run_config "osmosis-ufo-bridged-1"; then
  run_benchmark "osmosis-ufo-bridged" "$OSMOSIS_UFO_BRIDGED_BINARY" 1
fi

# Osmosis with UFO (bridged) with 4 validators
if should_run_config "osmosis-ufo-bridged-4"; then
  run_benchmark "osmosis-ufo-bridged" "$OSMOSIS_UFO_BRIDGED_BINARY" 4
fi

# Osmosis with CometBFT with 1 validator
if should_run_config "osmosis-comet-1"; then
  run_benchmark "osmosis-comet" "$OSMOSIS_COMET_BINARY" 1
fi

# Osmosis with CometBFT with 4 validators
if should_run_config "osmosis-comet-4"; then
  run_benchmark "osmosis-comet" "$OSMOSIS_COMET_BINARY" 4
fi

# Step 3: Generate Jupyter notebook
echo "Step 3: Generating Jupyter notebook" | tee -a "$LOG_FILE"
echo "--------------------------------" | tee -a "$LOG_FILE"

# Use Nix if available
if command -v nix &> /dev/null; then
  echo "Using Nix for notebook generation..." | tee -a "$LOG_FILE"
  
  # Generate the notebook using the Nix-integrated Jupyter
  nix run .#python-viz -c python "$SCRIPT_DIR/generate_benchmark_notebook.py" "$RESULTS_FILE" "$RUN_DIR" --run-name "$RUN_NAME" | tee -a "$LOG_FILE"
else
  # Fallback to system Python with package checks
  echo "Nix not available, checking system Python installation..." | tee -a "$LOG_FILE"
  
  if ! python -c "import pandas" &>/dev/null || ! python -c "import matplotlib" &>/dev/null; then
    echo "Warning: Required Python packages not found" | tee -a "$LOG_FILE"
    echo "Installing required Python packages..." | tee -a "$LOG_FILE"
    pip install pandas matplotlib numpy ipython notebook seaborn ipykernel jupyter_client | tee -a "$LOG_FILE"
  fi
  
  # Generate the notebook
  python "$SCRIPT_DIR/generate_benchmark_notebook.py" "$RESULTS_FILE" "$RUN_DIR" --run-name "$RUN_NAME" | tee -a "$LOG_FILE"
fi

echo
echo "Benchmark completed successfully!" | tee -a "$LOG_FILE"
echo "Jupyter notebook generated at: $RUN_DIR/benchmark_analysis.ipynb" | tee -a "$LOG_FILE"
echo
echo "To view the notebook, run:"
echo "nix run .#notebook $RUN_DIR/benchmark_analysis.ipynb"
echo
echo "Or for a more interactive experience:"
echo "nix develop .#jupyter"
echo
echo "To run another benchmark with different configurations, use:"
echo "./benchmark_assay/run_performance_tests.sh --configurations fauxmosis-comet-1,osmosis-ufo-4,osmosis-comet-1"
echo

# Display example usage of specific configurations
echo "Example of testing specific configurations:"
echo "./benchmark_assay/run_performance_tests.sh --configurations fauxmosis-comet-1,osmosis-ufo-4,osmosis-comet-1"
echo 