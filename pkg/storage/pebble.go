package storage

import (
	"github.com/cockroachdb/pebble"
)

type noopLogger struct{}

func (l noopLogger) Infof(format string, args ...interface{})  {}
func (l noopLogger) Fatalf(format string, args ...interface{}) {}

type DB struct {
	db *pebble.DB
}

// Open opens a database at the given path and returns a DB instance.
func Open(path string) (*DB, error) {
	db, err := pebble.Open(path, &pebble.Options{
		Logger: noopLogger{},
		MemTableSize:            4 << 20,
		MaxConcurrentCompactions: func() int { return 2 },
	})
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Get(key []byte) ([]byte, error) {
	val, closer, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	
	ret := make([]byte, len(val))
	copy(ret, val)
	return ret, nil
}

func (db *DB) Put(key, value []byte) error {
	return db.db.Set(key, value, pebble.Sync)
}

func (db *DB) Delete(key []byte) error {
	return db.db.Delete(key, pebble.Sync)
}

func (db *DB) NewBatch() *Batch {
	return &Batch{batch: db.db.NewBatch()}
}

func (db *DB) NewIter(opts *pebble.IterOptions) (*Iterator, error) {
	iter, err := db.db.NewIter(opts)
	if err != nil {
		return nil, err
	}
	return &Iterator{iter: iter}, nil
}

type Batch struct {
	batch *pebble.Batch
}

func (b *Batch) Put(key, value []byte) error {
	return b.batch.Set(key, value, nil)
}

func (b *Batch) Delete(key []byte) error {
	return b.batch.Delete(key, nil)
}

func (b *Batch) Commit() error {
	return b.batch.Commit(pebble.Sync)
}

func (b *Batch) Close() error {
	return b.batch.Close()
}

type Iterator struct {
	iter *pebble.Iterator
}

func (i *Iterator) SeekGE(key []byte) bool { return i.iter.SeekGE(key) }
func (i *Iterator) SeekLT(key []byte) bool { return i.iter.SeekLT(key) }
func (i *Iterator) Next() bool             { return i.iter.Next() }
func (i *Iterator) Prev() bool             { return i.iter.Prev() }
func (i *Iterator) Valid() bool            { return i.iter.Valid() }
func (i *Iterator) Key() []byte            { return i.iter.Key() }
func (i *Iterator) Value() []byte          { return i.iter.Value() }
func (i *Iterator) Error() error           { return i.iter.Error() }
func (i *Iterator) Close() error           { return i.iter.Close() }
