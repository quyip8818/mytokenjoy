package ratelimit

import (
	"context"
	"log/slog"
	"strings"

	pkgrl "github.com/tokenjoy/backend/internal/pkg/ratelimit"
)

// Open creates a Redis-backed Limiter. If Redis is unavailable, returns nil.
// Callers should check for nil and decide on fallback behavior.
func Open(ctx context.Context, redisURL string, logger *slog.Logger) pkgrl.Limiter {
	if strings.TrimSpace(redisURL) == "" {
		if logger != nil {
			logger.Info("ratelimit: disabled (no REDIS_URL)")
		}
		return nil
	}
	limiter, err := NewRedisLimiter(ctx, redisURL, logger)
	if err != nil {
		if logger != nil {
			logger.Warn("ratelimit: redis unavailable, disabled", "error", err)
		}
		return nil
	}
	if logger != nil {
		logger.Info("ratelimit: enabled")
	}
	return limiter
}
