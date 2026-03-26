package storage

import (
	"bytes"
	"encoding/gob"
)

func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	return dec.Decode(v)
}
