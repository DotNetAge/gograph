package graph

// Mutator defines the interface for modifying storage, satisfied by *tx.Transaction.
type Mutator interface {
        Put(key, value []byte) error
        Delete(key []byte) error
}
