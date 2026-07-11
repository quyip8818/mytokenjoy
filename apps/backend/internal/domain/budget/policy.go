package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	return s.store.Budget().OverrunPolicy(ctx)
}

func (s *service) UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	if err := s.store.Budget().SetOverrunPolicy(ctx, policy); err != nil {
		return types.OverrunPolicyConfig{}, fmt.Errorf("persist overrun policy: %w", err)
	}
	return policy, nil
}
