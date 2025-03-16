package consensus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Transaction represents a basic transaction in the system
type Transaction []byte

// Block represents a block in the blockchain
type Block struct {
	Height      int64
	Transactions []Transaction
	Proposer    *Validator
	Timestamp   time.Time
	Hash        string
	PrevHash    string
}

// NewBlock creates a new block with the given parameters
func NewBlock(height int64, txs []Transaction, proposer *Validator, prevHash string) *Block {
	block := &Block{
		Height:      height,
		Transactions: txs,
		Proposer:    proposer,
		Timestamp:   time.Now(),
		PrevHash:    prevHash,
	}
	block.Hash = block.CalculateHash()
	return block
}

// CalculateHash computes the hash of the block
func (b *Block) CalculateHash() string {
	// Simple hash calculation for demonstration purposes
	data := fmt.Sprintf("%d-%v-%s-%s", b.Height, b.Timestamp, b.PrevHash, b.Proposer.ID)
	for _, tx := range b.Transactions {
		data += string(tx)
	}
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// String returns a string representation of the block
func (b *Block) String() string {
	return fmt.Sprintf("Block{Height: %d, Hash: %s, PrevHash: %s, Proposer: %s, TxCount: %d}",
		b.Height, b.Hash, b.PrevHash, b.Proposer.ID, len(b.Transactions))
} 