// Package tx provides transaction management for gograph's storage layer.
package tx

import "errors"

var (
	ErrTransactionClosed   = errors.New("transaction closed")
	ErrReadOnlyTransaction = errors.New("read only transaction")
)
