package firestore_counter_cache

import (
	"math/rand"
	"testing"
)

func BenchmarkSingleThread(b *testing.B) {
	b.ReportAllocs()
	cache := NewUInt32Cache[int](nil, nil, "")
	key := rand.Int()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Update(key, 1)
	}
}

func BenchmarkMultiThread(b *testing.B) {
	b.ReportAllocs()
	cache := NewUInt32Cache[int](nil, nil, "")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Update(123456789, 1)
		}
	})
}
