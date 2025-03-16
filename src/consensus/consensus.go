package consensus

import (
	"fmt"
	"sync"
)

// ConsensusState represents the current state of the consensus process
type ConsensusState struct {
	Height            int64
	ValidatorSet      *ValidatorSet
	ProposerSelector  ProposerSelector
	PrevoteSet        *VoteSet
	PrecommitSet      *VoteSet
	LockedBlock       *Block
	ProposedBlock     *Block
	CommittedBlocks   []*Block
	CurrentProposer   *Validator
	Mutex             sync.RWMutex
}

// NewConsensusState creates a new consensus state with the given validator set and proposer selector
func NewConsensusState(validatorSet *ValidatorSet, proposerSelector ProposerSelector) *ConsensusState {
	cs := &ConsensusState{
		Height:           1, // Start at height 1
		ValidatorSet:     validatorSet,
		ProposerSelector: proposerSelector,
		CommittedBlocks:  make([]*Block, 0),
	}

	// Initialize vote sets
	cs.PrevoteSet = NewVoteSet(VoteTypePrevote, cs.Height, validatorSet)
	cs.PrecommitSet = NewVoteSet(VoteTypePrecommit, cs.Height, validatorSet)
	
	// Select the initial proposer
	cs.CurrentProposer = proposerSelector.SelectProposer(cs.Height, validatorSet)

	return cs
}

// ProposeBlock proposes a new block for the current height
func (cs *ConsensusState) ProposeBlock(txs []Transaction) (*Block, error) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	// Get the last committed block hash for the previous hash
	prevHash := ""
	if len(cs.CommittedBlocks) > 0 {
		prevHash = cs.CommittedBlocks[len(cs.CommittedBlocks)-1].Hash
	}

	// Create a new block
	block := NewBlock(cs.Height, txs, cs.CurrentProposer, prevHash)
	cs.ProposedBlock = block

	fmt.Printf("Proposed block at height %d by %s: %s\n", 
		cs.Height, cs.CurrentProposer.ID, block.Hash)

	return block, nil
}

// Prevote casts a prevote for the given validator
func (cs *ConsensusState) Prevote(validator *Validator, blockHash string) (*Vote, error) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	// Create a prevote
	vote := NewVote(VoteTypePrevote, cs.Height, validator, blockHash)
	
	// Add vote to prevote set
	if ok := cs.PrevoteSet.AddVote(vote); !ok {
		return nil, fmt.Errorf("failed to add prevote")
	}

	fmt.Printf("Validator %s prevoted for block %s at height %d\n", 
		validator.ID, blockHash, cs.Height)

	return vote, nil
}

// Precommit casts a precommit for the given validator
func (cs *ConsensusState) Precommit(validator *Validator, blockHash string) (*Vote, error) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	// Create a precommit
	vote := NewVote(VoteTypePrecommit, cs.Height, validator, blockHash)
	
	// Add vote to precommit set
	if ok := cs.PrecommitSet.AddVote(vote); !ok {
		return nil, fmt.Errorf("failed to add precommit")
	}

	fmt.Printf("Validator %s precommitted for block %s at height %d\n", 
		validator.ID, blockHash, cs.Height)

	// Check if we have a +2/3 majority
	if hash, ok := cs.PrecommitSet.HasTwoThirdsMajority(); ok {
		fmt.Printf("Block %s has +2/3 precommits at height %d\n", hash, cs.Height)
		
		// If we do, we can commit the block (NOTE: we need to avoid the deadlock)
		// Since we're already holding the mutex, we should commit directly
		// without calling CommitBlock which would try to acquire the mutex again
		if cs.ProposedBlock != nil && cs.ProposedBlock.Hash == hash {
			// Add the block to the committed blocks
			cs.CommittedBlocks = append(cs.CommittedBlocks, cs.ProposedBlock)
			fmt.Printf("Committed block at height %d: %s\n", cs.Height, cs.ProposedBlock.Hash)
			
			// Advance to the next height
			cs.Height++
			
			// Reset vote sets and other state for the new height
			cs.PrevoteSet = NewVoteSet(VoteTypePrevote, cs.Height, cs.ValidatorSet)
			cs.PrecommitSet = NewVoteSet(VoteTypePrecommit, cs.Height, cs.ValidatorSet)
			cs.ProposedBlock = nil
			cs.LockedBlock = nil
			
			// Select the next proposer
			cs.CurrentProposer = cs.ProposerSelector.SelectProposer(cs.Height, cs.ValidatorSet)
			fmt.Printf("New proposer for height %d: %s\n", cs.Height, cs.CurrentProposer.ID)
		}
	}

	return vote, nil
}

// CommitBlock commits a block and advances to the next height
// This function is only used when the block is committed externally,
// not through the Precommit function
func (cs *ConsensusState) CommitBlock(block *Block) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	// Add the block to the committed blocks
	cs.CommittedBlocks = append(cs.CommittedBlocks, block)
	fmt.Printf("Committed block at height %d: %s\n", cs.Height, block.Hash)

	// Advance to the next height
	cs.Height++
	
	// Reset vote sets and other state for the new height
	cs.PrevoteSet = NewVoteSet(VoteTypePrevote, cs.Height, cs.ValidatorSet)
	cs.PrecommitSet = NewVoteSet(VoteTypePrecommit, cs.Height, cs.ValidatorSet)
	cs.ProposedBlock = nil
	cs.LockedBlock = nil
	
	// Select the next proposer
	cs.CurrentProposer = cs.ProposerSelector.SelectProposer(cs.Height, cs.ValidatorSet)
	fmt.Printf("New proposer for height %d: %s\n", cs.Height, cs.CurrentProposer.ID)
}

// GetLatestCommittedBlock returns the latest committed block
func (cs *ConsensusState) GetLatestCommittedBlock() *Block {
	cs.Mutex.RLock()
	defer cs.Mutex.RUnlock()

	if len(cs.CommittedBlocks) == 0 {
		return nil
	}
	return cs.CommittedBlocks[len(cs.CommittedBlocks)-1]
}

// GetCurrentHeight returns the current height
func (cs *ConsensusState) GetCurrentHeight() int64 {
	cs.Mutex.RLock()
	defer cs.Mutex.RUnlock()
	return cs.Height
}

// RunConsensusRound executes a full consensus round with the given validators and proposed block
// This is a simplified version for demonstration purposes
func (cs *ConsensusState) RunConsensusRound(txs []Transaction) (*Block, error) {
	// Propose a block
	block, err := cs.ProposeBlock(txs)
	if err != nil {
		return nil, err
	}

	// Prevote phase
	for _, validator := range cs.ValidatorSet.Validators {
		_, err := cs.Prevote(validator, block.Hash)
		if err != nil {
			return nil, err
		}
	}

	// Check if we have a +2/3 majority on prevotes
	if hash, ok := cs.PrevoteSet.HasTwoThirdsMajority(); !ok || hash != block.Hash {
		return nil, fmt.Errorf("failed to get +2/3 majority on prevotes")
	}

	// Precommit phase
	for _, validator := range cs.ValidatorSet.Validators {
		_, err := cs.Precommit(validator, block.Hash)
		if err != nil {
			return nil, err
		}
	}

	// The block will be committed in the precommit phase if there is a +2/3 majority
	// Return the latest committed block
	return cs.GetLatestCommittedBlock(), nil
} 