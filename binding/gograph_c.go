// Package binding provides CGO bindings for GoGraph, enabling cross-language integration.
//
// This package serves as an adapter layer between GoGraph's pure Go core
// and C-based languages (Python, C++, Java, etc.) through SWIG.
//
// Architecture:
//   - Pure Go core: github.com/DotNetAge/gograph/pkg/* (no CGO dependencies)
//   - CGO adapter: this package (gograph/binding)
//   - One-way dependency: binding -> core (no circular dependencies)
package binding

/*
#include "gograph_c.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/DotNetAge/gograph/pkg/storage"
	"github.com/DotNetAge/gograph/pkg/tx"
)

// ============================================================================
// Global Registry - Manages handles for C code
// ============================================================================

var (
	databaseRegistry    = make(map[uint64]*storage.DB)
	transactionRegistry = make(map[uint64]*tx.Transaction)
	registryMu          sync.RWMutex

	nextDatabaseID    uint64 = 1
	nextTransactionID uint64 = 1
)

// ============================================================================
// Type Conversions
// ============================================================================

// newError creates a new ErrorInfo with the given code and message
func newError(code C.int, msg string) C.ErrorInfo {
	return C.ErrorInfo{
		code:    code,
		message: C.CString(msg),
	}
}

// ============================================================================
// Database Management API Implementation
// ============================================================================

// gograph_database_new creates a new database instance.
func gograph_database_new(dbPath *C.char, errInfo *C.ErrorInfo) C.DatabaseHandle {
	path := C.GoString(dbPath)

	db, err := storage.Open(path)
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_STORAGE_FAILURE, err.Error())
		}
		return nil
	}

	registryMu.Lock()
	handle := nextDatabaseID
	nextDatabaseID++
	databaseRegistry[handle] = db
	registryMu.Unlock()

	return C.DatabaseHandle(uintptr(handle))
}

// gograph_database_open opens an existing database.
func gograph_database_open(dbPath *C.char, errInfo *C.ErrorInfo) C.DatabaseHandle {
	// For now, same as new - can be extended later
	return gograph_database_new(dbPath, errInfo)
}

// gograph_database_close closes the database.
func gograph_database_close(handle C.DatabaseHandle) C.int {
	registryMu.Lock()
	defer registryMu.Unlock()

	db, exists := databaseRegistry[uint64(uintptr(handle))]
	if !exists {
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	if err := db.Close(); err != nil {
		return C.GOGRAPH_ERROR_STORAGE_FAILURE
	}

	delete(databaseRegistry, uint64(uintptr(handle)))
	return C.GOGRAPH_OK
}

// gograph_database_free frees the database handle.
func gograph_database_free(handle C.DatabaseHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(databaseRegistry, uint64(uintptr(handle)))
}

// ============================================================================
// Transaction Management API Implementation
// ============================================================================

func gograph_transaction_begin(
	dbHandle C.DatabaseHandle,
	readOnly C.bool,
	errInfo *C.ErrorInfo,
) C.TransactionHandle {
	registryMu.RLock()
	db := databaseRegistry[uint64(uintptr(dbHandle))]
	registryMu.RUnlock()

	if db == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "database not found")
		}
		return nil
	}

	txManager := tx.NewManager(db)
	tx, err := txManager.Begin(bool(readOnly))
	if err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_EXEC_FAILURE, err.Error())
		}
		return nil
	}

	registryMu.Lock()
	handle := nextTransactionID
	nextTransactionID++
	transactionRegistry[handle] = tx
	registryMu.Unlock()

	return C.TransactionHandle(uintptr(handle))
}

func gograph_transaction_commit(txHandle C.TransactionHandle, errInfo *C.ErrorInfo) C.int {
	registryMu.Lock()
	defer registryMu.Unlock()

	tx, exists := transactionRegistry[uint64(uintptr(txHandle))]
	if !exists {
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	if err := tx.Commit(); err != nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_EXEC_FAILURE, err.Error())
		}
		return C.GOGRAPH_ERROR_EXEC_FAILURE
	}

	delete(transactionRegistry, uint64(uintptr(txHandle)))
	return C.GOGRAPH_OK
}

func gograph_transaction_rollback(txHandle C.TransactionHandle) C.int {
	registryMu.Lock()
	defer registryMu.Unlock()

	tx, exists := transactionRegistry[uint64(uintptr(txHandle))]
	if !exists {
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	tx.Rollback()
	delete(transactionRegistry, uint64(uintptr(txHandle)))
	return C.GOGRAPH_OK
}

func gograph_transaction_free(txHandle C.TransactionHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(transactionRegistry, uint64(uintptr(txHandle)))
}

// ============================================================================
// Cypher Query API Implementation
// ============================================================================

func gograph_cypher_execute(
	txHandle C.TransactionHandle,
	query *C.char,
	params *C.char,
	result *C.QueryResult,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual Cypher query execution
	// This is a placeholder - the real implementation would use the cypher package

	queryStr := C.GoString(query)

	// Placeholder response
	result.columns = nil
	result.column_count = 0
	result.rows = nil
	result.row_count = 0

	_ = queryStr // Use variable to avoid compiler warning

	return C.GOGRAPH_OK
}

func gograph_node_create(
	txHandle C.TransactionHandle,
	label *C.char,
	properties *C.char,
	nodeId *C.uint64_t,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual node creation
	// This is a placeholder

	return C.GOGRAPH_OK
}

func gograph_node_get(
	txHandle C.TransactionHandle,
	nodeId C.uint64_t,
	node *C.Node,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual node retrieval
	// This is a placeholder

	return C.GOGRAPH_OK
}

func gograph_node_delete(
	txHandle C.TransactionHandle,
	nodeId C.uint64_t,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual node deletion
	// This is a placeholder

	return C.GOGRAPH_OK
}

func gograph_relationship_create(
	txHandle C.TransactionHandle,
	relType *C.char,
	startNodeId C.uint64_t,
	endNodeId C.uint64_t,
	properties *C.char,
	relId *C.uint64_t,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual relationship creation
	// This is a placeholder

	return C.GOGRAPH_OK
}

func gograph_relationship_get(
	txHandle C.TransactionHandle,
	relId C.uint64_t,
	rel *C.Relationship,
	errInfo *C.ErrorInfo,
) C.int {
	registryMu.RLock()
	tx := transactionRegistry[uint64(uintptr(txHandle))]
	registryMu.RUnlock()

	if tx == nil {
		if errInfo != nil {
			*errInfo = newError(C.GOGRAPH_ERROR_NOT_FOUND, "transaction not found")
		}
		return C.GOGRAPH_ERROR_NOT_FOUND
	}

	// TODO: Implement actual relationship retrieval
	// This is a placeholder

	return C.GOGRAPH_OK
}

// ============================================================================
// Memory Management API Implementation
// ============================================================================

func gograph_query_result_free(result *C.QueryResult) {
	if result == nil {
		return
	}

	if result.columns != nil {
		C.free(unsafe.Pointer(result.columns))
	}

	if result.rows != nil {
		C.free(unsafe.Pointer(result.rows))
	}
}

func gograph_node_free(node *C.Node) {
	if node == nil {
		return
	}

	if node.label != nil {
		C.free(unsafe.Pointer(node.label))
	}

	if node.properties != nil {
		C.free(unsafe.Pointer(node.properties))
	}
}

func gograph_relationship_free(rel *C.Relationship) {
	if rel == nil {
		return
	}

	if rel._type != nil {
		C.free(unsafe.Pointer(rel._type))
	}

	if rel.properties != nil {
		C.free(unsafe.Pointer(rel.properties))
	}
}

func gograph_path_free(path *C.Path) {
	if path == nil {
		return
	}

	if path.nodes != nil {
		C.free(unsafe.Pointer(path.nodes))
	}

	if path.relationships != nil {
		C.free(unsafe.Pointer(path.relationships))
	}
}

func gograph_value_free(val *C.Value) {
	if val == nil {
		return
	}

	// TODO: Implement proper union handling for C.Value
	// For now, just return since we're not using Value types yet
}

func gograph_error_free(errInfo *C.ErrorInfo) {
	if errInfo != nil && errInfo.message != nil {
		C.free(unsafe.Pointer(errInfo.message))
	}
}
