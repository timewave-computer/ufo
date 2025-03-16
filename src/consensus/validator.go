package consensus

import (
	"fmt"
)

// Validator represents a consensus participant with voting power
type Validator struct {
	ID          string
	Address     string
	VotingPower int64
}

// NewValidator creates a new validator with the given ID, address, and voting power
func NewValidator(id, address string, votingPower int64) *Validator {
	return &Validator{
		ID:          id,
		Address:     address,
		VotingPower: votingPower,
	}
}

// String returns a string representation of the validator
func (v *Validator) String() string {
	return fmt.Sprintf("Validator{ID: %s, Address: %s, VotingPower: %d}", v.ID, v.Address, v.VotingPower)
}

// ValidatorSet represents a set of validators participating in consensus
type ValidatorSet struct {
	Validators []*Validator
	TotalPower int64
}

// NewValidatorSet creates a new validator set with the given validators
func NewValidatorSet(validators []*Validator) *ValidatorSet {
	var totalPower int64
	for _, v := range validators {
		totalPower += v.VotingPower
	}

	return &ValidatorSet{
		Validators: validators,
		TotalPower: totalPower,
	}
}

// HasTwoThirdsMajority checks if the given voting power represents a +2/3 majority
func (vs *ValidatorSet) HasTwoThirdsMajority(votingPower int64) bool {
	return votingPower > vs.TotalPower*2/3
}

// HasOneTenthMajority checks if the given voting power represents a +1/10 majority (used for prevotes)
func (vs *ValidatorSet) HasOneTenthMajority(votingPower int64) bool {
	return votingPower > vs.TotalPower/10
}

// Size returns the number of validators in the set
func (vs *ValidatorSet) Size() int {
	return len(vs.Validators)
}

// GetValidatorByIndex returns the validator at the given index
func (vs *ValidatorSet) GetValidatorByIndex(index int) *Validator {
	if index < 0 || index >= len(vs.Validators) {
		return nil
	}
	return vs.Validators[index]
}

// GetValidatorByID returns the validator with the given ID
func (vs *ValidatorSet) GetValidatorByID(id string) *Validator {
	for _, v := range vs.Validators {
		if v.ID == id {
			return v
		}
	}
	return nil
} 