package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type BudgetRepository interface {
	AddGroupConsumed(ctx context.Context, groupID string, amountCNY float64) error
	GetGroupBudget(ctx context.Context, groupID string) (budget, consumed float64, found bool, err error)
	Groups(ctx context.Context) ([]types.BudgetGroup, error)
	SetGroups(ctx context.Context, groups []types.BudgetGroup) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
}
