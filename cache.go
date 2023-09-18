package firestore_counter_cache

import (
	"cloud.google.com/go/firestore"
	"context"
	"maps"
	"sync"
	"sync/atomic"
)

/*
*
We do not support uint8 because it does not support atomic operations.
*/
type number interface {
	int32 | uint32 | int64 | uint64
}

type CounterCache[K comparable, V number] struct {
	readMap atomic.Pointer[map[K]*V]

	dirtyMap map[K]*V
	lock     sync.Mutex
	misses   int

	docID     func(key K) (id string)
	fieldName string

	add  func(addr *V, delta V) (old V)
	swap func(addr *V, new V) (old V)

	client *firestore.Client
}

// Update
// Updates value in cache or creates entry if not already created
// /**
func (c *CounterCache[K, V]) Update(key K, delta V) (old V) {
	//Timestamp update is always required
	m := c.readMap.Load()
	val, ok := (*m)[key]
	if ok {
		return c.add(val, delta)
	}
	c.lock.Lock()
	// Check if dirty map was promoted
	val, ok = (*c.readMap.Load())[key]
	if ok {
		// If map was promoted, do not report miss or try to create key in dirty map
		c.lock.Unlock()
		return c.add(val, delta)
	} else {
		// Access dirty map is still no key in readonly map
		val, ok = c.dirtyMap[key]
		if ok {
			c.miss()
			c.lock.Unlock()
			return c.add(val, delta)
		}
		// Add new entry to dirty map
		val = &delta
		c.dirtyMap[key] = val
		c.lock.Unlock()
		return delta
	}
}

// Commit changes to database
// Entries with zero delta will be moved to dirty map and deleted during next commit operation if they will stay zero
func (c *CounterCache[K, V]) Commit(ctx context.Context) {

	writer := c.client.BulkWriter(ctx)
	// We lock to make sure that dirty map can be cleaned up safely
	c.lock.Lock()

	// Iterating over dirty map to clean it and commit changes
	for k, v := range c.dirtyMap {
		val := c.swap(v, 0)
		//Delete entries with zero delta
		if val == 0 {
			delete(c.dirtyMap, k)
		}
		// Add update to queue
		_, _ = writer.Update(c.client.Doc(c.docID(k)), []firestore.Update{
			{
				Path:  c.fieldName,
				Value: firestore.Increment(val),
			},
		})
	}
	var newReadonly map[K]*V
	if c.misses >= len(c.dirtyMap) {
		newReadonly = c.dirtyMap
		c.dirtyMap = make(map[K]*V)
		c.misses = 0
	} else {
		newReadonly = make(map[K]*V)
	}

	//Iterating over read map. Because commit operation is blocking, read map cannot be changed during process => no data skipped
	for k, v := range *c.readMap.Load() {
		val := c.swap(v, 0)
		if val == 0 {
			// We move entries with zero delta to dirty map instead of completely deleting them.
			// So if value is updated after condition fired, changes will not be lost.
			// If value stays zero, entry will be removed from dirty map during next commit.
			c.dirtyMap[k] = v
			continue
		}
		//Update and keep entry in readOnly map
		_, _ = writer.Update(c.client.Doc(c.docID(k)), []firestore.Update{
			{
				Path:  c.fieldName,
				Value: firestore.Increment(val),
			},
		})
		newReadonly[k] = v
	}
	// Swap old map with new one. As said above, no data can be lost because we move entries to dirty map before final deletion
	c.readMap.Store(&newReadonly)
	// Unlock before flushing writes because we do not need maps anymore
	c.lock.Unlock()
	writer.End()
}

// Call only when locked. Records a miss and merges dirty map to readonly map if the misses count is greater than length of dirty map
func (c *CounterCache[K, V]) miss() {
	c.misses++
	if c.misses < len(c.dirtyMap) {
		return
	}
	rmap := *c.readMap.Load()
	// Create new readonly map
	newRmap := maps.Clone(rmap)
	maps.Copy(newRmap, c.dirtyMap)
	c.readMap.Store(&newRmap)
	//Reset dirty map
	c.dirtyMap = make(map[K]*V)
	c.misses = 0
}
