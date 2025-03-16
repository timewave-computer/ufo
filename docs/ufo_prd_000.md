# ```Product Requirements Document (PRD)```
**Title:** In-Memory BFT Consensus and ABCI++ Mock for Cosmos SDK Testing  
**Date:** March 2025  
**Owner:** [Your Name / Team]  
**Status:** Draft  

## ```1. Overview```  
This document outlines the requirements for building an **in-memory BFT consensus system with a mock ABCI++ application** to facilitate testing Cosmos SDK applications. Developed by Timewave, the system is designed to enhance and accelerate multi-chain test environments by providing ultra-fast consensus with sub-millisecond block times. The system should be lightweight, simulate the essential phases of **CometBFT consensus**, and provide a minimal **ABCI++ interface** for transaction execution and state management. The system supports **pluggable transaction types** to enable simulation of multiple blockchain environments.

## ```2. Goals & Objectives```  
### ```2.1 Primary Goals```  
✅ Simulate a basic **BFT consensus** mechanism (proposal, voting, commit).  
✅ Provide an **in-memory ABCI++ mock** with support for `PrepareProposal`, `ProcessProposal`, `FinalizeBlock`, and `Commit`.  
✅ Enable testing of **Cosmos SDK applications** without requiring a full CometBFT node.  
✅ Support **transaction execution** and maintain a simple in-memory state.  
✅ Implement a **pluggable transaction type system** to support multiple blockchain environments.

### ```2.2 Non-Goals```  
❌ Real networking or cryptographic signatures.  
❌ Byzantine behavior simulation.  
❌ Full-fledged blockchain storage (only in-memory state).  

---

## ```3. System Architecture```  
The system consists of three main components:  
1. **BFT Consensus Simulation** (Validator Set, Proposal Round, Voting, Commit).  
2. **ABCI++ Mock Application** (Processes transactions, maintains state).  
3. **Transaction Processor System** (Pluggable handlers for different transaction types).

### ```3.1 BFT Consensus Simulation```  
- **Validators** (fixed set for now, round-robin proposer selection).  
- **Block Proposal** (simple round-robin selection).  
- **Voting (Prevote & Precommit)** (quorum-based commitment).  
- **Finalization & Commit** (state is committed if 2/3 of validators precommit).  

### ```3.2 ABCI++ Mock Application```  
- **PrepareProposal** (Validators generate proposed blocks).  
- **ProcessProposal** (Validators vote on proposals).  
- **DeliverTx** (Execute transactions in a simple key-value store).  
- **FinalizeBlock** (Mark a block as executed).  
- **Commit** (Store block state & generate a simple state root).  

### ```3.3 Transaction Processor System```  
- **Transaction Processor Interface** (Common interface for all transaction types).
- **Transaction Registry** (Central registry for all transaction processors).
- **Chain-Specific Processors** (Implementations for different blockchain environments).
- **Pluggable Architecture** (Easy addition of new transaction types and blockchains).

---

## ```4. Functional Requirements```  

### ```4.1 BFT Consensus Implementation```  
| Feature | Description | Priority |
|---------|------------|----------|
| **Validator Set** | Fixed list of validators with equal voting power. | High |
| **Block Proposal** | A single validator proposes a block each round. | High |
| **Prevote Phase** | Validators vote on a proposed block. | High |
| **Precommit Phase** | If 2/3+ validators agree, the block is committed. | High |
| **Commit Phase** | The finalized block height is updated. | High |
| **Leader Rotation** | Round-robin selection of the next proposer. | Medium |

---

### ```4.2 ABCI++ Mock Application```  
| Feature | Description | Priority |
|---------|------------|----------|
| **PrepareProposal** | The proposer constructs a block from pending transactions. | High |
| **ProcessProposal** | Validators check and approve the proposed block. | High |
| **DeliverTx** | Execute transaction logic (simple key-value store). | High |
| **FinalizeBlock** | Mark a block as executed. | High |
| **Commit** | Store block state & generate a simple state root. | High |
| **CheckTx** | Validate transactions before including in a block. | Medium |

---

### ```4.3 Transaction Processor System```  
| Feature | Description | Priority |
|---------|------------|----------|
| **Processor Interface** | Common interface for all transaction processors. | High |
| **Transaction Registry** | Central registry to manage multiple processors. | High |
| **KV Store Transactions** | Basic key-value store transaction support. | High |
| **Osmosis Transactions** | Support for Osmosis-specific transactions. | High |
| **Multiple Chain Support** | Ability to simulate multiple blockchain environments. | Medium |
| **Custom Processor Registration** | API for adding new transaction processors. | Medium |

---

## ```5. Open Questions```  
- Should validators be configurable, or is a fixed set sufficient?  
- How should we handle **validator voting power changes** (if at all)?  
- Should we introduce **timeouts for consensus rounds**, or assume ideal network conditions?  
- Should the ABCI++ mock support **state snapshots** and rollback?  
- How realistic should the state root computation be?  
- Should this system be able to **replace CometBFT in a Cosmos SDK node**, or just provide an independent testing environment?  
- How should we handle **cross-chain transactions** if multiple chain types are simulated simultaneously?

---

## ```6. Success Criteria```  
- A **Cosmos SDK module** should be able to run against the ABCI++ mock without requiring CometBFT.  
- The system should be able to **simulate consensus** over multiple blocks.  
- Transactions should be **executed and committed** correctly in the ABCI++ mock.  
- The system should support **multiple transaction types** through the pluggable architecture.
- Adding a **new blockchain environment** should be straightforward using the processor interface.
