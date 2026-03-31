// Package tx provides transaction management for gograph's storage layer.
package tx

import "errors"

var (
	// ErrTransactionCommitted is returned when attempting to use a committed transaction.
	ErrTransactionCommitted = errors.New("transaction already committed")

	// ErrTransactionRolledBack is returned when attempting to use a rolled back transaction.
	ErrTransactionRolledBack = errors.New("transaction already rolled back")
)
