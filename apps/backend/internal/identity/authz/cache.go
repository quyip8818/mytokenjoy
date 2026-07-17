package authz

import (
	"container/list"
	"sync"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type cacheKey struct {
	companyID uuid.UUID
	memberID  uuid.UUID
	revision  int64
}

type cacheValue struct {
	member      types.Member
	permissions []string
	readOnly    bool
}

type lruEntry struct {
	key   cacheKey
	value cacheValue
}

// LRUCache is a thread-safe O(1) LRU cache backed by a doubly-linked list + map.
type LRUCache struct {
	mu      sync.Mutex
	maxSize int
	ll      *list.List
	items   map[cacheKey]*list.Element
}

func NewLRUCache(maxSize int) *LRUCache {
	if maxSize <= 0 {
		maxSize = 4096
	}
	return &LRUCache{
		maxSize: maxSize,
		ll:      list.New(),
		items:   make(map[cacheKey]*list.Element, maxSize),
	}
}

func (c *LRUCache) Get(companyID uuid.UUID, memberID uuid.UUID, revision int64) (types.Member, []string, bool, bool) {
	key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.items[key]
	if !ok {
		return types.Member{}, nil, false, false
	}
	c.ll.MoveToFront(elem)
	entry := elem.Value.(*lruEntry)
	return entry.value.member, append([]string(nil), entry.value.permissions...), entry.value.readOnly, true
}

func (c *LRUCache) Put(companyID uuid.UUID, memberID uuid.UUID, revision int64, member types.Member, permissions []string, readOnly bool) {
	key := cacheKey{companyID: companyID, memberID: memberID, revision: revision}
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.ll.MoveToFront(elem)
		elem.Value.(*lruEntry).value = cacheValue{
			member:      member,
			permissions: append([]string(nil), permissions...),
			readOnly:    readOnly,
		}
		return
	}
	if c.ll.Len() >= c.maxSize {
		oldest := c.ll.Back()
		if oldest != nil {
			c.ll.Remove(oldest)
			delete(c.items, oldest.Value.(*lruEntry).key)
		}
	}
	entry := &lruEntry{
		key: key,
		value: cacheValue{
			member:      member,
			permissions: append([]string(nil), permissions...),
			readOnly:    readOnly,
		},
	}
	c.items[key] = c.ll.PushFront(entry)
}
