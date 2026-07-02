package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type BudgetRepository interface {
	Tree(ctx context.Context) ([]types.BudgetNode, error)
	SetTree(ctx context.Context, tree []types.BudgetNode) error
	Groups(ctx context.Context) ([]types.BudgetGroup, error)
	SetGroups(ctx context.Context, groups []types.BudgetGroup) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
}
