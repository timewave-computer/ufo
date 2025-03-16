# UFO Benchmarking Guide

UFO was created to accelerate Timewave's multi-chain test suite. When testing complex multi-chain interactions, CometBFT is one of the largest sources of overhead.

The benchmarking tools in this repository allow developers to:

1. Test performance across different block times (from standard 1000ms down to extremely low 1ms)
2. Compare UFO's performance to standard CometBFT at each block time setting
3. Test different integration modes (Fauxmosis, Bridged, Patched)
4. Evaluate performance with different validator configurations (1 or 4 validators)
5. Generate detailed reports and visualizations of the results
6. Make data-driven decisions about optimal settings for your specific use case

## Benchmark Types

UFO provides multiple benchmark types to evaluate different aspects of performance:

### Block Time Benchmarks

These benchmarks focus on how performance changes across different block time settings, from the standard 1000ms down to ultra-fast 1ms block times.

Key metrics measured include:
- Transaction throughput (transactions per second)
- Transaction latency
- Success rates
- Resource usage (CPU, memory)

### UFO vs CometBFT Benchmarks

These benchmarks directly compare UFO against standard CometBFT when running identical workloads.

The comparison includes:
- Transaction throughput
- Latency
- Resource efficiency 
- Stability under load

### Configuration Benchmarks

These benchmarks compare different integration approaches:

1. **Fauxmosis with CometBFT**: Baseline using mock Cosmos SDK with standard CometBFT
2. **Fauxmosis with UFO**: Mock Cosmos SDK with UFO consensus
3. **Osmosis with UFO (Bridged)**: Real Osmosis connected to UFO via bridge
4. **Osmosis with UFO (Patched)**: Real Osmosis with UFO integrated via patching
5. **Osmosis with CometBFT**: Baseline using real Osmosis with standard CometBFT

Each configuration can be tested with either a single validator or a network of 4 validators.

## Benchmarking Scripts

Several scripts are provided for different benchmarking scenarios:

1. **benchmark_node.sh**: Core script for benchmarking a single node type
2. **run_performance_tests.sh**: Comprehensive script for testing all configurations
3. **run_quick_comparative_benchmark.sh**: Simplified benchmark for quick comparisons
4. **visualize_benchmark.py**: Python script for generating visualizations

## Running Benchmarks

### Quick Comparative Benchmark

For a quick overview of all configurations:

```bash
./benchmark_assay/run_quick_comparative_benchmark.sh
```

This will:
- Run tests with 3 representative block times (1000ms, 100ms, 10ms)
- Test all 10 configurations (5 binary types × 2 validator counts)
- Use a small transaction count for quick results
- Generate a comparative visualization

### Full Performance Benchmark

For comprehensive performance testing:

```bash
./benchmark_assay/run_performance_tests.sh
```

This will:
- Run tests with 9 block times (1000ms to 1ms)
- Test all configurations 
- Use a larger transaction count
- Generate detailed performance reports

### Custom Benchmark Options

You can customize the benchmarks with various options:

```bash
./benchmark_assay/run_performance_tests.sh \
  --block-times 1000,500,200,100,50,20,10,5,2,1 \
  --tx-count 5000 \
  --duration 300 \
  --configurations fauxmosis-ufo-1,osmosis-ufo-patched-1,osmosis-comet-1
```

Where:
- `--block-times`: Comma-separated list of block times to test (in milliseconds)
- `--tx-count`: Number of transactions to process in each test
- `--duration`: Duration of each test in seconds
- `--configurations`: Specific configurations to test
  - Format: `<binary-type>-<validator-count>`
  - Options: fauxmosis-comet-1, fauxmosis-comet-4, fauxmosis-ufo-1, fauxmosis-ufo-4, osmosis-ufo-bridged-1, osmosis-ufo-bridged-4, osmosis-ufo-patched-1, osmosis-ufo-patched-4, osmosis-comet-1, osmosis-comet-4

### Individual Node Benchmarking

For testing a specific binary type and configuration:

```bash
./benchmark_assay/benchmark_node.sh \
  --binary-type fauxmosis-ufo \
  --binary-path /path/to/fauxmosis-ufo \
  --validators 1 \
  --block-times 100,10,1 \
  --tx-count 1000
```

Where:
- `--binary-type`: Type of binary to test
- `--binary-path`: Path to the binary being tested
- `--validators`: Number of validators (1 or 4)
- `--block-times`: Comma-separated list of block times to test
- `--tx-count`: Number of transactions to send

## Understanding Benchmark Output

The benchmarking process generates several outputs:

### Results CSV

A CSV file containing all benchmark results is saved at `benchmark_results/<run_name>/results.csv`. This file includes:

- Configuration details (binary type, validator count)
- Block time
- Transactions per second (TPS)
- Latency in milliseconds
- CPU usage percentage
- Memory usage percentage
- Blocks produced
- Average transactions per block

### Jupyter Notebook

A comprehensive Jupyter notebook is generated at `benchmark_results/<run_name>/benchmark_analysis.ipynb`, which includes:

- Data loading and preparation
- Performance visualizations
- Comparative analysis
- Insights and conclusions

To open the notebook:

```bash
nix run .#notebook "benchmark_results/<run_name>/benchmark_analysis.ipynb"
```

Or using the development environment:

```bash
nix develop .#jupyter
```

### Visualization

Several visualization charts are generated in the `benchmark_results/<run_name>/notebook_visualizations/` directory:

1. **TPS by Block Time**: Transactions per second at different block times
2. **TPS by Configuration**: Comparative transaction throughput across configurations
3. **Latency by Block Time**: Transaction latency at different block times
4. **CPU Usage by Configuration**: CPU usage comparison
5. **Memory Usage by Configuration**: Memory usage comparison
6. **Performance Heatmap**: Comprehensive comparison of all metrics

## Interpreting Results

When analyzing benchmark results, consider these patterns:

### Block Time Impact

1. **Transaction Throughput (TPS)**:
   - UFO typically shows higher TPS than CometBFT, with the gap widening at lower block times
   - TPS generally increases as block time decreases until hardware/network limitations are reached

2. **Success Rate**:
   - CometBFT typically shows degraded success rates at block times below 100ms
   - UFO maintains high success rates even at extremely low block times

3. **Latency**:
   - Transaction latency should decrease with block time but may increase if the system becomes overloaded
   - UFO generally shows lower latency than CometBFT at all block times

4. **Resource Usage**:
   - CometBFT typically shows higher CPU/memory usage at lower block times
   - UFO's resource usage increases more gradually as block time decreases

## Network Simulation

To simulate network conditions like latency and packet loss, you can use tools like `tc` before running the benchmark:

```bash
# Add 100ms latency to localhost
sudo tc qdisc add dev lo root netem delay 100ms

# Run the benchmark
./benchmark_assay/run_performance_tests.sh

# Remove the network simulation
sudo tc qdisc del dev lo root
```

## Long-Running Tests

For more realistic benchmark results in production-like scenarios:

1. Increase `--tx-count` to 100,000 or more
2. Increase `--duration` to 3600 (1 hour) or more
3. Use the `--run-name` option to identify the long-running test
