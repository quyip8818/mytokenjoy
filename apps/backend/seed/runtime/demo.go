package runtime

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

func ApplyDemo(ctx context.Context, st store.Store, cfg config.Config) error {
	if err := ApplyUsageBuckets(ctx, st, cfg); err != nil {
		return fmt.Errorf("apply usage buckets: %w", err)
	}
	if err := ApplyRechargeOrders(ctx, st); err != nil {
		return fmt.Errorf("apply recharge orders: %w", err)
	}
	if err := ApplyUsageLedger(ctx, st, cfg); err != nil {
		return fmt.Errorf("apply usage ledger: %w", err)
	}
	return nil
}
