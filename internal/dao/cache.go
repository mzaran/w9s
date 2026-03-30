package dao

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CacheConfig controls caching behavior.
type CacheConfig struct {
	TTL      time.Duration
	MaxItems int
}

// DefaultCacheConfig returns sensible defaults for the cache.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:      10 * time.Second,
		MaxItems: 200,
	}
}

// cachedResult holds a cached value with its insertion timestamp.
type cachedResult[T any] struct {
	value     T
	timestamp time.Time
}

// Cache is a generic TTL-based cache with bounded size.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*cachedResult[any]
	config  CacheConfig
	hits    atomic.Int64
	misses  atomic.Int64
}

// NewCache creates a new Cache with the given configuration.
func NewCache(cfg CacheConfig) *Cache {
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Second
	}
	if cfg.MaxItems <= 0 {
		cfg.MaxItems = 200
	}
	return &Cache{
		items:  make(map[string]*cachedResult[any]),
		config: cfg,
	}
}

// Get retrieves a value from the cache. Returns the value and true if found and not expired.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	if !ok {
		c.mu.RUnlock()
		c.misses.Add(1)
		return nil, false
	}
	if time.Since(entry.timestamp) > c.config.TTL {
		c.mu.RUnlock()
		c.mu.Lock()
		if e, exists := c.items[key]; exists && time.Since(e.timestamp) > c.config.TTL {
			delete(c.items, key)
		}
		c.mu.Unlock()
		c.misses.Add(1)
		return nil, false
	}
	v := entry.value
	c.mu.RUnlock()
	c.hits.Add(1)
	return v, true
}

// Set stores a value in the cache.
func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.config.MaxItems {
		c.evictOldest()
	}
	c.items[key] = &cachedResult[any]{value: value, timestamp: time.Now()}
}

// Invalidate removes specific keys from the cache.
func (c *Cache) Invalidate(keys ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, key := range keys {
		delete(c.items, key)
	}
}

// InvalidatePrefix removes all entries whose keys start with the given prefix.
func (c *Cache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
		}
	}
}

// InvalidateAll clears the entire cache.
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cachedResult[any])
}

// Stats returns hit and miss counts.
func (c *Cache) Stats() (hits, misses int64) {
	return c.hits.Load(), c.misses.Load()
}

