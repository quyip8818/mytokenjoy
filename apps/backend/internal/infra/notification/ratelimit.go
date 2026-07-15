package notification

import (
	"sync"
	"time"
)

// RateLimiter provides per-user per-channel rate limiting using a sliding window counter.
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string]*rateBucket
	limit   int           // max events per window
	window  time.Duration // window duration
}

type rateBucket struct {
	count    int
	windowAt time.Time
}

// NewRateLimiter creates a rate limiter.
// Default: 5 SMS per hour, 20 emails per hour per user.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		windows: make(map[string]*rateBucket),
		limit:   limit,
		window:  window,
	}
}

// Allow checks if the given key (e.g. "user:channel") is within rate limits.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, ok := rl.windows[key]
	if !ok || now.Sub(bucket.windowAt) >= rl.window {
		// New or expired window — reset
		rl.windows[key] = &rateBucket{count: 1, windowAt: now}
		// Opportunistic cleanup: prune expired entries every 1000 calls
		if len(rl.windows) > 1000 {
			rl.pruneExpiredLocked(now)
		}
		return true
	}

	if bucket.count >= rl.limit {
		return false
	}
	bucket.count++
	return true
}

// pruneExpiredLocked removes entries whose window has expired. Must be called with mu held.
func (rl *RateLimiter) pruneExpiredLocked(now time.Time) {
	for k, b := range rl.windows {
		if now.Sub(b.windowAt) >= rl.window {
			delete(rl.windows, k)
		}
	}
}

// Reset clears all rate limit state (for testing).
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.windows = make(map[string]*rateBucket)
}

// DefaultSMSRateLimiter returns a limiter of 5 SMS per user per hour.
func DefaultSMSRateLimiter() *RateLimiter {
	return NewRateLimiter(5, time.Hour)
}

// DefaultEmailRateLimiter returns a limiter of 20 emails per user per hour.
func DefaultEmailRateLimiter() *RateLimiter {
	return NewRateLimiter(20, time.Hour)
}
