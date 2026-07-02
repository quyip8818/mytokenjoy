package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type BudgetRepository interface {
	Tree(ctx context.Context) ([]types.BudgetNode, error)
	SetTree(ctx context.Context, tree []types.BudgetNode) error
	AddGroupConsumed(ctx context.Context, groupID string, amountCNY float64) error
	RollupDepartmentConsumed(ctx context.Context, departmentID string, amountCNY float64) error
	GetDepartmentBudget(ctx context.Context, departmentID string) (budget, consumed float64, found bool, err error)
	GetGroupBudget(ctx context.Context, groupID string) (budget, consumed float64, found bool, err error)
	Groups(ctx context.Context) ([]types.BudgetGroup, error)
	SetGroups(ctx context.Context, groups []types.BudgetGroup) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
}
