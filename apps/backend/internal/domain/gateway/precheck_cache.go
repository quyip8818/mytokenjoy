package gateway

import (
	"context"
	"sync"

	"github.com/tokenjoy/backend/internal/store"
)

// PrecheckCache is a process-local invalidate-on-write cache for key validity data.
// Budget remain checks bypass this cache and go directly to Redis.
type PrecheckCache struct {
	mu      sync.RWMutex
	entries map[string]*store.PrecheckContextRow // keyHash → row
	byKeyID map[string]string                    // platformKeyID → keyHash (reverse index)
	loader  store.GatewayPrecheckRepository
}

func NewPrecheckCache(loader store.GatewayPrecheckRepository) *PrecheckCache {
	return &PrecheckCache{
		entries: make(map[string]*store.PrecheckContextRow),
		byKeyID: make(map[string]string),
		loader:  loader,
	}
}

// Get returns the cached row, loading from PG on cache miss.
func (c *PrecheckCache) Get(ctx context.Context, keyHash string) (*store.PrecheckContextRow, error) {
	c.mu.RLock()
	row, ok := c.entries[keyHash]
	c.mu.RUnlock()
	if ok {
		return row, nil
	}
	// Cache miss — load from PG.
	row, err := c.loader.LoadPrecheckContext(ctx, keyHash)
	if err != nil {
		return nil, err
	}
	if row != nil {
		c.mu.Lock()
		c.entries[keyHash] = row
		c.byKeyID[row.PlatformKeyID] = keyHash
		c.mu.Unlock()
	}
	return row, nil
}

// InvalidateKey removes a single key from the cache by key hash.
func (c *PrecheckCache) InvalidateKey(keyHash string) {
	c.mu.Lock()
	if row, ok := c.entries[keyHash]; ok && row != nil {
		delete(c.byKeyID, row.PlatformKeyID)
	}
	delete(c.entries, keyHash)
	c.mu.Unlock()
}

// InvalidateByKeyID removes a cached entry by platform key ID.
// This is used by keys service which operates on key IDs, not hashes.
func (c *PrecheckCache) InvalidateByKeyID(platformKeyID string) {
	c.mu.Lock()
	if keyHash, ok := c.byKeyID[platformKeyID]; ok {
		delete(c.entries, keyHash)
		delete(c.byKeyID, platformKeyID)
	}
	c.mu.Unlock()
}

// InvalidateCompany removes all cached keys belonging to a company.
// Call this when company status changes (freeze/unfreeze).
func (c *PrecheckCache) InvalidateCompany(companyID int64) {
	c.mu.Lock()
	for k, row := range c.entries {
		if row != nil && row.CompanyID == companyID {
			delete(c.byKeyID, row.PlatformKeyID)
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// InvalidateAll clears all cache entries.
func (c *PrecheckCache) InvalidateAll() {
	c.mu.Lock()
	c.entries = make(map[string]*store.PrecheckContextRow)
	c.byKeyID = make(map[string]string)
	c.mu.Unlock()
}
