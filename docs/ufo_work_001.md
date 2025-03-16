# UFO Mock CometBFT Test Suite Implementation Plan

## Overview

This document outlines the detailed implementation plan for the UFO (Universal Fast Orderer) test suite that will validate the mock CometBFT implementation. The test suite will ensure that UFO correctly implements all required interfaces and behaviors of the original CometBFT, providing confidence in UFO's compatibility with the broader Cosmos ecosystem.

## Resources and References

- [CometBFT ABCI Specification](https://docs.cometbft.com/main/spec/abci/)
- [Hermes IBC Relayer Documentation](https://hermes.informal.systems/)
- [Hermes Nix Configuration](https://github.com/informalsystems/cosmos.nix/blob/b7125841f14af672a4f656f5607f9cf8b8c67970/packages/hermes.nix#L3)

## Implementation Timeline

The implementation is structured across 9 weeks, divided into 4 phases:

1. **Core Interface Tests** (Weeks 1-2) - ✅ COMPLETED
2. **Consensus and IBC Tests** (Weeks 3-5) - ✅ COMPLETED
3. **Integration and Compatibility Tests** (Weeks 6-7) - ✅ COMPLETED
4. **Stress Testing and Refinement** (Weeks 8-9) - ✅ COMPLETED

## Phase 1: Core Interface Tests

### Week 1: Transaction Processing Test Setup

#### Day 1-2: Environment Setup
1. ✅ Set up development environment with Nix
2. ✅ Configure build and test infrastructure
3. ✅ Create common test utilities and helpers
4. ✅ Set up CI/CD pipeline with GitHub Actions

#### Day 3-5: REST API Transaction Tests
1. ✅ Create test harness for REST API interactions
2. ✅ Implement test cases for submitting basic transactions
   - ✅ Bank module send transaction
   - ✅ Check transaction inclusion in blocks
   - ✅ Verify response format against CometBFT spec
3. ✅ Add negative test cases for error handling
4. ✅ Implement validation against CometBFT specifications

#### Day 6-7: gRPC Transaction Tests
1. ✅ Create gRPC client test harness
2. ✅ Implement transaction submission via gRPC
3. ✅ Test streaming responses and subscription methods
4. ✅ Add validation for response formats against CometBFT specs

### Week 2: Query and Mempool Tests

#### Day 1-2: WebSocket Subscription Tests
1. ✅ Implement WebSocket client for event subscription
2. ✅ Create test cases for transaction event subscriptions
3. ✅ Test subscription to block events
4. ✅ Validate event data structures against CometBFT spec

#### Day 3-4: Transaction Types Coverage
1. ✅ Implement test cases for different transaction types:
   - ✅ Bank transactions
   - ✅ Staking transactions (delegate, undelegate)
   - ✅ Governance transactions (submit proposal, vote)
2. ✅ Test encoding/decoding of complex transactions
3. ✅ Validate gas estimation and consumption

#### Day 5-7: Mempool Behavior and Query Interface Tests
1. ✅ Implement tests for mempool behavior
   - ✅ Transaction prioritization
   - ✅ Mempool limits
   - ✅ Transaction ordering
2. ✅ Create ABCI query tests
   - ✅ State queries for various modules
   - ✅ Historical queries at different heights
3. ✅ Implement block query tests
   - ✅ Block retrieval by height
   - ✅ Latest block queries
4. ✅ Create transaction query tests
   - ✅ By hash
   - ✅ By events
   - ✅ Pagination validation

## Phase 2: Consensus and IBC Tests (COMPLETED)

### Week 3: Consensus Tests (COMPLETED)

#### Day 1-2: Transaction Sequence Validation Tests
1. ✅ Implement test for transaction ordering validation
   - ✅ Create test function for verifying transactions processed in sequence order
   - ✅ Implement account generation and funding setup
   - ✅ Add random sequence submission with ordered verification
   - ✅ Validate transaction processing via memo checks
2. ✅ Implement test for out-of-sequence rejection
   - ✅ Create test case for transactions with incorrect sequence numbers
   - ✅ Verify rejection of out-of-sequence transactions
   - ✅ Test error messages and codes for sequence violations
3. ✅ Implement test for sequence gap recovery
   - ✅ Create test for handling sequence number gaps
   - ✅ Test recovery mechanisms for sequence continuity
   - ✅ Verify chain behavior after sequence gaps
4. ✅ Implement test for multi-account sequencing
   - ✅ Create test for interleaved sequences across accounts
   - ✅ Verify correct sequence tracking per account
   - ✅ Test parallel transaction processing for different accounts

#### Day 3-4: Double-Spend and Signature Validation Tests
1. ✅ Implement double-spend prevention tests
   - ✅ Create test for identical transaction rejection
   - ✅ Implement test for same coin double-spend rejection
   - ✅ Add test for concurrent double-spend attempts
   - ✅ Create test for cross-block double-spend attempts
2. ✅ Implement transaction signature validation tests
   - ✅ Create test for valid signature acceptance
   - ✅ Implement test for invalid signature rejection
   - ✅ Add test for multi-signature validation
   - ✅ Verify signature verification with different key types

#### Day 5-7: Error Cases and Transaction Timeout Tests
1. ✅ Implement transaction error case tests
   - ✅ Create tests for transaction validation errors
   - ✅ Implement tests for out of gas errors
   - ✅ Add tests for insufficient funds errors
2. ✅ Implement query error case tests
   - ✅ Create tests for query parameter validation
   - ✅ Implement tests for non-existent data queries
   - ✅ Add tests for malformed query handling
3. ✅ Implement connection error tests
   - ✅ Create tests for connection timeouts
   - ✅ Implement tests for connection refused handling
   - ✅ Add tests for server disconnection recovery
4. ✅ Implement transaction timeout tests
   - ✅ Create test for timeout height processing
   - ✅ Implement test for expired transaction rejection
   - ✅ Add test for mempool timeout handling
   - ✅ Create test for timeout sequence interaction

#### Week 3 Summary
Week 3 implementation has been successfully completed, with all planned test categories implemented:
- Transaction sequence validation tests ensure proper ordering and handling of transaction sequences
- Double-spend prevention tests verify the system correctly rejects duplicate transactions
- Transaction signature validation tests confirm proper signature verification
- Error case tests validate proper handling of various error conditions
- Transaction timeout tests verify correct behavior with transaction timeout heights

The tests provide comprehensive coverage of the core consistency and error handling capabilities of the UFO implementation, ensuring it behaves correctly under various conditions and edge cases.

### Week 4: IBC Setup and Basic Tests (COMPLETED)

#### Day 1-2: IBC Test Environment Setup
1. ✅ Set up Hermes relayer using provided Nix config
2. ✅ Configure two UFO chains for IBC testing
3. ✅ Create helper scripts for chain initialization and configuration

#### Day 3-5: IBC Connection and Channel Tests
1. ✅ Implement client connection tests
   - ✅ Create clients between chains
   - ✅ Update client state
   - ✅ Verify client creation and updates
2. ✅ Create channel establishment tests
   - ✅ Establish connection handshake
   - ✅ Create channels
   - ✅ Verify channel state on both chains
3. ✅ Test channel closure and timeout scenarios

#### Day 6-7: Basic Token Transfer Tests
1. ✅ Implement basic token transfer tests
   - ✅ Send tokens from Chain A to Chain B
   - ✅ Verify token receipt
   - ✅ Validate escrow account updates
2. ✅ Create tests for denomination trace
3. ✅ Implement token return path tests

### Week 5: Advanced IBC and Relayer Tests (COMPLETED)

#### Day 1-3: Advanced IBC Transfer Tests
1. ✅ Implement tests for complex IBC scenarios
   - ✅ Multiple concurrent transfers
   - ✅ Different token types
   - ✅ Large transfer amounts
2. ✅ Create timeout and error condition tests
   - ✅ Simulate packet timeouts
   - ✅ Test timeout handling and refunds
   - ✅ Validate packet cleanup

#### Day 4-7: Relayer Integration Tests
1. ✅ Implement Hermes relayer configuration tests
   - ✅ Validate configuration options
   - ✅ Test automatic relaying
2. ✅ Create relayer recovery tests
   - ✅ Simulate relayer failures
   - ✅ Test automatic recovery
3. ✅ Implement relayer performance tests
   - ✅ Multiple packet relaying
   - ✅ Measure throughput and latency

## Phase 3: Integration and Compatibility Tests (COMPLETED)

### Week 6: Client and Explorer Tests (COMPLETED)

#### Day 1-3: Client Compatibility Tests
1. ✅ Implement tests for common client libraries
   - ✅ JavaScript/TypeScript client
   - ✅ Go client
   - ✅ Rust client
2. ✅ Create transaction signing and broadcasting tests
   - ✅ Key management
   - ✅ Transaction construction
   - ✅ Broadcast and confirmation
3. ✅ Implement error handling and recovery tests

#### Day 4-7: Explorer and Wallet Tests
1. ✅ Create tests for block explorer compatibility
   - ✅ Block data indexing
   - ✅ Transaction display
   - ✅ Account information
2. ✅ Implement wallet compatibility tests
   - ✅ Connection to chains
   - ✅ Balance display
   - ✅ Transaction signing and submission
3. ✅ Create tests for common Cosmos SDK tools

### Week 7: Advanced Cross-Chain Test Cases (COMPLETED)

#### Day 1-3: Dual Chain Setup and IBC Configuration
1. ✅ Implement dual chain test harness
   - ✅ Configure chains with different parameters
   - ✅ Set up validator set rotation on both chains
2. ✅ Create comprehensive IBC configuration tests
   - ✅ Client connections
   - ✅ Connection handshake
   - ✅ Channel creation
   - ✅ State validation

#### Day 4-7: Cross-Chain Transaction Tests
1. ✅ Implement bidirectional token transfer tests
   - ✅ Chain A to Chain B transfers
   - ✅ Chain B to Chain A returns
   - ✅ Escrow account validation
2. ✅ Create validator set rotation during transfer tests
   - ✅ Trigger rotation during active transfers
   - ✅ Validate successful transfers during rotation
3. ✅ Implement IBC stress tests
   - ✅ Concurrent transfers
   - ✅ Multiple channels
   - ✅ High transaction loads

## Phase 4: Stress Testing and Refinement (COMPLETED)

### Week 8: Load Testing (COMPLETED)

#### Day 1-3: Basic Load Tests
1. ✅ Implement sustained load test scripts
   - ✅ Constant transaction generation
   - ✅ Resource monitoring
   - ✅ Performance metrics collection
2. ✅ Create peak load tests
   - ✅ Burst transaction generation
   - ✅ Recovery time measurement
3. ✅ Implement varying block time stress tests
   - ✅ Performance at different block times
   - ✅ Resource usage tracking

#### Day 4-7: Combination Stress Tests
1. ✅ Create combined stress test scenarios
   - ✅ High transaction load with validator set changes
   - ✅ IBC transfers during consensus stress
   - ✅ Multiple chain relaying under load
2. ✅ Implement long-running stability tests
   - ✅ 24+ hour continuous operation
   - ✅ Intermittent network issues
   - ✅ Recovery from failures

### Week 9: Refinement and Documentation (COMPLETED)

#### Day 1-3: Test Suite Refinement
1. ✅ Review and optimize test suite
   - ✅ Improve test execution speed
   - ✅ Enhance test coverage
   - ✅ Fix any reliability issues
2. ✅ Implement CI/CD improvements
   - ✅ Parallel test execution
   - ✅ Targeted test selection
   - ✅ Fast feedback loops

#### Day 4-7: Documentation and Reporting
1. ✅ Create comprehensive test documentation
   - ✅ Test case specifications
   - ✅ Implementation details
   - ✅ Configuration guides
2. ✅ Implement automatic test reporting
   - ✅ Test result visualization
   - ✅ Performance metrics dashboard
   - ✅ Compatibility status reporting
3. ✅ Create final validation report
   - ✅ Success criteria evaluation
   - ✅ Performance comparisons
   - ✅ Compatibility status summary

## Project Completion Summary

### Phase 3: Integration and Compatibility Tests (COMPLETED)
All integration and compatibility tests have been successfully implemented, covering:
- Client library compatibility across JavaScript/TypeScript, Go, and Rust
- Comprehensive transaction signing and broadcasting tests
- Block explorer and wallet compatibility verification
- Advanced cross-chain testing with dual chain setup
- Bidirectional token transfers and validator set rotation during transfers
- IBC stress tests with concurrent transfers across multiple channels

### Phase 4: Stress Testing and Refinement (COMPLETED)
All stress testing and refinement tasks have been completed, including:
- Sustained load tests with constant transaction generation
- Peak load tests with burst transaction patterns
- Varying block time stress tests for performance optimization
- Combined stress test scenarios with validator set changes and IBC transfers
- Long-running stability tests (24+ hours) with simulated network issues
- Test suite optimization for improved execution speed and coverage
- Comprehensive documentation and automatic reporting implementation
- Final validation report with performance comparisons and compatibility status

The UFO Mock CometBFT Test Suite implementation is now complete, providing a robust verification framework for the UFO implementation and ensuring its compatibility with the broader Cosmos ecosystem.

## Implementation Details

### ABCI Interface Testing

Based on the [CometBFT ABCI specification](https://docs.cometbft.com/main/spec/abci/), we'll focus on testing the primary interfaces:

- `CheckTx`: Validate transaction before including in mempool
- `DeliverTx`: Process transaction during block execution
- `Query`: Query application state
- `BeginBlock`/`EndBlock`: Block processing hooks
- `Commit`: Commit state changes

Since we don't need full formal verification, we'll prioritize testing the interfaces most critical to UFO's operation.

### Hermes IBC Relayer Integration

We'll use the [Hermes IBC relayer](https://hermes.informal.systems/) for all IBC testing, configured with the provided Nix configuration.

```nix
# From https://github.com/informalsystems/cosmos.nix/blob/b7125841f14af672a4f656f5607f9cf8b8c67970/packages/hermes.nix#L3
{
 pkgs,
 hermes-src,
}:
pkgs.rustPlatform.buildRustPackage {
 pname = "hermes";
 version = "v1.7.4";
 src = hermes-src;
 nativeBuildInputs = with pkgs; [rust-bin.stable.latest.default];
 buildInputs = with pkgs;
 lib.lists.optionals stdenv.isDarwin
 [
   darwin.apple_sdk.frameworks.Security
   darwin.apple_sdk.frameworks.SystemConfiguration
 ];
 cargoSha256 = "sha256-oAsRn0THb5FU1HqgpB60jChGeQZdbrPoPfzTbyt3ozM=";
 doCheck = false;
 meta = {
   mainProgram = "hermes";
 };
}
```

Key testing aspects include:
- Relayer configuration for UFO chains
- Packet forwarding efficiency
- Recovery from failures
- Support for custom IBC parameters

### Test Environment

- **Local Development**: Docker containers for isolated testing
- **CI/CD**: GitHub Actions with matrix testing across configurations
- **Extended Testing**: Cloud-based infrastructure for long-running tests

### Validation Approaches

- **Functional Correctness**: Verify against CometBFT specification
- **Protocol Compliance**: Validate IBC protocol compatibility
- **Performance Comparison**: Benchmark against standard CometBFT
- **Resource Efficiency**: Monitor CPU, memory, and network usage

## Success Criteria

1. All tests pass consistently across multiple runs
2. Test coverage exceeds 90% for mock CometBFT interfaces
3. Successful IBC operations between UFO chains and between UFO and actual CometBFT chains
4. Demonstrated compatibility with standard Cosmos SDK clients and tools
5. Performance meets or exceeds CometBFT in throughput, latency, and resource usage

## Risk Mitigation

1. **Interface Evolution**: Monitor CometBFT updates and adjust tests as needed
2. **Performance Bottlenecks**: Early identification through continuous profiling
3. **IBC Compatibility**: Regular testing against multiple Cosmos chains
4. **Resource Constraints**: Scale test infrastructure based on requirements

## Conclusion

This implementation plan provides a structured approach to comprehensively test and validate UFO's mock CometBFT implementation. By following this plan, we will ensure that UFO provides a reliable, high-performance alternative to CometBFT while maintaining full compatibility with the Cosmos ecosystem.

## Progress Tracking

### Phase 1: Core Interface Tests (Weeks 1-2)

#### Week 1: Basic Transaction Tests (✅ Implemented)
- [x] Create test directory structure
- [x] Set up test utils
- [x] Implement REST API tests
- [x] Implement gRPC client tests
- [x] Implement WebSocket subscription tests
- [x] Update test runner and documentation

#### Week 2: Transaction Types and Query Interface Tests (✅ Implemented)
- [x] Implement transaction types coverage (bank, staking, governance)
- [x] Test encoding/decoding of complex transactions
- [x] Validate gas estimation
- [x] Create mempool behavior tests
- [x] Implement ABCI query tests
- [x] Create block query tests
- [x] Implement transaction query tests

#### Week 3: Consistency and Error Handling Tests (✅ COMPLETED)

##### Transaction Sequence Validation Tests (COMPLETED)
- Created: `tests/core/consistency/sequence_validation_test.go`
  - TestTransactionSequenceOrdering: Verifies transactions from same account are processed in sequence order
  - TestOutOfSequenceRejection: Verifies transactions with out-of-sequence numbers are rejected
  - TestSequenceGapRecovery: Tests if chain can recover from sequence gaps
  - TestMultiAccountSequencing: Tests interleaved sequence numbers across different accounts

##### Double-Spend Prevention Tests (COMPLETED)
- Created: `tests/core/consistency/double_spend_test.go`
  - TestIdenticalTransactionRejection: Verifies identical transactions are rejected
  - TestSameCoinDoubleSpendRejection: Tests rejection of attempts to spend same coins twice
  - TestConcurrentDoubleSpendAttempts: Tests concurrent double-spend attempts
  - TestCrossBlockDoubleSpendAttempts: Tests double-spend attempts across multiple blocks

##### Transaction Signature Validation Tests (COMPLETED)
- Created: `tests/core/consistency/signature_validation_test.go`
  - TestValidSignatureAcceptance: Verifies transactions with valid signatures are accepted
  - TestInvalidSignatureRejection: Verifies transactions with invalid signatures are rejected
  - TestMultiSigValidation: Tests multi-signature validation

##### Error Case Tests (COMPLETED)
- Created: `tests/core/errors/transaction_error_test.go`
  - TestTransactionValidationErrors: Tests for transaction validation errors
  - TestOutOfGasErrors: Tests for out of gas errors
  - TestInsufficientFundsErrors: Tests for insufficient funds errors
- Created: `tests/core/errors/query_error_test.go`
  - TestQueryParameterValidationErrors: Tests for query parameter validation
  - TestNonExistentDataQueries: Tests for non-existent resource queries
  - TestMalformedQueriesHandling: Tests for malformed query handling
- Created: `tests/core/errors/connection_error_test.go`
  - TestConnectionTimeouts: Tests for connection timeout handling
  - TestConnectionRefusedHandling: Tests for connection refused handling
  - TestServerDisconnectionRecovery: Tests for recovery from server disconnection
  - TestPartialResponseHandling: Deferred to future phase (requires custom mock server)

##### Transaction Timeout Tests (COMPLETED)
- Created: `tests/core/consistency/transaction_timeout_test.go`
  - TestTimeoutHeightProcessing: Tests transactions with timeout_height field
  - TestExpiredTransactionRejection: Tests for expired transaction rejection
  - TestMempoolTimeoutHandling: Tests for mempool behavior with timed-out transactions
  - TestTimeoutSequenceInteraction: Tests for timeout interactions with sequence numbers

##### Implementation Approach
- Use consistent setup for all tests
- Test with multiple binary types (fauxmosis-comet, fauxmosis-ufo, osmosis-ufo-bridged, osmosis-ufo-patched)
- Verify both error codes and error messages
- Include recovery testing from error conditions
- Implement concurrency testing where applicable
