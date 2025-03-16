# UFO PRD-001: Mock CometBFT Validation Test Suite

## Document Control
- **Document Status:** Draft
- **Version:** 0.1
- **Last Updated:** 2025-03-15
- **Owner:** Sam Hart

## 1. Introduction

### 1.1 Purpose
This Product Requirements Document (PRD) outlines a testing strategy for validating that the mock CometBFT implementation within UFO correctly implements the required interfaces and behaviors of the original CometBFT. Successful implementation of these tests will provide confidence in UFO's compatibility.

### 1.2 Background
UFO provides a lightweight consensus alternative to CometBFT for Cosmos applications. To ensure seamless integration, UFO implements mock interfaces that mimic CometBFT's behavior. Thorough validation of these interfaces is critical to ensure applications built for CometBFT function correctly when using UFO.

### 1.3 Scope
This PRD covers the test suite design and implementation to validate the mock CometBFT interfaces. It includes transaction processing, query interfaces, consensus mechanisms, IBC protocol compatibility, and infrastructure compatibility.

## 2. Goals and Objectives

### 2.1 Primary Goals
- Develop a robust test suite for validating mock CometBFT implementations
- Create automated tests that can be integrated into CI/CD pipelines
- Provide confidence that UFO-based chains can interoperate properly with other Cosmos chains
- Validate that UFO's unique features (e.g., validator set rotation) function properly in conjunction with CometBFT interfaces

### 2.2 Success Criteria
- All tests pass consistently across multiple runs
- Test coverage exceeds 90% for mock CometBFT interfaces
- Successful IBC operations between UFO chains and between UFO and actual CometBFT chains
- Demonstrated compatibility with standard Cosmos SDK clients and tools

## 3. User Stories and Requirements

### 3.1 User Stories
- As a blockchain developer, I want to use UFO as a drop-in replacement for CometBFT, so that my application can benefit from UFO's performance without code modifications.
- As a developer, I want a testing harness that gives me confidence that my UFO integration works correctly.

### 3.2 Functional Requirements
1. Mock CometBFT must implement all public interfaces of CometBFT required for Cosmos SDK compatibility
2. Transaction submission must work via all standard interfaces (REST, gRPC, WebSockets)
3. Query functionality must return valid responses in the expected format
4. Consensus must proceed according to the UFO rules while maintaining expected CometBFT output format
5. IBC protocol handlers must allow for cross-chain communication
6. UFO-specific features must not interfere with CometBFT interface compatibility

## 4. Testing Requirements

### 4.1 Core Transaction Processing Tests

#### 4.1.1 REST API Transaction Test
- Submit transactions via the REST API endpoint
- Verify transactions are included in blocks
- Validate response format matches CometBFT specification
- Test error conditions and response codes

#### 4.1.2 gRPC Transaction Test
- Submit transactions using the gRPC interface
- Verify confirmation of transaction inclusion
- Test streaming responses for transaction submission
- Validate error handling and error codes

#### 4.1.3 WebSocket Subscription Test
- Subscribe to transaction events via WebSocket
- Verify events fire when transactions are processed
- Test subscription to specific event types
- Validate event data structure matches CometBFT specification

#### 4.1.4 Transaction Types Coverage
- Test all supported transaction types (bank send, staking, governance, etc.)
- Verify proper encoding/decoding of transaction data
- Test complex transactions with multiple messages
- Validate gas estimation and consumption

#### 4.1.5 Mempool Behavior Test
- Verify transactions are properly queued and prioritized
- Test mempool limits and behavior under load
- Validate transaction eviction policies
- Ensure transactions are included in blocks in the expected order

### 4.2 Query Interface Tests

#### 4.2.1 ABCI Query Test
- Verify queries to application state return correct data
- Test path-based queries to various modules
- Validate query proof generation
- Test historical queries at different heights

#### 4.2.2 Block Query Test
- Test retrieving blocks by height returns valid block data
- Validate block structure and metadata
- Test querying for the latest block
- Test block result queries

#### 4.2.3 Transaction Query Test
- Test retrieving transactions by hash
- Validate transaction data and inclusion proofs
- Test querying transactions by event
- Validate search functionality and pagination

#### 4.2.4 Validator Set Query
- Verify queries for current validator set return correct data
- Test historical validator set queries
- Validate validator updates are correctly reported
- Test consensus state queries

#### 4.2.5 Event Query Test
- Test subscription to events
- Validate event filtering functionality
- Test complex event queries
- Verify event data structure

### 4.3 Consensus Tests

#### 4.3.1 Block Production Test
- Verify blocks are produced at expected intervals
- Validate block structure and signatures
- Test under various load conditions
- Measure block time stability

#### 4.3.2 Validator Rotation Test
- Start with validator set A
- Trigger rotation to validator set B
- Verify consensus continues without interruption
- Test frequent validator set changes

#### 4.3.3 Proposer Selection Test
- Track block proposers over time
- Verify proper round-robin selection
- Test weighted selection based on voting power
- Validate proposer schedule calculation

#### 4.3.4 Byzantine Fault Tolerance Test
- Deliberately introduce faults in minority of validators
- Verify chain continues to produce blocks
- Test various fault scenarios (non-responsiveness, incorrect voting, etc.)
- Measure consensus recovery time

#### 4.3.5 Block Finality Test
- Verify transactions achieve finality
- Test confirmation times
- Validate fork choice rules
- Measure finality latency

### 4.4 IBC Protocol Tests

#### 4.4.1 Chain Connection Test
- Connect two Fauxmosis-UFO chains via IBC
- Verify client creation on both chains
- Validate light client updates
- Test misbehavior detection

