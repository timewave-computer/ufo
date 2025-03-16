# UFO Integration with Cosmos SDK Applications

This document describes the integration approaches between UFO (Universal Fast Orderer) and Cosmos SDK applications, including various integration modes and benchmarking capabilities.

## Overview

UFO serves as a drop-in replacement for CometBFT in Cosmos SDK applications. It provides a lightweight consensus system that can significantly improve performance and reduce resource usage. 

UFO was specifically created by Timewave to enhance and accelerate multi-chain test environments. The standard CometBFT consensus introduces significant overhead and slow block times, which can extend development cycles when testing across multiple chains. UFO addresses this by providing ultra-fast consensus with sub-millisecond block times while maintaining compatibility with Cosmos SDK applications.

This document focuses on:

1. Different integration approaches (Bridged and Patched)
2. Mock application integration (Fauxmosis with CometBFT and Fauxmosis with UFO)
3. Benchmarking and performance comparison
4. Implementation details and architecture

## Integration Approaches

UFO provides multiple integration methods to accommodate different use cases:

### 1. Mock Integration (Fauxmosis)

Fauxmosis is a lightweight mock Cosmos SDK application designed for testing and benchmarking. This integration:

- Provides a clean, simplified testing environment
- Implements a mock ABCI application
- Focuses on performance benchmarking
- Simulates the validator set and consensus
- Implements a CometBFT-compatible RPC server

Fauxmosis can run with either:
- CometBFT consensus (fauxmosis-comet)
- UFO consensus (fauxmosis-ufo)

### 2. Bridged Integration

In bridged mode, UFO connects to a running Osmosis instance through adapter interfaces:

- UFO and Osmosis run as separate processes
- Communication occurs through defined interfaces
- Minimal changes to Osmosis codebase
- Easier to update components independently
- Suitable for development and testing

### 3. Patched Integration

Patch-based integration directly modifies the Osmosis source code:

- Applies patches to replace CometBFT with UFO at the source level
- Creates a single integrated binary
- More efficient performance (no inter-process communication)
- Deeper integration at the code level
- Suitable for production deployments

## Architecture

The integration architecture uses adapters to translate between UFO and CometBFT interfaces:

```
+---------------+       +----------------+       +--------------+
| Cosmos App    | <---> | CometBFT       | <---> | Consensus    |
| (Osmosis)     |       | Adapter        |       | Engine (UFO) |
+---------------+       +----------------+       +--------------+
```

This adapter-based approach allows UFO to:

1. Implement the CometBFT interface required by Cosmos SDK apps
2. Redirect consensus operations to the UFO engine
3. Maintain compatibility with existing Cosmos SDK applications
4. Provide significant performance improvements

## Fauxmosis Integration

The Fauxmosis-UFO integration is implemented in the `src/fauxmosis-integration` package.

### Key Components

The Fauxmosis integration consists of:

1. **FauxmosisUFOIntegration**: Main integration class that coordinates all components
2. **CometBFTAdapter**: Translates between the UFO consensus engine and CometBFT interfaces
3. **RPCHTTPServer**: HTTP server for simulating the CometBFT RPC API
4. **RPCClient**: Mock implementation of the CometBFT RPC client

### Using Fauxmosis-UFO

To run the Fauxmosis-UFO integration:

```bash
go run ./cmd/fauxmosis-ufo/main.go
```

Or build and run:

```bash
go build -o fauxmosis-ufo ./cmd/fauxmosis-ufo/main.go
./fauxmosis-ufo
```

### Configuration

Fauxmosis with UFO can be configured for different validator counts and block times:

```bash
./benchmark_assay/benchmark_node.sh --binary-type fauxmosis-ufo --validators 4 --block-times 1000,100,10
```

## Osmosis Bridged Integration

The bridged integration approach consists of the following components:

1. **Bridge Client**: Connects UFO to a running Osmosis instance
2. **Adapter Layer**: Translates between UFO and CometBFT interfaces
3. **Communication Protocol**: Defines how the two systems interact
4. **Independent Processes**: Allows UFO and Osmosis to run separately

To build the bridged mode version:

```bash
go build -o osmosis-ufo-bridged ./cmd/osmosis-ufo-bridged/main.go
```

## Osmosis Patched Integration

The patch-based integration approach consists of the following components:

1. **Adapter Interface**: The `OsmosisUFOAdapter` type provides a bridge between Osmosis and UFO
2. **Source Patches**: Direct modifications to the Osmosis codebase
3. **Dependency Replacement**: Swaps CometBFT dependencies with UFO implementations
4. **Single Binary**: Produces an integrated application with UFO consensus

You can build Osmosis with UFO patch-based integration using the Nix-integrated build script:

```bash
nix run .#build-osmosis -- /path/to/osmosis/source
```

This will create a binary called `osmosisd-ufo` in the Osmosis directory that uses UFO for consensus instead of CometBFT.

## Benchmarking

UFO includes comprehensive benchmark tools for all integration modes. The benchmarks compare:

1. Fauxmosis with CometBFT
2. Fauxmosis with UFO
3. Osmosis with UFO (Bridged)
4. Osmosis with UFO (Patched)
5. Osmosis with CometBFT (baseline)

### Running Benchmarks

To run benchmarks:

```bash
# Quick comparative benchmark
./benchmark_assay/run_quick_comparative_benchmark.sh

# Custom benchmark with specific configurations
./benchmark_assay/run_performance_tests.sh --block-times 1000,100,10 --tx-count 5000 --duration 180

# Specific configurations
./benchmark_assay/run_performance_tests.sh --configurations fauxmosis-comet-1,osmosis-ufo-bridged-1,osmosis-ufo-patched-1,osmosis-comet-1
```

### Benchmark Metrics

The benchmarks measure:

- Transactions per second (TPS)
- Latency
- CPU usage
- Memory usage
- Block production rate
- Scalability across validator counts

## Implementation Details

### Validator Management

UFO implements validators compatible with CometBFT's model:

```go
type Validator struct {
    ID           string
    Address      string
    VotingPower  int64
}
```

### Consensus Protocol

UFO implements a simplified consensus protocol that provides:

1. Fast block production (down to sub-millisecond)
2. Validator set rotation
3. Transaction ordering and finality
4. In-memory state synchronization

### Note on Build Determinism

The UFO project uses Nix for its primary build system, which enables reproducible builds with precise dependency management. However, when using `go mod tidy` or other Go tools directly, you may encounter errors related to incompatible dependencies:

```
go: github.com/cosmos/cosmos-sdk@v0.50.4: reading github.com/osmosis-labs/cosmos-sdk/go.mod at revision v0.47.5-osmo-v24: unknown revision v0.47.5-osmo-v24
```

This is expected behavior because:

1. The project uses specific forks of Cosmos SDK libraries that are managed through Nix
2. The `go.mod` file includes replacements that point to local directories (e.g., `./osmosis-fork/osmosis`)
3. The local replacement directories are populated during the Nix build process

**Solution:**
- Use Nix for building and running the project (`nix build`, `nix develop`, etc.)
- If needed, an empty `go.sum` file can be created to satisfy IDE linting tools
- The true dependency resolution happens through Nix, not Go modules 