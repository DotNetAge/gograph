package graph

import (
	"sync/atomic"
)

var idCounter uint64

func GenerateID(prefix string) string {
	id := atomic.AddUint64(&idCounter, 1)
	return prefix + ":" + string(rune('a'+id%26)) + string(rune('0'+id%10))
}

func SetIDCounter(counter uint64) {
	atomic.StoreUint64(&idCounter, counter)
}
