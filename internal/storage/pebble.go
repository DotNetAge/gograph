package storage

import (
	"github.com/cockroachdb/pebble"
)

type DB struct {
	db *pebble.DB
}

func Open(path string) (*DB, error) {
	db, err := pebble.Open(path, &pebble.Options{
		MemTableSize:            4 << 20,
		MaxConcurrentCompactions: func() int { return 2 },
	})
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Get(key []byte) ([]byte, error) {
	value, closer, err := d.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	result := make([]byte, len(value))
	copy(result, value)
	return result, nil
}

func (d *DB) Put(key, value []byte) error {
	return d.db.Set(key, value, pebble.NoSync)
}

func (d *DB) Delete(key []byte) error {
	return d.db.Delete(key, pebble.NoSync)
}

func (d *DB) NewBatch() *Batch {
	return &Batch{batch: d.db.NewBatch()}
}

func (d *DB) NewIter(opts *pebble.IterOptions) (*Iterator, error) {
	iter, err := d.db.NewIter(nil)
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
	return b.batch.Commit(pebble.NoSync)
}

func (b *Batch) Close() error {
	b.batch.Close()
	return nil
}

type Iterator struct {
	iter *pebble.Iterator
}

func (i *Iterator) SeekGE(key []byte) bool {
	return i.iter.SeekGE(key)
}

func (i *Iterator) SeekLT(key []byte) bool {
	return i.iter.SeekLT(key)
}

func (i *Iterator) Next() bool {
	return i.iter.Next()
}

func (i *Iterator) Prev() bool {
	return i.iter.Prev()
}

func (i *Iterator) Valid() bool {
	return i.iter.Valid()
}

func (i *Iterator) Key() []byte {
	return i.iter.Key()
}

func (i *Iterator) Value() []byte {
	return i.iter.Value()
}

func (i *Iterator) Error() error {
	return i.iter.Error()
}

func (i *Iterator) Close() error {
	return i.iter.Close()
}
