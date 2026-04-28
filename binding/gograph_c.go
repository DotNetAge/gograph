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

//export database_new
func database_new(dbPath *C.char, errInfo *C.ErrorInfo) C.DatabaseHandle {
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

//export database_open
func database_open(dbPath *C.char, errInfo *C.ErrorInfo) C.DatabaseHandle {
	return database_new(dbPath, errInfo)
}

//export database_close
func database_close(handle C.DatabaseHandle) C.int {
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

//export database_free
func database_free(handle C.DatabaseHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(databaseRegistry, uint64(uintptr(handle)))
}

// ============================================================================
// Transaction Management API Implementation
// ============================================================================

//export transaction_begin
func transaction_begin(
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

//export transaction_commit
func transaction_commit(txHandle C.TransactionHandle, errInfo *C.ErrorInfo) C.int {
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

//export transaction_rollback
func transaction_rollback(txHandle C.TransactionHandle) C.int {
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

//export transaction_free
func transaction_free(txHandle C.TransactionHandle) {
	registryMu.Lock()
	defer registryMu.Unlock()

	delete(transactionRegistry, uint64(uintptr(txHandle)))
}

// ============================================================================
// Cypher Query API Implementation
// ============================================================================

//export cypher_execute
func cypher_execute(
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

	queryStr := C.GoString(query)

	result.columns = nil
	result.column_count = 0
	result.rows = nil
	result.row_count = 0

	_ = queryStr

	return C.GOGRAPH_OK
}

//export node_create
func node_create(
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

	return C.GOGRAPH_OK
}

//export node_get
func node_get(
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

	return C.GOGRAPH_OK
}

//export node_delete
func node_delete(
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

	return C.GOGRAPH_OK
}

//export relationship_create
func relationship_create(
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

	return C.GOGRAPH_OK
}

//export relationship_get
func relationship_get(
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

	return C.GOGRAPH_OK
}

// ============================================================================
// Memory Management API Implementation
// ============================================================================

//export query_result_free
func query_result_free(result *C.QueryResult) {
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

//export node_free
func node_free(node *C.Node) {
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

//export relationship_free
func relationship_free(rel *C.Relationship) {
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

//export path_free
func path_free(path *C.Path) {
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

//export value_free
func value_free(val *C.Value) {
	if val == nil {
		return
	}
}

//export error_free
func error_free(errInfo *C.ErrorInfo) {
	if errInfo != nil && errInfo.message != nil {
		C.free(unsafe.Pointer(errInfo.message))
	}
}
