package alloy

import (
	"sync"
	"testing"
)

func BenchmarkClearBundleCache(b *testing.B) {
	for i := 0; i < 100; i++ {
		bundleCache.Store("bundle"+string(rune(i))+".js", "test content")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClearBundleCache()
		for j := 0; j < 100; j++ {
			bundleCache.Store("bundle"+string(rune(j))+".js", "test content")
		}
	}
}

func BenchmarkConcurrentBundleCacheAccess(b *testing.B) {
	bundleCache.Store("shared.js", "shared content")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bundleCache.Load("shared.js")
		}
	})
}

func BenchmarkConcurrentBundleCacheStores(b *testing.B) {
	counter := 0
	mu := sync.Mutex{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			key := "bundle" + string(rune(counter%10))
			mu.Unlock()
			bundleCache.Store(key, "content")
		}
	})
}
