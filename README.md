# Firestore Counter Cache

Firestore has limitations on how often one document can be updated + frequent updates add up to costs.
So if there is no need in realtime updates, we can cache counter delta in memory and then commit to Firestore.

It is basically simpler implementation of `sync.Map`.

Example usage:
```go
// Init cache using one of the methods from caches.go
cache := NewUInt32Cache[string](db, func (k string) string {
    return fmt.Sprintf("collection/%s", k)
}, "Counter field name")
// Atomically add delta. If entry does not exist, it is automatically created
cache.Update("yourawesomeid", 1)
// Commit changes to database (increments fields). May be scheduled to run periodically or invoked when needed
// During commit operation dirty map is locked. This will slow down some operations
cache.Commit(context)
```

Benchmark:
```
goos: windows
goarch: amd64
pkg: go-counter-map
cpu: Intel(R) Core(TM) i5-9500F CPU @ 3.00GHz
BenchmarkMultiThread
BenchmarkMultiThread-6          51986534                23.65 ns/op            4 B/op          1 allocs/op
PASS
```