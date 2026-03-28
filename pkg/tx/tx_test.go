package tx

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/pkg/storage"
)

func TestTransaction(t *testing.T) {
	path := "/tmp/gograph_tx_test"
	defer os.RemoveAll(path)

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	mgr := NewManager(store)
	tx, err := mgr.Begin(false)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	key := []byte("tx_key")
	value := []byte("tx_value")

	err = tx.Put(key, value)
	if err != nil {
		t.Fatalf("failed to put in tx: %v", err)
	}

	if tx.IsReadOnly() {
		t.Error("expected transaction to not be read-only")
	}

	if tx.IsClosed() {
		t.Error("expected transaction to not be closed")
	}

	_, err = store.Get(key)
	if err == nil {
		t.Error("expected key not to be in store before commit")
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	if !tx.IsClosed() {
		t.Error("expected transaction to be closed after commit")
	}

	_, err = tx.Get(key)
	if err == nil {
		t.Error("expected error getting from closed transaction")
	}

	err = tx.Put(key, value)
	if err == nil {
		t.Error("expected error putting to closed transaction")
	}

	err = tx.Delete(key)
	if err == nil {
		t.Error("expected error deleting from closed transaction")
	}

	got, err := store.Get(key)
	if err != nil {
		t.Fatalf("failed to get key after commit: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(got))
	}

	tx, _ = mgr.Begin(false)
	err = tx.Delete(key)
	if err != nil {
		t.Fatalf("failed to delete in tx: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit delete tx: %v", err)
	}

	_, err = store.Get(key)
	if err == nil {
		t.Error("expected error getting deleted key")
	}

	tx, _ = mgr.Begin(false)
	tx.Put([]byte("roll"), []byte("back"))
	tx.Rollback()

	if !tx.IsClosed() {
		t.Error("expected transaction to be closed after rollback")
	}

	_, err = store.Get([]byte("roll"))
	if err == nil {
		t.Error("expected key not to be in store after rollback")
	}

	tx, _ = mgr.Begin(false)
	tx.Rollback()

	tx, _ = mgr.Begin(true)
	if !tx.IsReadOnly() {
		t.Error("expected read-only transaction")
	}

	err = tx.Put([]byte("ro"), []byte("test"))
	if err == nil {
		t.Error("expected error putting to read-only transaction")
	}

	err = tx.Delete([]byte("ro"))
	if err == nil {
		t.Error("expected error deleting from read-only transaction")
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit read-only tx: %v", err)
	}

	if !tx.IsClosed() {
		t.Error("expected read-only transaction to be closed after commit")
	}
}
