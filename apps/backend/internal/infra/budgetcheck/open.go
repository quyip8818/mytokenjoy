package budgetcheck

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
)

// Open returns the GatewayBudgetCheck store for cfg. Empty REDIS_URL or an
// unreachable Redis yields Noop so boot and the Gateway never depend on Redis.
func Open(ctx context.Context, cfg config.Config, logger *slog.Logger) Store {
	if !cfg.GatewayBudgetCheckEnabled() {
		return Noop{}
	}
	store, err := newRedisStore(ctx, cfg.RedisURL, cfg.GatewayBudgetCheckTTL())
	if err != nil {
		if logger != nil {
			logger.Warn("gateway budget check disabled: redis unavailable", "error", err)
		}
		return Noop{}
	}
	if logger != nil {
		logger.Info("gateway budget check enabled")
	}
	return store
}
