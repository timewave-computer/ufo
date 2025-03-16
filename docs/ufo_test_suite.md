# UFO Test Suite

This test suite validates the functionality of UFO, a mock CometBFT implementation designed to provide a low latency alternative to CometBFT for test purposes while maintaining compatibility with the Cosmos ecosystem.

## Structure

The test suite is organized into the following directories:

- `utils/`: Common utility functions used by all tests
- `core/`: Core functionality tests
  - `rest/`: REST API interface tests
  - `grpc/`: gRPC interface tests
  - `websocket/`: WebSocket subscription tests
  - `abci/`: ABCI interface tests
  - `mempool/`: Mempool behavior tests
- `consensus/`: Consensus algorithm tests (coming soon)
- `ibc/`: IBC protocol tests (coming soon)
- `integration/`: Integration tests with external systems (coming soon)
- `stress/`: Stress and performance tests (coming soon)

## Running the Tests

You can run the tests using the following commands:

```bash
# Run all tests
go test ./... -v

# Run a specific test category
go test ./core/rest -v
go test ./core/grpc -v
go test ./core/websocket -v
go test ./core/abci -v
go test ./core/mempool -v

# Run a specific test
go test ./core/rest -run TestTransactionSubmission -v
```

Alternatively, use the Makefile:

```bash
# Run all tests
make test

# Run a specific test category
make test-rest
make test-grpc
make test-websocket
make test-abci
make test-mempool
```

## Configuration

Tests can be configured to run against different binary types:

- `fauxmosis-comet`: The mock CometBFT implementation with Osmosis
- `fauxmosis-ufo`: The UFO implementation with Osmosis
- `osmosis-ufo-bridged`: Osmosis with UFO in bridged mode
- `osmosis-ufo-patched`: Osmosis with UFO in patched mode
- `osmosis-comet`: Standard Osmosis with CometBFT (baseline)
