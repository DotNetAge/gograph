// Package storage provides the底层 storage layer for gograph using Pebble as the
// underlying key-value store.
package storage

import (
	"bytes"
	"encoding/gob"
)

// Marshal encodes a value to bytes using Gob encoding.
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes bytes to a value using Gob encoding.
func Unmarshal(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	return dec.Decode(v)
}