#### 4.4.2 Channel Establishment Test
- Create IBC channels between chains
- Test channel handshake protocol
- Verify channel state on both chains
- Test channel closure

#### 4.4.3 Token Transfer Test
- Transfer tokens from Chain A to Chain B
- Verify balances update correctly on both chains
- Test denomination trace
- Validate escrow mechanics

#### 4.4.4 Packet Timeout Test
- Create scenarios where IBC packets time out
- Verify timeout handling
- Test packet cleanup
- Validate refund mechanics

#### 4.4.5 Relayer Behavior Test
- Verify relayer correctly forwards packets between chains
- Test relayer recovery from failures
- Validate path selection
- Measure relayer efficiency

### 4.5 Infrastructure Compatibility Tests

#### 4.5.1 Client Compatibility Test
- Verify standard Cosmos SDK clients can connect to UFO
- Test client libraries (JavaScript, Go, Rust)
- Validate transaction signing and broadcasting
- Test error handling and recovery

#### 4.5.2 Explorer Compatibility
- Test that block explorers can properly index UFO chains
- Verify transaction display
- Validate validator information
- Test event and log display

#### 4.5.3 Wallet Compatibility
- Test common wallets can connect to UFO chains
- Verify transaction signing
- Validate balance display
- Test staking operations

### 4.6 Advanced IBC Cross-Chain Test Case

#### 4.6.1 Dual Chain Setup
- Launch two instances of Fauxmosis-UFO with distinct chain IDs
- Configure both with UFO's validator set rotation functionality
- Verify independent operation of each chain

#### 4.6.2 IBC Configuration
- Create client connections between both chains
- Establish connection handshake
- Create transfer channel between chains
- Validate connection state

#### 4.6.3 Cross-Chain Transactions
- Send tokens from Chain A to Chain B
- Verify tokens are received on Chain B
- Verify escrow account on Chain A is updated correctly
- Test token return path

#### 4.6.4 Validator Set Rotation During IBC Transfers
- Initiate validator set rotation on Chain A
- During rotation, send token transfers to Chain B
- Verify transfers complete successfully despite validator changes
- Validate client state updates with new validator set

#### 4.6.5 Bidirectional Load Testing
- Generate constant transaction load on both chains
- Perform bidirectional token transfers
- Trigger validator set rotations on both chains at different intervals
- Verify system maintains consistency and liveness

## 5. Success Metrics

### 5.1 Test Coverage
- 100% coverage of CometBFT public interfaces
- All transaction types tested
- All query paths tested
- All IBC interfaces covered

### 5.2 Performance Metrics
- Block production rate within 5% of target
- Transaction confirmation time comparable to or better than CometBFT
- Resource usage (CPU, memory) lower than standard CometBFT
- Short blocktimes see minimal increase in transaction drop rate
- Confirm low latency consensus works with IBC relayers

### 5.3 Reliability Metrics
- Test suite passes 100 consecutive runs
- No consensus failures during extended test periods
- No data inconsistencies during validator set rotations
- No IBC packet timeouts under normal conditions

## 6. Open Questions

### 6.1 Technical Questions
1. **IBC Compatibility Edge Cases**: What edge cases in IBC might expose differences between UFO and real CometBFT implementations?
2. **Client Versioning**: How should mock CometBFT handle version information to maximize compatibility with tools expecting specific CometBFT versions?
3. **Consensus Parameters**: Which CometBFT consensus parameters should be exposed for configuration in UFO, and which should be hardcoded?
4. **Light Client Implementation**: How complete does our light client implementation need to be for proper IBC operation?

### 6.2 Testing Questions
1. **Test Environment**: What is the optimal environment configuration for running these tests (local, cloud, network conditions)?
2. **Determinism**: How can we ensure test determinism when dealing with timing-sensitive consensus operations?
3. **Test Isolation**: How do we properly isolate tests that might affect each other when operating on the same chain?
4. **Mocked Components**: Which components should be fully implemented vs. mocked for testing efficiency?
5. **Stress Testing Thresholds**: What load levels should we test to validate performance under stress?

## 7. Timeline and Milestones

### 7.1 Phase 1: Core Interface Tests (2 weeks)
- Implement transaction processing tests
- Implement query interface tests
- Validate basic consensus operation

### 7.2 Phase 2: Consensus and IBC Tests (3 weeks)
- Implement detailed consensus tests
- Develop IBC protocol tests
- Create validator rotation tests

### 7.3 Phase 3: Integration and Compatibility Tests (2 weeks)
- Implement infrastructure compatibility tests
- Integrate with explorers and wallets
- Develop advanced cross-chain test cases

### 7.4 Phase 4: Stress Testing and Refinement (2 weeks)
- Conduct load testing
- Refine test suite based on findings
- Document test results and compatibility status

## 8. Appendix

### 8.1 Related Documents
- UFO Technical Specification
- CometBFT Interface Documentation
- IBC Protocol Specification
- Cosmos SDK Integration Guide

### 8.2 Glossary
- **UFO**: Universal Fast Orderer, a lightweight consensus alternative to CometBFT
- **CometBFT**: The consensus engine used by default in Cosmos SDK applications
- **IBC**: Inter-Blockchain Communication Protocol
- **Fauxmosis**: Mock implementation of Osmosis using UFO
- **ABCI**: Application Blockchain Interface
- **Validator Set Rotation**: UFO's mechanism for changing validator sets

### 8.3 Test Architecture Diagram
(To be added: Diagram showing test components and their interactions)

### 8.4 Resource Requirements
- Development resources: 2-3 engineers
- Test infrastructure: Cloud-based testing environment
- CI/CD integration: GitHub Actions or similar
- External dependencies: IBC relayer, block explorer, wallet software for compatibility testing 