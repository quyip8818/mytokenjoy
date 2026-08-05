package runtime

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

func ApplyDemo(ctx context.Context, st store.Store, cfg config.Config, clk clock.Clock) error {
	if err := ApplyUsageBuckets(ctx, st, cfg, clk); err != nil {
		return fmt.Errorf("apply usage buckets: %w", err)
	}
	if err := ApplyRechargeOrders(ctx, st); err != nil {
		return fmt.Errorf("apply recharge orders: %w", err)
	}
	if err := ApplyUsageLedger(ctx, st, cfg); err != nil {
		return fmt.Errorf("apply usage ledger: %w", err)
	}
	if err := ApplyNotifications(ctx, st); err != nil {
		return fmt.Errorf("apply notifications: %w", err)
	}
	return nil
}
