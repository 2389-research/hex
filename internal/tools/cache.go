// ABOUTME: LRU cache for tool results
// ABOUTME: Caches read-only tool operations to improve performance

package tools

import (
	"container/list"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached tool result
type CacheEntry struct {
	Result    *Result
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ResultCache is an LRU cache for tool results
type ResultCache struct {
	mu        sync.RWMutex
	capacity  int
	ttl       time.Duration
	cache     map[string]*list.Element
	evictList *list.List
	hits      int64
	misses    int64
	evictions int64
}

type cacheItem struct {
	key   string
	entry *CacheEntry
}

// NewResultCache creates a new LRU cache with the given capacity and TTL
func NewResultCache(capacity int, ttl time.Duration) *ResultCache {
	return &ResultCache{
		capacity:  capacity,
		ttl:       ttl,
		cache:     make(map[string]*list.Element, capacity),
		evictList: list.New(),
	}
}

// generateKey creates a cache key from tool name and parameters
func (c *ResultCache) generateKey(toolName string, params map[string]interface{}) string {
	// Ensure params is not nil
	if params == nil {
		params = make(map[string]interface{})
	}

	// Create deterministic key from tool name + params
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// If marshaling fails, create a key from tool name only
		// This prevents cache hits for unmarshalable params but doesn't break execution
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s:ERROR_UNMARSHALLABLE_%d", toolName, time.Now().UnixNano())))
		return fmt.Sprintf("%x", hash)
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", toolName, paramsJSON)))
	return fmt.Sprintf("%x", hash)
}

// Get retrieves a cached result if available and not expired
func (c *ResultCache) Get(toolName string, params map[string]interface{}) (*Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(toolName, params)

	elem, ok := c.cache[key]
	if !ok {
		c.misses++
		return nil, false
	}

	item := elem.Value.(*cacheItem)

	// Check expiration
	if item.entry.IsExpired() {
		c.evictList.Remove(elem)
		delete(c.cache, key)
		c.misses++
		c.evictions++
		return nil, false
	}

	// Move to front (most recently used)
	c.evictList.MoveToFront(elem)
	c.hits++

	// Deep copy to prevent aliasing and data corruption
	cached := item.entry.Result
	resultCopy := &Result{
		ToolName: cached.ToolName,
		Success:  cached.Success,
		Output:   cached.Output,
		Error:    cached.Error,
		Metadata: make(map[string]interface{}),
	}
	for k, v := range cached.Metadata {
		resultCopy.Metadata[k] = v
	}

	return resultCopy, true
}

// Set stores a result in the cache
func (c *ResultCache) Set(toolName string, params map[string]interface{}, result *Result) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(toolName, params)

	// Deep copy the result to prevent external modifications
	resultCopy := &Result{
		ToolName: result.ToolName,
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		Metadata: make(map[string]interface{}),
	}
	for k, v := range result.Metadata {
		resultCopy.Metadata[k] = v
	}

	// Check if already exists
	if elem, ok := c.cache[key]; ok {
		// Update existing entry
		item := elem.Value.(*cacheItem)
		item.entry = &CacheEntry{
			Result:    resultCopy,
			ExpiresAt: time.Now().Add(c.ttl),
		}
		c.evictList.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &CacheEntry{
		Result:    resultCopy,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	item := &cacheItem{
		key:   key,
		entry: entry,
	}

	elem := c.evictList.PushFront(item)
	c.cache[key] = elem

	// Evict oldest if over capacity
	if c.evictList.Len() > c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			oldestItem := oldest.Value.(*cacheItem)
			delete(c.cache, oldestItem.key)
			c.evictions++
		}
	}
}

// Clear removes all entries from the cache
func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element, c.capacity)
	c.evictList.Init()
}

// Stats returns cache statistics
func (c *ResultCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.evictList.Len(),
		Capacity:  c.capacity,
		HitRate:   hitRate,
	}
}

// CacheStats contains cache statistics
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	Size      int
	Capacity  int
	HitRate   float64
}

// IsCacheable determines if a tool's results should be cached
func IsCacheable(toolName string) bool {
	// Only cache read-only operations
	cacheableTools := map[string]bool{
		"read_file": true,
		"grep":      true,
		"glob":      true,
	}

	return cacheableTools[toolName]
}
