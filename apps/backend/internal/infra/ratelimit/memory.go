package ratelimit

import (
	"context"
	"sync"
	"time"

	pkgrl "github.com/tokenjoy/backend/internal/pkg/ratelimit"
)

// MemoryLimiter is a local in-memory rate limiter used as fallback when Redis
// is unavailable. Only used for login rate limiting (fail-closed).
// Not precise across multiple instances, but still effective per-instance.
type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*memoryBucket
	done    chan struct{}
}

type memoryBucket struct {
	count    int
	windowAt time.Time
}

// NewMemoryLimiter creates a local memory rate limiter.
func NewMemoryLimiter() *MemoryLimiter {
	m := &MemoryLimiter{
		buckets: make(map[string]*memoryBucket),
		done:    make(chan struct{}),
	}
	go m.cleanup()
	return m
}

func (m *MemoryLimiter) AllowTokenBucket(_ context.Context, key string, _ int, burst int) (pkgrl.Result, error) {
	return m.AllowSlidingWindow(context.Background(), key, burst, 1)
}

func (m *MemoryLimiter) AllowSlidingWindow(_ context.Context, key string, maxReq int, windowSec int) (pkgrl.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	window := time.Duration(windowSec) * time.Second

	b, ok := m.buckets[key]
	if !ok || now.Sub(b.windowAt) >= window {
		m.buckets[key] = &memoryBucket{count: 1, windowAt: now}
		return pkgrl.Result{
			Allowed:   true,
			Remaining: int64(maxReq - 1),
			Limit:     int64(maxReq),
			ResetAt:   now.Add(window),
		}, nil
	}

	b.count++
	remaining := maxReq - b.count
	allowed := b.count <= maxReq

	return pkgrl.Result{
		Allowed:   allowed,
		Remaining: int64(max(remaining, 0)),
		Limit:     int64(maxReq),
		ResetAt:   b.windowAt.Add(window),
	}, nil
}

func (m *MemoryLimiter) Close() error {
	close(m.done)
	return nil
}

func (m *MemoryLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for k, b := range m.buckets {
				if now.Sub(b.windowAt) > 10*time.Minute {
					delete(m.buckets, k)
				}
			}
			m.mu.Unlock()
		case <-m.done:
			return
		}
	}
}

var _ pkgrl.Limiter = (*MemoryLimiter)(nil)
