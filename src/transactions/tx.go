// Package transactions provides transaction handling
package transactions

// Transaction represents a blockchain transaction
type Transaction struct {
	Hash     string
	From     string
	To       string
	Amount   int64
	Sequence uint64
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount int64, sequence uint64) *Transaction {
	return &Transaction{
		Hash:     "tx-hash", // Simplified for stub
		From:     from,
		To:       to,
		Amount:   amount,
		Sequence: sequence,
	}
}
