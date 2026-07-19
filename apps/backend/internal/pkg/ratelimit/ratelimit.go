package ratelimit

import (
	"context"
	"time"
)

// Result holds the outcome of a rate limit check.
type Result struct {
	Allowed   bool
	Remaining int64
	Limit     int64
	ResetAt   time.Time
}

// Limiter is the interface for rate limiting operations.
type Limiter interface {
	// AllowTokenBucket checks a token bucket rate limit.
	// Returns whether the request is allowed and remaining tokens.
	AllowTokenBucket(ctx context.Context, key string, rate int, burst int) (Result, error)

	// AllowSlidingWindow checks a sliding window rate limit.
	// Returns whether the request is allowed within the window.
	AllowSlidingWindow(ctx context.Context, key string, max int, windowSec int) (Result, error)

	// Close releases resources.
	Close() error
}
