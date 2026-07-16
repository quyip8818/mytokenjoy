package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	return s.store.Budget().OverrunPolicy(ctx)
}

func (s *service) UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		if err := tx.Budget().SetOverrunPolicy(ctx, policy); err != nil {
			return fmt.Errorf("persist overrun policy: %w", err)
		}
		return nil
	})
	if err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	return policy, nil
}