// evictOldest removes the oldest entry. Caller must hold the write lock.
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range c.items {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// --- Cached client wrapper ---

type cachedClient struct {
	inner    WarewulfClient
	cache    *Cache
	nodes    NodeManager
	profiles ProfileManager
	images   ImageManager
	overlays OverlayManager
}

// NewCachedClient wraps a WarewulfClient with TTL-based caching for read operations.
func NewCachedClient(inner WarewulfClient, cfg CacheConfig) WarewulfClient {
	cache := NewCache(cfg)
	return &cachedClient{
		inner:    inner,
		cache:    cache,
		nodes:    &cachedNodeManager{inner: inner.Nodes(), cache: cache},
		profiles: &cachedProfileManager{inner: inner.Profiles(), cache: cache},
		images:   &cachedImageManager{inner: inner.Images(), cache: cache},
		overlays: &cachedOverlayManager{inner: inner.Overlays(), cache: cache},
	}
}

func (c *cachedClient) Nodes() NodeManager       { return c.nodes }
func (c *cachedClient) Profiles() ProfileManager  { return c.profiles }
func (c *cachedClient) Images() ImageManager      { return c.images }
func (c *cachedClient) Overlays() OverlayManager  { return c.overlays }
func (c *cachedClient) Power() PowerManager       { return c.inner.Power() }
func (c *cachedClient) ServerInfo() (*ServerInfo, error) { return c.inner.ServerInfo() }
func (c *cachedClient) HasPower() bool            { return c.inner.HasPower() }
func (c *cachedClient) Close() error              { return c.inner.Close() }

// --- cachedNodeManager ---

type cachedNodeManager struct {
	inner NodeManager
	cache *Cache
}

func (m *cachedNodeManager) List() (map[string]*WwNode, error) {
	const key = "nodes:list"
	if v, ok := m.cache.Get(key); ok {
		return v.(map[string]*WwNode), nil
	}
	result, err := m.inner.List()
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedNodeManager) Get(id string) (*WwNode, error) {
	key := fmt.Sprintf("nodes:get:%s", id)
	if v, ok := m.cache.Get(key); ok {
		return v.(*WwNode), nil
	}
	result, err := m.inner.Get(id)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedNodeManager) GetRaw(id string) (*WwNode, error) {
	key := fmt.Sprintf("nodes:raw:%s", id)
	if v, ok := m.cache.Get(key); ok {
		return v.(*WwNode), nil
	}
	result, err := m.inner.GetRaw(id)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedNodeManager) GetFields(id string) ([]NodeField, error) {
	key := fmt.Sprintf("nodes:fields:%s", id)
	if v, ok := m.cache.Get(key); ok {
		return v.([]NodeField), nil
	}
	result, err := m.inner.GetFields(id)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedNodeManager) GetOverlayInfo(id string) (*NodeOverlayInfo, error) {
	return m.inner.GetOverlayInfo(id)
}

func (m *cachedNodeManager) BuildOverlays(id string) error {
	m.cache.InvalidatePrefix("nodes:")
	return m.inner.BuildOverlays(id)
}

func (m *cachedNodeManager) BuildAllOverlays() error {
	m.cache.InvalidatePrefix("nodes:")
	return m.inner.BuildAllOverlays()
}

func (m *cachedNodeManager) Add(id string, node *WwNode) error {
	m.cache.InvalidatePrefix("nodes:")
	return m.inner.Add(id, node)
}

func (m *cachedNodeManager) Update(id string, node *WwNode) error {
	m.cache.InvalidatePrefix("nodes:")
	return m.inner.Update(id, node)
}

func (m *cachedNodeManager) Delete(id string) error {
	m.cache.InvalidatePrefix("nodes:")
	return m.inner.Delete(id)
}

// --- cachedProfileManager ---

type cachedProfileManager struct {
	inner ProfileManager
	cache *Cache
}

func (m *cachedProfileManager) List() (map[string]*WwProfile, error) {
	const key = "profiles:list"
	if v, ok := m.cache.Get(key); ok {
		return v.(map[string]*WwProfile), nil
	}
	result, err := m.inner.List()
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedProfileManager) Get(id string) (*WwProfile, error) {
	key := fmt.Sprintf("profiles:get:%s", id)
	if v, ok := m.cache.Get(key); ok {
		return v.(*WwProfile), nil
	}
	result, err := m.inner.Get(id)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedProfileManager) Add(id string, profile *WwProfile) error {
	m.cache.InvalidatePrefix("profiles:")
	return m.inner.Add(id, profile)
}

func (m *cachedProfileManager) Update(id string, profile *WwProfile) error {
	m.cache.InvalidatePrefix("profiles:")
	return m.inner.Update(id, profile)
}

func (m *cachedProfileManager) Delete(id string) error {
	m.cache.InvalidatePrefix("profiles:")
	return m.inner.Delete(id)
}

// --- cachedImageManager ---

type cachedImageManager struct {
	inner ImageManager
	cache *Cache
}

func (m *cachedImageManager) List() (map[string]*WwImage, error) {
	const key = "images:list"
	if v, ok := m.cache.Get(key); ok {
		return v.(map[string]*WwImage), nil
	}
	result, err := m.inner.List()
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedImageManager) Get(name string) (*WwImage, error) {
	key := fmt.Sprintf("images:get:%s", name)
	if v, ok := m.cache.Get(key); ok {
		return v.(*WwImage), nil
	}
	result, err := m.inner.Get(name)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedImageManager) Build(name string) error {
	m.cache.InvalidatePrefix("images:")
	return m.inner.Build(name)
}

func (m *cachedImageManager) Import(name string, source string) error {
	m.cache.InvalidatePrefix("images:")
	return m.inner.Import(name, source)
}

func (m *cachedImageManager) Update(name string, newName string) error {
	m.cache.InvalidatePrefix("images:")
	return m.inner.Update(name, newName)
}

func (m *cachedImageManager) Delete(name string) error {
	m.cache.InvalidatePrefix("images:")
	return m.inner.Delete(name)
}

// --- cachedOverlayManager ---

type cachedOverlayManager struct {
	inner OverlayManager
	cache *Cache
}

func (m *cachedOverlayManager) List() (map[string]*WwOverlay, error) {
	const key = "overlays:list"
	if v, ok := m.cache.Get(key); ok {
		return v.(map[string]*WwOverlay), nil
	}
	result, err := m.inner.List()
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedOverlayManager) Get(name string) (*WwOverlay, error) {
	key := fmt.Sprintf("overlays:get:%s", name)
	if v, ok := m.cache.Get(key); ok {
		return v.(*WwOverlay), nil
	}
	result, err := m.inner.Get(name)
	if err != nil {
		return nil, err
	}
	m.cache.Set(key, result)
	return result, nil
}

func (m *cachedOverlayManager) GetFile(name, path, renderNode string) (*OverlayFile, error) {
	// File contents are not cached — they may be rendered per-node.
	return m.inner.GetFile(name, path, renderNode)
}

func (m *cachedOverlayManager) Create(name string) error {
	m.cache.InvalidatePrefix("overlays:")
	return m.inner.Create(name)
}

func (m *cachedOverlayManager) AddFile(name, path, content string) error {
	m.cache.InvalidatePrefix("overlays:")
	return m.inner.AddFile(name, path, content)
}

func (m *cachedOverlayManager) Delete(name string, force bool) error {
	m.cache.InvalidatePrefix("overlays:")
	return m.inner.Delete(name, force)
}

func (m *cachedOverlayManager) DeleteFile(name, path string) error {
	m.cache.InvalidatePrefix("overlays:")
	return m.inner.DeleteFile(name, path)
}
