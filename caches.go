package go_counter_map

import (
	"cloud.google.com/go/firestore"
	"sync"
	"sync/atomic"
)

func NewInt32Cache[K comparable](client *firestore.Client, docIdFunc func(key K) string, fieldName string) *CounterCache[K, int32] {
	c := CounterCache[K, int32]{
		readMap:   atomic.Pointer[map[K]*int32]{},
		dirtyMap:  make(map[K]*int32),
		lock:      sync.Mutex{},
		misses:    0,
		docID:     docIdFunc,
		fieldName: fieldName,
		add:       atomic.AddInt32,
		swap:      atomic.SwapInt32,
		client:    client,
	}
	rmap := make(map[K]*int32)
	c.readMap.Store(&rmap)
	return &c
}
func NewUInt32Cache[K comparable](client *firestore.Client, docIdFunc func(key K) string, fieldName string) *CounterCache[K, uint32] {
	c := CounterCache[K, uint32]{
		readMap:   atomic.Pointer[map[K]*uint32]{},
		dirtyMap:  make(map[K]*uint32),
		lock:      sync.Mutex{},
		misses:    0,
		docID:     docIdFunc,
		fieldName: fieldName,
		add:       atomic.AddUint32,
		swap:      atomic.SwapUint32,
		client:    client,
	}
	rmap := make(map[K]*uint32)
	c.readMap.Store(&rmap)
	return &c
}
func NewInt64Cache[K comparable](client *firestore.Client, docIdFunc func(key K) string, fieldName string) *CounterCache[K, int64] {
	c := CounterCache[K, int64]{
		readMap:   atomic.Pointer[map[K]*int64]{},
		dirtyMap:  make(map[K]*int64),
		lock:      sync.Mutex{},
		misses:    0,
		docID:     docIdFunc,
		fieldName: fieldName,
		add:       atomic.AddInt64,
		swap:      atomic.SwapInt64,
		client:    client,
	}
	rmap := make(map[K]*int64)
	c.readMap.Store(&rmap)
	return &c
}
func NewUInt64Cache[K comparable](client *firestore.Client, docIdFunc func(key K) string, fieldName string) *CounterCache[K, uint64] {
	c := CounterCache[K, uint64]{
		readMap:   atomic.Pointer[map[K]*uint64]{},
		dirtyMap:  make(map[K]*uint64),
		lock:      sync.Mutex{},
		misses:    0,
		docID:     docIdFunc,
		fieldName: fieldName,
		add:       atomic.AddUint64,
		swap:      atomic.SwapUint64,
		client:    client,
	}
	rmap := make(map[K]*uint64)
	c.readMap.Store(&rmap)
	return &c
}