package consensus

// ProposerSelector implements different strategies for selecting a proposer
type ProposerSelector interface {
	// SelectProposer selects a proposer for the given height
	SelectProposer(height int64, validatorSet *ValidatorSet) *Validator
}

// RoundRobinProposerSelector implements a simple round-robin proposer selection
type RoundRobinProposerSelector struct{}

// NewRoundRobinProposerSelector creates a new round-robin proposer selector
func NewRoundRobinProposerSelector() *RoundRobinProposerSelector {
	return &RoundRobinProposerSelector{}
}

// SelectProposer selects a proposer for the given height using a round-robin strategy
// The proposer is selected based on the height modulo the number of validators
func (rr *RoundRobinProposerSelector) SelectProposer(height int64, validatorSet *ValidatorSet) *Validator {
	if validatorSet == nil || validatorSet.Size() == 0 {
		return nil
	}

	// Simple round-robin selection based on height
	index := int(height) % validatorSet.Size()
	return validatorSet.GetValidatorByIndex(index)
} 