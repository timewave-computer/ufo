package consensus

import (
	"time"
)

// VoteType represents the type of a vote
type VoteType int

const (
	// VoteTypePrevote represents a prevote in the consensus process
	VoteTypePrevote VoteType = iota
	// VoteTypePrecommit represents a precommit in the consensus process
	VoteTypePrecommit
)

// Vote represents a vote in the consensus process
type Vote struct {
	Type        VoteType
	Height      int64
	Validator   *Validator
	BlockHash   string // Hash of the block being voted on, empty if nil vote
	Timestamp   time.Time
}

// NewVote creates a new vote with the given parameters
func NewVote(voteType VoteType, height int64, validator *Validator, blockHash string) *Vote {
	return &Vote{
		Type:        voteType,
		Height:      height,
		Validator:   validator,
		BlockHash:   blockHash,
		Timestamp:   time.Now(),
	}
}

// VoteSet collects votes for a specific height
type VoteSet struct {
	Type           VoteType
	Height         int64
	Votes          map[string]*Vote // Map of validator ID to vote
	VotingPower    map[string]int64 // Map of block hash to total voting power
	ValidatorSet   *ValidatorSet
}

// NewVoteSet creates a new vote set for the given height and vote type
func NewVoteSet(voteType VoteType, height int64, validatorSet *ValidatorSet) *VoteSet {
	return &VoteSet{
		Type:           voteType,
		Height:         height,
		Votes:          make(map[string]*Vote),
		VotingPower:    make(map[string]int64),
		ValidatorSet:   validatorSet,
	}
}

// AddVote adds a vote to the vote set
func (vs *VoteSet) AddVote(vote *Vote) bool {
	// Check if vote is valid
	if vote.Height != vs.Height || vote.Type != vs.Type {
		return false
	}

	// Check if validator has already voted
	_, ok := vs.Votes[vote.Validator.ID]
	if ok {
		return false // Validator has already voted
	}

	// Add vote
	vs.Votes[vote.Validator.ID] = vote

	// Update voting power for the block hash
	if vote.BlockHash != "" {
		vs.VotingPower[vote.BlockHash] += vote.Validator.VotingPower
	}

	return true
}

// HasTwoThirdsMajority checks if any block has +2/3 of the voting power
func (vs *VoteSet) HasTwoThirdsMajority() (string, bool) {
	for blockHash, power := range vs.VotingPower {
		if vs.ValidatorSet.HasTwoThirdsMajority(power) {
			return blockHash, true
		}
	}
	return "", false
}

// HasOneTenthMajority checks if any block has +1/10 of the voting power
func (vs *VoteSet) HasOneTenthMajority() (string, bool) {
	for blockHash, power := range vs.VotingPower {
		if vs.ValidatorSet.HasOneTenthMajority(power) {
			return blockHash, true
		}
	}
	return "", false
}

// GetVotingPower returns the total voting power for a specific block hash
func (vs *VoteSet) GetVotingPower(blockHash string) int64 {
	return vs.VotingPower[blockHash]
} 