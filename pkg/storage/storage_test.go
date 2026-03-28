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
	if err != nil {
		t.Fatalf("failed to put key: %v", err)
	}

	got, err := db.Get(key)
	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(got))
	}

	err = db.Delete(key)
	if err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	_, err = db.Get(key)
	if err == nil {
		t.Error("expected error getting deleted key")
	}

	batch := db.NewBatch()
	batch.Put(key, value)
	err = batch.Commit()
	if err != nil {
		t.Fatalf("failed to commit batch: %v", err)
	}

	got, err = db.Get(key)
	if err != nil {
		t.Fatalf("failed to get key after batch commit: %v", err)
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

	iter.SeekLT([]byte("zzz"))
	if iter.Valid() {
		if string(iter.Key()) != string(key) {
			t.Error("expected to find key when seeking backwards")
		}
	}

	if iter.Value() == nil {
		t.Error("expected value from iterator")
	}

	if iter.Error() != nil {
		t.Errorf("unexpected iterator error: %v", iter.Error())
	}

	// Test iterator Next method with multiple keys
	key1 := []byte("key1")
	key2 := []byte("key2")
	key3 := []byte("key3")

	// Clear existing keys
	db.Delete(key)

	db.Put(key1, []byte("value1"))
	db.Put(key2, []byte("value2"))
	db.Put(key3, []byte("value3"))

	iter, err = db.NewIter(nil)
	if err != nil {
		t.Fatalf("failed to create iter: %v", err)
	}

	keys := []string{}
	for iter.SeekGE(key1); iter.Valid(); iter.Next() {
		keys = append(keys, string(iter.Key()))
	}

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	if keys[0] != "key1" || keys[1] != "key2" || keys[2] != "key3" {
		t.Errorf("expected keys [key1, key2, key3], got %v", keys)
	}

	// Test iterator Prev method
	iter.SeekLT([]byte("key4"))
	if !iter.Valid() {
		t.Error("expected iterator to be valid after SeekLT")
	}

	if string(iter.Key()) != "key3" {
		t.Errorf("expected key3, got %s", iter.Key())
	}

	if !iter.Prev() {
		t.Error("expected Prev to return true")
	}

	if string(iter.Key()) != "key2" {
		t.Errorf("expected key2, got %s", iter.Key())
	}

	if !iter.Prev() {
		t.Error("expected Prev to return true")
	}

	if string(iter.Key()) != "key1" {
		t.Errorf("expected key1, got %s", iter.Key())
	}

	if iter.Prev() {
		t.Error("expected Prev to return false at beginning")
	}

	iter.Close()

	batch = db.NewBatch()
	batch.Delete(key)
	batch.Commit()

	_, err = db.Get(key)
	if err == nil {
		t.Error("expected error getting deleted key")
	}

	batch.Close()
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

	if string(LabelKeyPrefix("User")) != "label:User:" {
		t.Error("invalid label key prefix")
	}
	if string(PropertyKeyPrefix("User", "name")) != "prop:User:name:" {
		t.Error("invalid property key prefix")
	}
	if string(AdjKeyPrefix("1")) != "adj:1:" {
		t.Error("invalid adj key prefix")
	}
	if string(AdjKeyPrefixNodeAndType("1", "KNOWS")) != "adj:1:KNOWS:" {
		t.Error("invalid adj key prefix node and type")
	}
	if string(AdjKeyPrefixNodeAndTypeAndDir("1", "KNOWS", "out")) != "adj:1:KNOWS:out:" {
		t.Error("invalid adj key prefix node and type and dir")
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
