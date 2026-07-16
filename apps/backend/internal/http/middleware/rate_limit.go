package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/infra/ratelimit"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

// RateLimitTenant applies per-tenant (companyID) token bucket rate limiting.
// Fail-open: if limiter is nil or Redis errors, requests pass through.
func RateLimitTenant(limiter ratelimit.Limiter, rate, burst int, dryRun bool, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			companyID := ctxcompany.ID(r.Context())
			if companyID == 0 {
				// No tenant resolved yet — skip (login/public routes).
				next.ServeHTTP(w, r)
				return
			}
			key := fmt.Sprintf("rl:api:%d", companyID)
			result, err := limiter.AllowTokenBucket(r.Context(), key, rate, burst)
			if err != nil {
				// Fail-open on Redis error.
				logger.Warn("rate_limit: redis error, fail-open", "error", err, "key", key)
				next.ServeHTTP(w, r)
				return
			}
			ratelimit.WriteHeaders(w, result)
			if !result.Allowed {
				if dryRun {
					logger.Warn("rate_limit: would reject (dry-run)", "layer", "tenant", "key", key)
					next.ServeHTTP(w, r)
					return
				}
				ratelimit.WriteRejection(w, result)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitLogin applies per-IP sliding window rate limiting for login endpoints.
// Fail-closed: if Redis is unavailable, falls back to the memory limiter.
func RateLimitLogin(redisLimiter ratelimit.Limiter, memoryLimiter ratelimit.Limiter, max, windowSec int, dryRun bool, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			key := fmt.Sprintf("rl:login:%s", ip)

			var result ratelimit.Result
			var err error

			if redisLimiter != nil {
				result, err = redisLimiter.AllowSlidingWindow(r.Context(), key, max, windowSec)
			}
			if redisLimiter == nil || err != nil {
				// Fail-closed: use memory fallback.
				if err != nil {
					logger.Warn("rate_limit: login redis error, using memory fallback", "error", err)
				}
				result, _ = memoryLimiter.AllowSlidingWindow(r.Context(), key, max, windowSec)
			}

			ratelimit.WriteHeaders(w, result)
			if !result.Allowed {
				if dryRun {
					logger.Warn("rate_limit: login would reject (dry-run)", "ip", ip)
					next.ServeHTTP(w, r)
					return
				}
				ratelimit.WriteRejection(w, result)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitLoginPaths applies per-IP login rate limiting only on the specified paths (POST only).
func RateLimitLoginPaths(paths []string, redisLimiter ratelimit.Limiter, memoryLimiter ratelimit.Limiter, max, windowSec int, dryRun bool, logger *slog.Logger) func(http.Handler) http.Handler {
	inner := RateLimitLogin(redisLimiter, memoryLimiter, max, windowSec, dryRun, logger)
	pathSet := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		pathSet[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		limited := inner(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				if _, ok := pathSet[r.URL.Path]; ok {
					limited.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
