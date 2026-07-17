package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type BudgetRepository interface {
	AcquireBudgetLock(ctx context.Context) error
	OrgNodeBudget() OrgNodeBudgetRepository
	GetProjectBudget(ctx context.Context, projectID uuid.UUID) (budget, consumed float64, found bool, err error)
	GetProjectMemberBudget(ctx context.Context, projectID, memberID uuid.UUID) (float64, bool, error)
	Projects(ctx context.Context) ([]types.Project, error)
	SetProjects(ctx context.Context, projects []types.Project) error
	OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error
	AlertRules(ctx context.Context) ([]types.AlertRule, error)
	SetAlertRules(ctx context.Context, rules []types.AlertRule) error
	BudgetApprovals(ctx context.Context) ([]types.BudgetApproval, error)
	SetBudgetApprovals(ctx context.Context, items []types.BudgetApproval) error
	UpdateBudgetApproval(ctx context.Context, id, status string, rejectReason *string, resolvedAt time.Time) error
}
