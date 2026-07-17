package gateway

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

const (
	// DefaultCacheMaxSize is the maximum number of entries in the precheck cache.
	// At ~200 bytes per entry, 10,000 entries ≈ 2MB.
	DefaultCacheMaxSize = 10000
	// DefaultCacheTTL is the fallback expiration for cached entries.
	// Invalidate-on-write keeps most data fresh; TTL is a safety net for multi-instance drift.
	DefaultCacheTTL = 5 * time.Minute
	// negativeCacheTTL is the expiration for negative (key-not-found) entries to prevent
	// cache penetration from invalid key hashes.
	negativeCacheTTL = 30 * time.Second
)

// cacheEntry holds a cached row with metadata for LRU eviction and TTL expiration.
// A nil row indicates a negative cache entry (key not found in PG).
type cacheEntry struct {
	row      *store.PrecheckContextRow
	cachedAt time.Time
}

// PrecheckCache is a process-local LRU cache with TTL for key validity data.
// Budget remain checks bypass this cache and go directly to Redis.
type PrecheckCache struct {
	mu      sync.Mutex
	entries map[string]*list.Element // keyHash → list element
	byKeyID map[uuid.UUID]string     // platformKeyID → keyHash (reverse index)
	order   *list.List               // front = most recently used
	loader  store.GatewayPrecheckRepository
	maxSize int
	ttl     time.Duration
}

type lruEntry struct {
	keyHash string
	entry   cacheEntry
}

// CacheOption configures the PrecheckCache.
type CacheOption func(*PrecheckCache)

// WithMaxSize overrides the default max cache size.
func WithMaxSize(n int) CacheOption {
	return func(c *PrecheckCache) {
		if n > 0 {
			c.maxSize = n
		}
	}
}

// WithTTL overrides the default TTL for cached entries.
func WithTTL(d time.Duration) CacheOption {
	return func(c *PrecheckCache) {
		if d > 0 {
			c.ttl = d
		}
	}
}

func NewPrecheckCache(loader store.GatewayPrecheckRepository, opts ...CacheOption) *PrecheckCache {
	c := &PrecheckCache{
		entries: make(map[string]*list.Element),
		byKeyID: make(map[uuid.UUID]string),
		order:   list.New(),
		loader:  loader,
		maxSize: DefaultCacheMaxSize,
		ttl:     DefaultCacheTTL,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Get returns the cached row, loading from PG on cache miss.
func (c *PrecheckCache) Get(ctx context.Context, keyHash string) (*store.PrecheckContextRow, error) {
	c.mu.Lock()
	if el, ok := c.entries[keyHash]; ok {
		le := el.Value.(*lruEntry)
		ttl := c.ttl
		if le.entry.row == nil {
			ttl = negativeCacheTTL
		}
		if time.Since(le.entry.cachedAt) < ttl {
			c.order.MoveToFront(el)
			row := le.entry.row
			c.mu.Unlock()
			return row, nil
		}
		// Expired — remove and reload.
		c.removeLocked(keyHash, le)
	}
	c.mu.Unlock()

	// Cache miss — load from PG.
	row, err := c.loader.LoadPrecheckContext(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.putLocked(keyHash, cacheEntry{
		row:      row,
		cachedAt: time.Now(),
	})
	c.mu.Unlock()

	return row, nil
}

// putLocked inserts or replaces a cache entry. Caller must hold c.mu.
func (c *PrecheckCache) putLocked(keyHash string, entry cacheEntry) {
	if el, ok := c.entries[keyHash]; ok {
		// Update existing entry.
		le := el.Value.(*lruEntry)
		// Clean old reverse index if row changed.
		if le.entry.row != nil {
			delete(c.byKeyID, le.entry.row.PlatformKeyID)
		}
		le.entry = entry
		c.order.MoveToFront(el)
	} else {
		// New entry — evict if full.
		for c.order.Len() >= c.maxSize {
			c.evictOldestLocked()
		}
		el := c.order.PushFront(&lruEntry{keyHash: keyHash, entry: entry})
		c.entries[keyHash] = el
	}
	// Update reverse index.
	if entry.row != nil {
		c.byKeyID[entry.row.PlatformKeyID] = keyHash
	}
}

// evictOldestLocked removes the least recently used entry. Caller must hold c.mu.
func (c *PrecheckCache) evictOldestLocked() {
	el := c.order.Back()
	if el == nil {
		return
	}
	le := el.Value.(*lruEntry)
	c.removeLocked(le.keyHash, le)
}

// removeLocked removes an entry from all indexes. Caller must hold c.mu.
func (c *PrecheckCache) removeLocked(keyHash string, le *lruEntry) {
	if le.entry.row != nil {
		delete(c.byKeyID, le.entry.row.PlatformKeyID)
	}
	if el, ok := c.entries[keyHash]; ok {
		c.order.Remove(el)
	}
	delete(c.entries, keyHash)
}

// InvalidateKey removes a single key from the cache by key hash.
func (c *PrecheckCache) InvalidateKey(keyHash string) {
	c.mu.Lock()
	if el, ok := c.entries[keyHash]; ok {
		le := el.Value.(*lruEntry)
		c.removeLocked(keyHash, le)
	}
	c.mu.Unlock()
}

// InvalidateByKeyID removes a cached entry by platform key ID.
// This is used by keys service which operates on key IDs, not hashes.
func (c *PrecheckCache) InvalidateByKeyID(platformKeyID uuid.UUID) {
	c.mu.Lock()
	if keyHash, ok := c.byKeyID[platformKeyID]; ok {
		if el, found := c.entries[keyHash]; found {
			le := el.Value.(*lruEntry)
			c.removeLocked(keyHash, le)
		}
	}
	c.mu.Unlock()
}

// InvalidateCompany removes all cached keys belonging to a company.
// Call this when company status changes (freeze/unfreeze).
func (c *PrecheckCache) InvalidateCompany(companyID uuid.UUID) {
	c.mu.Lock()
	var toRemove []string
	for keyHash, el := range c.entries {
		le := el.Value.(*lruEntry)
		if le.entry.row != nil && le.entry.row.CompanyID == companyID {
			toRemove = append(toRemove, keyHash)
		}
	}
	for _, keyHash := range toRemove {
		if el, ok := c.entries[keyHash]; ok {
			le := el.Value.(*lruEntry)
			c.removeLocked(keyHash, le)
		}
	}
	c.mu.Unlock()
}

// InvalidateAll clears all cache entries.
func (c *PrecheckCache) InvalidateAll() {
	c.mu.Lock()
	c.entries = make(map[string]*list.Element)
	c.byKeyID = make(map[uuid.UUID]string)
	c.order.Init()
	c.mu.Unlock()
}

// Verify PrecheckCache satisfies the shared PrecheckCacheInvalidator interface.
var _ types.PrecheckCacheInvalidator = (*PrecheckCache)(nil)
