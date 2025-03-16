// Package osmosis provides Osmosis chain integration
package osmosis

// Chain represents an Osmosis chain
type Chain struct {
	Name string
}

// NewChain creates a new Osmosis chain instance
func NewChain(name string) *Chain {
	return &Chain{Name: name}
}
