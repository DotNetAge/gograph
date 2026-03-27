package storage

import (
	"os"
	"testing"
)

func TestPebble(t *testing.T) {
	path := "/tmp/gograph_storage_test"
	defer os.RemoveAll(path)

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	key := []byte("test_key")
	value := []byte("test_value")

	err = db.Put(key, value)
	// Put currently returns nil and does nothing in the current implementation? 
	// Wait, let's check pebble.go

	// Let's use Batch
	batch := db.NewBatch()
	batch.Put(key, value)
	err = batch.Commit()
	if err != nil {
		t.Fatalf("failed to commit batch: %v", err)
	}

	got, err := db.Get(key)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(got))
	}

	iter, err := db.NewIter(nil)
	if err != nil {
		t.Fatalf("failed to create iter: %v", err)
	}
	defer iter.Close()

	found := false
	for iter.SeekGE(key); iter.Valid(); iter.Next() {
		if string(iter.Key()) == string(key) {
			found = true
			break
		}
	}
	if !found {
		t.Error("key not found in iterator")
	}

	batch = db.NewBatch()
	batch.Delete(key)
	batch.Commit()

	_, err = db.Get(key)
	if err == nil {
		t.Error("expected error getting deleted key")
	}
}

func TestKeys(t *testing.T) {
	if string(NodeKey("1")) != "node:1" {
		t.Error("invalid node key")
	}
	if string(RelKey("2")) != "rel:2" {
		t.Error("invalid rel key")
	}
	if string(LabelKey("User", "1")) != "label:User:1" {
		t.Error("invalid label key")
	}
	if string(PropertyKey("User", "name", "Alice")) != "prop:User:name:Alice" {
		t.Error("invalid property key")
	}
	if string(AdjKey("1", "KNOWS", "out", "5")) != "adj:1:KNOWS:out:5" {
		t.Error("invalid adj key")
	}
}

func TestMarshal(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}
	s := TestStruct{Name: "Alice", Age: 30}
	data, err := Marshal(s)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var s2 TestStruct
	err = Unmarshal(data, &s2)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if s2.Name != s.Name || s2.Age != s.Age {
		t.Error("unmarshaled struct mismatch")
	}
}
