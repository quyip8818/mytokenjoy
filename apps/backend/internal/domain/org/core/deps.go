package core

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

// Store is the narrow store surface the org domain needs.
type Store interface {
	Org() store.OrgRepository
	Company() store.CompanyRepository
	SchedulerLock() store.SchedulerLockRepository
	TenantBackgroundState() store.TenantBackgroundStateRepository
	RiverJob() store.RiverJobRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type Deps struct {
	Cfg         config.Config
	Store       Store
	Factory     datasource.Factory
	ModelLimits newapisync.ModelLimitsLifecycle
	Notifier    types.Notifier
	Delayer     common.Delayer
	Logger      *slog.Logger
	Grants      grants.Normalizer
	cryptoKey   []byte
}

func NewDeps(
	cfg config.Config,
	st Store,
	factory datasource.Factory,
	modelLimits newapisync.ModelLimitsLifecycle,
	notifier types.Notifier,
	delayer common.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
) *Deps {
	if logger == nil {
		logger = slog.Default()
	}
	return &Deps{
		Cfg:         cfg,
		Store:       st,
		Factory:     factory,
		ModelLimits: modelLimits,
		Notifier:    notifier,
		Delayer:     delayer,
		Logger:      logger,
		Grants:      grants,
	}
}

func (d *Deps) BudgetPeriod() string {
	return pkgbudget.PeriodMonthly
}
