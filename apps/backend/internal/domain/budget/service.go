package budget

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func generateBudgetID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%d-%x", prefix, time.Now().UnixMilli(), b)
}

type Service interface {
	GetTree(ctx context.Context) ([]types.BudgetNode, error)
	UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error)
	ListMemberBudgets(ctx context.Context, deptID string) ([]types.MemberBudget, error)
	UpdateMemberBudget(ctx context.Context, memberID string, personalBudget float64) (types.MemberBudget, error)
	ApplyAverageBudget(ctx context.Context, deptID string, personalBudget float64, recursive bool) error
	ListProjects(ctx context.Context) ([]types.Project, error)
	CreateProject(ctx context.Context, project types.Project) (types.Project, error)
	UpdateProject(ctx context.Context, id string, patch types.UpdateProjectInput) (types.Project, error)
	DeleteProject(ctx context.Context, id string) error
	GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error)
	ListAlerts(ctx context.Context) ([]types.AlertRule, error)
	CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error)
	UpdateAlert(ctx context.Context, id string, patch types.AlertRule) (types.AlertRule, error)
	DeleteAlert(ctx context.Context, id string) error
	ListApprovals(ctx context.Context) ([]types.BudgetApproval, error)
	ResolveApproval(ctx context.Context, id string, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error)
	GetProjectMemberConsumed(ctx context.Context, groupID string) (map[string]float64, error)
}

type service struct {
	cfg      config.Config
	store    store.Store
	delayer  common.Delayer
	logger   *slog.Logger
	enqueuer JobEnqueuer
}

func NewService(cfg config.Config, st store.Store, delayer common.Delayer, enqueuer JobEnqueuer) Service {
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
