package budget

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetTree(ctx context.Context) ([]types.BudgetNode, error)
	UpdateNode(ctx context.Context, id uuid.UUID, budget float64, reservedPool *float64) (types.BudgetNode, error)
	ListMemberBudgets(ctx context.Context, deptID uuid.UUID) ([]types.MemberBudget, error)
	UpdateMemberBudget(ctx context.Context, memberID uuid.UUID, personalBudget float64) (types.MemberBudget, error)
	ApplyAverageBudget(ctx context.Context, deptID uuid.UUID, personalBudget float64, recursive bool) error
	ListProjects(ctx context.Context) ([]types.Project, error)
	CreateProject(ctx context.Context, project types.Project) (types.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, patch types.UpdateProjectInput) (types.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error)
	ListAlerts(ctx context.Context) ([]types.AlertRule, error)
	CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error)
	UpdateAlert(ctx context.Context, id uuid.UUID, patch types.AlertRule) (types.AlertRule, error)
	DeleteAlert(ctx context.Context, id uuid.UUID) error
	ListApprovals(ctx context.Context) ([]types.BudgetApproval, error)
	ResolveApproval(ctx context.Context, id uuid.UUID, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error)
	GetProjectMemberConsumed(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]float64, error)
}

// Store is the narrow store surface the budget service needs.
type Store interface {
	Budget() store.BudgetRepository
	BudgetConsumed() store.BudgetConsumedRepository
	Org() store.OrgRepository
	Keys() store.KeysRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg      config.Config
	store    Store
	delayer  common.Delayer
	logger   *slog.Logger
	enqueuer JobEnqueuer
}

func NewService(cfg config.Config, st Store, delayer common.Delayer, enqueuer JobEnqueuer) Service {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	return &service{
		cfg:      cfg,
		store:    st,
		delayer:  delayer,
		logger:   slog.Default(),
		enqueuer: enqueuer,
	}
}
