# UFO Operation Guide

This document provides a comprehensive guide on operating and configuring the Universal Fast Orderer (UFO) system.

## Block Time Configuration

Block time is the average time interval between blocks being added to a blockchain. In traditional CometBFT/Tendermint-based chains:

- The default block time is typically around 1000ms (1 second)
- This setting balances throughput with network stability and consensus overhead
- Reducing block time below certain thresholds in CometBFT can lead to performance degradation

### UFO Block Times

UFO has been designed to efficiently operate at much shorter block times than CometBFT:

- Can operate stably at block times as low as 1-10ms
- Maintains high transaction success rates even at extremely low block times
- Shows minimal resource usage increases when block times are reduced
- Retains network stability and consensus guarantees at lower block times

### Performance Impact of Block Time Settings

Our benchmarks show the following general patterns when comparing UFO to CometBFT across different block times:

| Block Time | CometBFT Performance | UFO Performance |
|------------|----------------------|-----------------|
| 1000ms     | Stable, good throughput | Stable, slightly better throughput |
| 100ms      | Degraded success rate, increased resource usage | Stable, good throughput |
| 10ms       | Significant degradation | Minimal performance impact |
| 1ms        | Severe degradation or failure | Operational with moderate performance impact |

### Configuring Block Time in UFO

Block time in UFO is configured through the consensus parameters in the chain's genesis file or through governance parameters:

```json
{
  "consensus_params": {
    "block": {
      "time_iota_ms": "10"  // Block time in milliseconds
    }
  }
}
```

You can also configure this in your application's configuration file:

```toml
[consensus]
timeout_commit = "10ms"  # Equivalent to block time
```

## Validator Configuration

UFO supports multiple validator configurations to suit different use cases:

### Single Validator Mode

- Simplest configuration with a single validator
- Ideal for development, testing, and performance benchmarking
- Configured through the `--validators 1` flag in benchmark scripts

### Multi-Validator Mode

- Supports a set of validators with configurable voting power
- Currently supports up to 4 validators in standard configurations
- More realistic simulation of network conditions and consensus
- Configured through the `--validators 4` flag in benchmark scripts

### Validator Management

Validators in UFO are defined by:

```go
type Validator struct {
    ID           string  // Unique identifier
    Address      string  // Validator address
    VotingPower  int64   // Relative voting power
}
```

## Performance Tuning Parameters

UFO offers several parameters that can be adjusted to optimize performance:

### Transaction Processing

- **Transaction Pool Size**: Controls how many transactions can be queued
- **Batch Size**: Number of transactions processed in a single block
- **Concurrency Level**: Number of concurrent transaction processors

### Memory Management

- **State Cache Size**: Controls the size of the in-memory state cache
- **Transaction Cache**: Size of the transaction cache for rapid retrieval
- **Block Cache**: Number of recent blocks to keep in memory

### Network Configuration

- **Maximum Connections**: Limits the number of peer connections
- **Request Timeout**: Time to wait for RPC requests to complete
- **Socket Buffer Size**: Tuning parameter for network performance

## Integration Interfaces

UFO provides multiple interfaces for integration with external systems:

### ABCI++ Interface

UFO implements the ABCI++ interface for Cosmos SDK compatibility, including:

- `CheckTx`: Validates transactions before adding to mempool
- `PrepareProposal`: Prepares a block proposal
- `ProcessProposal`: Processes and validates a proposed block
- `FinalizeBlock`: Finalizes a block after consensus
- `Commit`: Commits a block to the chain

### CometBFT-Compatible RPC

UFO provides a CometBFT-compatible JSON-RPC interface:

- **HTTP Endpoint**: Accessible via HTTP/JSON-RPC on port 26657
- **WebSocket**: For real-time updates and subscriptions
- **gRPC**: For efficient binary communication

Common RPC methods include:
- `BroadcastTxSync`: Broadcasts a transaction synchronously
- `BroadcastTxAsync`: Broadcasts a transaction asynchronously
- `ABCIQuery`: Queries the application state
- `Block`: Retrieves block information
- `Status`: Gets the current node status

## Monitoring and Logging

### Logging Configuration

UFO supports multiple log levels:

- `debug`: Verbose logging for development and debugging
- `info`: Normal operational logs
- `warn`: Warning conditions
- `error`: Error conditions
- `fatal`: Severe errors that cause termination

Log format can be set to:
- `plain`: Human-readable text
- `json`: Structured JSON format for programmatic parsing

### Metrics and Monitoring

UFO exposes metrics for monitoring:

- **Prometheus Metrics**: Available on port 26660
- **Health Check Endpoint**: Accessible via HTTP GET at `/health`
- **Performance Metrics**: TPS, latency, resource usage statistics

## Deployment Modes

UFO supports different deployment models:

### Fauxmosis Mode

- Simplified mock Cosmos SDK app for testing
- Implements key Cosmos SDK interfaces for compatibility
- Ideal for development and quick testing

Command:
```bash
./fauxmosis-ufo 
```

### Osmosis Bridged Mode

- UFO runs as a separate process, connected to Osmosis
- Provides a bridge interface for compatibility
- Easier to upgrade components independently

Command:
```bash
./osmosis-ufo-bridged 
```

### Osmosis Patched Mode

- Modified Osmosis binary with UFO integration
- Single process for better performance
- Integrated with the Osmosis codebase

Command:
```bash
./osmosisd-ufo 
```

## Benchmarking and Testing

UFO includes a benchmarking suite:

### Single Node Benchmark

```bash
./benchmark_assay/benchmark_node.sh --binary-type fauxmosis-ufo --binary-path /path/to/fauxmosis-ufo --validators 1 --block-times 100,10,1 --tx-count 1000
```

### Comparative Benchmark

```bash
./benchmark_assay/run_quick_comparative_benchmark.sh
```

### Performance Tests

```bash
./benchmark_assay/run_performance_tests.sh --block-times 1000,100,10,1 --tx-count 5000 --duration 180
```

### Diagnostic Tools

- **Node Status**: `curl http://localhost:26657/status`
- **Health Check**: `curl http://localhost:26657/health`
- **Consensus State**: `curl http://localhost:26657/consensus_state`
