// Package storage provides the底层 storage layer for gograph using Pebble as the
// underlying key-value store.
package storage

import (
	"github.com/vmihailenco/msgpack/v5"
)

// Marshal encodes a value to bytes using MessagePack encoding.
// MessagePack is more compact and faster than Gob/JSON.
func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal decodes bytes to a value using MessagePack encoding.
func Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
