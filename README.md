![](./ufo.png)

# UFO (Universal Fast Orderer)

UFO is an in-memory mock consensus engine designed as a drop-in replacement for [CometBFT](https://github.com/cometbft/cometbft) to accelerate multi-chain simulation. When testing complex multi-chain interactions, the standard CometBFT consensus imposes significant overhead. UFO provides a fast alternative for testing Cosmos applications at sub-millisecond block times.

## Features

- Higher transaction throughput
- Low latency
- Support for ultra-low block times (down to sub-millisecond)
- Resource efficient

## Integration Modes

UFO offers multiple integration modes for different use cases. The Osmosis node is provided here as a practical test case because it uses the [Block SDK](https://github.com/skip-mev/block-sdk) which uses ABCI++ and it's a sovereign blockchain, making it easier to spin up for testing.

### 1. Mock Cosmos SDK (Fauxmosis)

Fauxmosis is a lightweight mock Cosmos SDK application designed for testing and benchmarking. It provides:
- Minimal dependencies
- Simplified application logic
- Fast startup and testing
- Focused performance benchmarking

### 2. UFO Bridged Mode

In bridged mode, UFO connects to a running Osmosis instance through adapter interfaces:
- UFO and Osmosis run as separate processes
- Communication occurs through defined interfaces
- Minimal changes to Osmosis codebase
- Easier to update components independently
- Suitable for development and testing

### 3. UFO Patched Mode

Patch-based integration directly modifies the Osmosis source code:
- Applies patches to replace CometBFT with UFO at the source level
- Creates a single integrated binary
- More efficient performance (no inter-process communication)
- Deeper integration at the code level
- Suitable for production deployments

### 4. CometBFT (Baseline)

Standard Osmosis with CometBFT serves as the baseline for performance comparisons:
- Original implementation without UFO
- Reference point for benchmarking
- Used to validate correctness and measure improvements

## Integration with Osmosis

UFO can be integrated with Osmosis using the provided adapter. The adapter creates a shim implementation of the CometBFT interfaces, allowing Osmosis to use UFO without extensive code modifications.

## Building Osmosis with UFO Integration

### Patch-Based Integration

You can build Osmosis with UFO patch-based integration using our Nix-integrated build script:

```bash
nix run .#build-osmosis -- /path/to/osmosis/source
```

This will create a binary called `osmosisd-ufo` in the Osmosis directory that uses UFO for consensus instead of CometBFT.

### Bridged Integration

To build the bridged mode version:

```bash
go build -o osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged/main.go
```

## Performance Testing

UFO includes a comprehensive performance testing suite that benchmarks all integration modes against each other.

To run the full performance test suite with the default configurations:

```bash
./benchmark_assay/run_performance_tests.sh
```

To run tests with specific parameters (block times, transaction count, and duration):

```bash
./benchmark_assay/run_performance_tests.sh --block-times 1000,100,10,1 --tx-count 5000 --duration 180
```

### Test Configurations

The following configurations can be tested:

```bash
./benchmark_assay/run_performance_tests.sh --configurations fauxmosis-comet-1,osmosis-ufo-bridged-1,osmosis-ufo-patched-1,osmosis-comet-1
```

Available configurations:
- `fauxmosis-comet-1`: Fauxmosis with CometBFT (1 validator)
- `fauxmosis-comet-4`: Fauxmosis with CometBFT (4 validators)
- `fauxmosis-ufo-1`: Fauxmosis with UFO (1 validator)
- `fauxmosis-ufo-4`: Fauxmosis with UFO (4 validators)
- `osmosis-ufo-bridged-1`: Osmosis with UFO (bridged mode, 1 validator)
- `osmosis-ufo-bridged-4`: Osmosis with UFO (bridged mode, 4 validators)
- `osmosis-ufo-patched-1`: Osmosis with UFO (patched mode, 1 validator)
- `osmosis-ufo-patched-4`: Osmosis with UFO (patched mode, 4 validators)
- `osmosis-comet-1`: Osmosis with CometBFT (1 validator)
- `osmosis-comet-4`: Osmosis with CometBFT (4 validators)

### Options

The performance testing suite supports several options:
- `--build-dir`: Directory to build binaries in (default: build)
- `--osmosis-comet-dir`: Directory with Osmosis+CometBFT source (default: /tmp/osmosis-comet-build)
- `--osmosis-ufo-patched-dir`: Directory with patched Osmosis+UFO (default: /tmp/osmosis-ufo-patched-build)
- `--osmosis-ufo-bridged-dir`: Directory for bridged Osmosis+UFO (default: /tmp/osmosis-ufo-bridged-build)
- `--fauxmosis-comet-dir`: Directory for Fauxmosis with CometBFT (default: /tmp/fauxmosis-comet-build)
- `--fauxmosis-ufo-dir`: Directory for Fauxmosis with UFO (default: /tmp/fauxmosis-ufo-build)
- `--configurations`: Comma-separated list of configurations to test
- `--block-times`: Comma-separated list of block times to test (in ms)
- `--tx-count`: Number of transactions to send
- `--duration`: Duration of each test (in seconds)
- `--visualize-only`: Skip tests and only visualize results
- `--no-visualize`: Skip visualization
- `--run-name`: Specify a name for this test run (default: benchmark_run_YYYYMMDD_HHMMSS)

### Results

Test results are saved in the `benchmark_results` directory, with each run getting its own subdirectory. Consolidated results are available in:
- `benchmark_results/results.csv`: All test results in CSV format
- `benchmark_results/notebook_visualizations/`: Visualization images
- `benchmark_results/benchmark_analysis.ipynb`: Jupyter notebook with analysis

### Visualizations

Visualizations are generated automatically after tests complete. You can regenerate them using:

```bash
./benchmark_assay/visualize_benchmark.py benchmark_results/results.csv benchmark_results/
```

### Quick Comparative Benchmark

To run a quick comparative benchmark between all implementation modes:

```bash
./benchmark_assay/run_quick_comparative_benchmark.sh
```

### Cleanup

To clean up benchmark results but keep specific runs:

```bash
./benchmark_assay/cleanup_benchmark_results.sh --keep-run benchmark_run_20250315_123045
```

### Individual Node Benchmarking

For benchmarking individual nodes with specific parameters:

```bash
./benchmark_assay/benchmark_node.sh --binary-type fauxmosis-ufo --binary-path /path/to/fauxmosis-ufo --validators 1 --block-times 100,10,1 --tx-count 1000
```

### Benchmarking Environment

For a dedicated benchmarking environment with all dependencies:

```bash
./benchmark_assay/run_benchmark_env.sh
```

## How It Works

### Fauxmosis Integration

The Fauxmosis mock application demonstrates UFO integration with a simplified Cosmos SDK app:
1. **Mock Application**: Provides a lightweight implementation of Cosmos SDK app
2. **UFO Integration**: Shows how UFO replaces CometBFT in a Cosmos SDK context
3. **Benchmarking**: Enables performance testing without external dependencies

### Osmosis Bridged Integration

The bridged integration approach consists of the following components:
1. **Bridge Client**: Connects UFO to a running Osmosis instance
2. **Adapter Layer**: Translates between UFO and CometBFT interfaces
3. **Communication Protocol**: Defines how the two systems interact
4. **Independent Processes**: Allows UFO and Osmosis to run separately

### Osmosis Patched Integration

The patch-based integration approach consists of the following components:
1. **Adapter Interface**: The `OsmosisUFOAdapter` type provides a bridge between Osmosis and UFO
2. **Source Patches**: Direct modifications to the Osmosis codebase
3. **Dependency Replacement**: Swaps CometBFT dependencies with UFO implementations
4. **Single Binary**: Produces an integrated application with UFO consensus

## Performance Comparison

The benchmark suite allows direct comparison between all integration modes:
- **Transactions per Second (TPS)**: Measure of throughput
- **Latency**: Transaction confirmation time
- **Resource Usage**: CPU and memory consumption
- **Block Times**: Performance at different block production rates
- **Scalability**: Performance with increasing validator counts

## Development

### Prerequisites

- Go 1.22 or later
- Nix (optional, for reproducible builds)
- Python with pandas and matplotlib (for performance visualization)

### Building

```bash
# Using Go directly
go build -o fauxmosis-comet ./cmd/fauxmosis-comet/main.go
go build -o fauxmosis-ufo ./cmd/fauxmosis-ufo/main.go
go build -o osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged/main.go

# Using Nix
nix build
```

### Development Shell

```bash
# General development
nix develop

# Benchmark-specific environment with visualization tools
nix develop .#benchmark
# or
./benchmark_assay/run_benchmark_env.sh
```

### 🛸