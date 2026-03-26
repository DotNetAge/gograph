package tx

import (
	"os"
	"testing"

	"github.com/DotNetAge/gograph/internal/storage"
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

	// Check that it's not in store yet
	_, err = store.Get(key)
	if err == nil {
		t.Error("expected key not to be in store before commit")
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	// Now it should be in store
	got, err := store.Get(key)
	if err != nil {
		t.Fatalf("failed to get key after commit: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(got))
	}

	// Test Rollback
	tx, _ = mgr.Begin(false)
	tx.Put([]byte("roll"), []byte("back"))
	tx.Rollback()

	_, err = store.Get([]byte("roll"))
	if err == nil {
		t.Error("expected key not to be in store after rollback")
	}
}
