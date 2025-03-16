// Package memstate provides in-memory state management
package memstate

import "crypto/sha256"

// State represents an in-memory state
type State struct {
	Data map[string][]byte
}

// NewState creates a new in-memory state
func NewState() *State {
	return &State{
		Data: make(map[string][]byte),
	}
}

// Set stores a value in the state
func (s *State) Set(key string, value []byte) {
	s.Data[key] = value
}

// Get retrieves a value from the state
func (s *State) Get(key string) []byte {
	return s.Data[key]
}

// Commit saves the current state and returns a hash
func (s *State) Commit() []byte {
	// Simple hash of all keys and values
	hasher := sha256.New()
	for k, v := range s.Data {
		hasher.Write([]byte(k))
		hasher.Write(v)
	}
	return hasher.Sum(nil)
}
