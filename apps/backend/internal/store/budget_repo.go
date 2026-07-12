package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type BudgetRepository interface {
	AcquireBudgetLock(ctx context.Context) error
	OrgNodeBudget() OrgNodeBudgetRepository
	GetGroupBudget(ctx context.Context, groupID string) (budget, consumed float64, found bool, err error)
	Groups(ctx context.Context) ([]types.BudgetGroup, error)
	SetGroups(ctx context.Context, groups []types.BudgetGroup) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
	BudgetApprovals(ctx context.Context) ([]types.BudgetApproval, error)
	SetBudgetApprovals(ctx context.Context, items []types.BudgetApproval) error
	UpdateBudgetApproval(ctx context.Context, id, status string, rejectReason *string, resolvedAt time.Time) error
}
