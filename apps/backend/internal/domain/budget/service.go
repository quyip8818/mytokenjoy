package budget

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/clock"
	"github.com/tokenjoy/backend/internal/support/simulate"
)

type Service interface {
	GetTree(ctx context.Context) ([]types.BudgetNode, error)
	GetTreeForPeriod(ctx context.Context, period string) ([]types.BudgetNode, error)
	CopyPeriod(ctx context.Context, toPeriod string) error
	RotatePeriod(ctx context.Context) error
	UpdateSnapshotNode(ctx context.Context, period string, nodeID uuid.UUID, budget float64, reservedPool *float64) error
	UpdateSnapshotMember(ctx context.Context, period string, memberID uuid.UUID, personalBudget float64) error
	UpdateSnapshotProject(ctx context.Context, period string, projectID uuid.UUID, budget float64) error
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
	GetProjectMemberConsumed(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]float64, error)
	MemberSummary(ctx context.Context, memberID uuid.UUID) (types.MemberBudgetSummary, error)
}

// Store is the narrow store surface the budget service needs.
type Store interface {
	Budget() store.BudgetRepository
	BudgetConsumed() store.BudgetConsumedRepository
	BudgetSnapshot() store.BudgetSnapshotRepository
	Org() store.OrgRepository
	Keys() store.KeysRepository
	Ledger() store.LedgerRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	Audit() store.AuditRepository
	TenantBackgroundState() store.TenantBackgroundStateRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg      config.Config
	store    Store
	clk      clock.Clock
	delayer  simulate.Delayer
	logger   *slog.Logger
	enqueuer JobEnqueuer
}

func NewService(cfg config.Config, st Store, delayer simulate.Delayer, enqueuer JobEnqueuer, clk clock.Clock) Service {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	return &service{
		cfg:      cfg,
		store:    st,
		clk:      clock.OrDefault(clk),
		delayer:  delayer,
		logger:   slog.Default(),
		enqueuer: enqueuer,
	}
}
