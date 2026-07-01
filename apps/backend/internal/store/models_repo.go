package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type ModelsRepository interface {
	Models(ctx context.Context) ([]types.ModelInfo, error)
	SetModels(ctx context.Context, models []types.ModelInfo) error
	RoutingRules(ctx context.Context) ([]types.RoutingRule, error)
	SetRoutingRules(ctx context.Context, rules []types.RoutingRule) error
}
