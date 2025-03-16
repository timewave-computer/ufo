package transactions

// TxProcessor defines the interface for transaction processors
type TxProcessor interface {
	Process(tx *Transaction) error
}

// TxProcessorRegistry manages transaction processors
type TxProcessorRegistry struct {
	processors map[string]TxProcessor
}

// NewTxProcessorRegistry creates a new transaction processor registry
func NewTxProcessorRegistry() *TxProcessorRegistry {
	return &TxProcessorRegistry{
		processors: make(map[string]TxProcessor),
	}
}

// Register adds a processor to the registry
func (r *TxProcessorRegistry) Register(name string, processor TxProcessor) {
	r.processors[name] = processor
}

// RegisterProcessor is an alias for Register to match the interface expected by the app
func (r *TxProcessorRegistry) RegisterProcessor(name string, processor TxProcessor) {
	r.Register(name, processor)
}

// Get retrieves a processor from the registry
func (r *TxProcessorRegistry) Get(name string) TxProcessor {
	return r.processors[name]
}

// KVStoreTxProcessor is a simple key-value store transaction processor
type KVStoreTxProcessor struct {
	state interface{}
}

// NewKVStoreTxProcessor creates a new KV store transaction processor
func NewKVStoreTxProcessor(state interface{}) *KVStoreTxProcessor {
	return &KVStoreTxProcessor{
		state: state,
	}
}

// Process processes a transaction
func (p *KVStoreTxProcessor) Process(tx *Transaction) error {
	// Simplified implementation for stub
	return nil
}
