package dao_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheGetSet(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 1 * time.Second, MaxItems: 10})

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	require.True(t, ok)
	assert.Equal(t, "value1", v)
}

func TestCacheGetMiss(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 1 * time.Second, MaxItems: 10})

	v, ok := c.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestCacheTTLExpiration(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 50 * time.Millisecond, MaxItems: 10})

	c.Set("key1", "value1")

	// Should be present immediately.
	v, ok := c.Get("key1")
	require.True(t, ok)
	assert.Equal(t, "value1", v)

	// Wait for TTL to expire.
	time.Sleep(100 * time.Millisecond)

	v, ok = c.Get("key1")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestCacheInvalidateKeys(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 10})

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	c.Invalidate("key1", "key3")

	_, ok := c.Get("key1")
	assert.False(t, ok)

	v, ok := c.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", v)

	_, ok = c.Get("key3")
	assert.False(t, ok)
}

func TestCacheInvalidatePrefix(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 100})

	c.Set("nodes:list", "node-data")
	c.Set("nodes:get:node1", "n1-data")
	c.Set("nodes:get:node2", "n2-data")
	c.Set("profiles:list", "profile-data")
	c.Set("images:list", "image-data")

	c.InvalidatePrefix("nodes:")

	_, ok := c.Get("nodes:list")
	assert.False(t, ok)
	_, ok = c.Get("nodes:get:node1")
	assert.False(t, ok)
	_, ok = c.Get("nodes:get:node2")
	assert.False(t, ok)

	// Other prefixes should be untouched.
	v, ok := c.Get("profiles:list")
	assert.True(t, ok)
	assert.Equal(t, "profile-data", v)

	v, ok = c.Get("images:list")
	assert.True(t, ok)
	assert.Equal(t, "image-data", v)
}

func TestCacheInvalidateAll(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 100})

	c.Set("key1", "v1")
	c.Set("key2", "v2")
	c.Set("key3", "v3")

	c.InvalidateAll()

	_, ok := c.Get("key1")
	assert.False(t, ok)
	_, ok = c.Get("key2")
	assert.False(t, ok)
	_, ok = c.Get("key3")
	assert.False(t, ok)
}

func TestCacheMaxItemsEviction(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 3})

	c.Set("key1", "v1")
	time.Sleep(1 * time.Millisecond)
	c.Set("key2", "v2")
	time.Sleep(1 * time.Millisecond)
	c.Set("key3", "v3")
	time.Sleep(1 * time.Millisecond)

	// Adding a 4th item should evict the oldest (key1).
	c.Set("key4", "v4")

	_, ok := c.Get("key1")
	assert.False(t, ok, "oldest entry should have been evicted")

	v, ok := c.Get("key4")
	assert.True(t, ok)
	assert.Equal(t, "v4", v)
}

func TestCacheStats(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 10})

	c.Set("key1", "v1")

	// Hit.
	c.Get("key1")
	// Miss.
	c.Get("missing")

	hits, misses := c.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)
}

func TestCacheDefaultConfig(t *testing.T) {
	cfg := dao.DefaultCacheConfig()
	assert.Equal(t, 10*time.Second, cfg.TTL)
	assert.Equal(t, 200, cfg.MaxItems)
}

func TestCacheZeroConfigDefaults(t *testing.T) {
	// Zero values should get sane defaults.
	c := dao.NewCache(dao.CacheConfig{})

	c.Set("key1", "v1")
	v, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "v1", v)
}

func TestCacheConcurrentAccess(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 1 * time.Second, MaxItems: 100})

	const goroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				c.Set(key, j)
				c.Get(key)
				if j%10 == 0 {
					c.InvalidatePrefix(fmt.Sprintf("key-%d-", id))
				}
			}
		}(i)
	}

	wg.Wait()
	// If we get here without a race condition, the test passes.
}

func TestCacheConcurrentReadWrite(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 1 * time.Second, MaxItems: 100})

	const writers = 10
	const readers = 20
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Writers.
	for i := 0; i < writers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				c.Set(fmt.Sprintf("key-%d", j), id)
			}
		}(i)
	}

	// Readers.
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				c.Get(fmt.Sprintf("key-%d", j))
			}
		}()
	}

	wg.Wait()
}

func TestCacheConcurrentInvalidateAll(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 1 * time.Second, MaxItems: 100})

	var wg sync.WaitGroup
	wg.Add(3)

	// Writer.
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			c.Set(fmt.Sprintf("key-%d", i), i)
		}
	}()

	// Invalidator.
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			c.InvalidateAll()
		}
	}()

	// Reader.
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			c.Get(fmt.Sprintf("key-%d", i))
		}
	}()

	wg.Wait()
}

// ---------- Cached Client Integration ----------

func TestCachedClientWriteInvalidatesCache(t *testing.T) {
	// Use the mock client as the inner client to verify cache behavior.
	// We can't import mock from here (it imports dao), so we test the
	// cache mechanics directly.
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 100})

	// Simulate caching a nodes:list result.
	c.Set("nodes:list", "cached-nodes")

	v, ok := c.Get("nodes:list")
	require.True(t, ok)
	assert.Equal(t, "cached-nodes", v)

	// Simulate a write operation invalidating the prefix.
	c.InvalidatePrefix("nodes:")

	_, ok = c.Get("nodes:list")
	assert.False(t, ok, "cache should be invalidated after write")
}

func TestCachedClientOverwriteValue(t *testing.T) {
	c := dao.NewCache(dao.CacheConfig{TTL: 10 * time.Second, MaxItems: 100})

	c.Set("key1", "original")
	c.Set("key1", "updated")

	v, ok := c.Get("key1")
	require.True(t, ok)
	assert.Equal(t, "updated", v)
}
