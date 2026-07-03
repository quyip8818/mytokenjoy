package authz

import (
	"sync"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type cacheKey struct {
	companyID int64
	memberID  string
	revision  int64
}

type cacheValue struct {
	member      types.Member
	permissions []string
	readOnly    bool
}

type LRUCache struct {
	mu      sync.Mutex
	maxSize int
	order   []cacheKey
	data    map[cacheKey]cacheValue
}

func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 4096
	}
	return &LRUCache{
		maxSize: maxSize,
		data:    make(map[cacheKey]cacheValue),
	}
}

func (c *LRUCache) Get(companyID int64, memberID string, revision int64) (types.Member, []string, bool, bool) {
	key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.data[key]
	if !ok {
		return types.Member{}, nil, false, false
	}
	c.touch(key)
	return value.member, append([]string(nil), value.permissions...), value.readOnly, true
}

func (c *LRUCache) Put(companyID int64, memberID string, revision int64, member types.Member, permissions []string, readOnly bool) {
	key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.data[key]; !ok && len(c.data) >= c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.data, oldest)
	}
	c.data[key] = cacheValue{
		member:      member,
		permissions: append([]string(nil), permissions...),
		readOnly:    readOnly,
	}
	c.touch(key)
}

func (c *LRUCache) touch(key cacheKey) {
	for i, existing := range c.order {
		if existing == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, key)
}
