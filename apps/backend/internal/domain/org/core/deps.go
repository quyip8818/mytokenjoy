package core

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Deps struct {
	Cfg         config.Config
	Store       store.Store
	Factory     datasource.Factory
	ModelLimits newapisync.ModelLimitsLifecycle
	Notifier    notification.Notifier
	Delayer     common.Delayer
	Logger      *slog.Logger
	Grants      grants.Normalizer
	cryptoKey   []byte
}

func NewDeps(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	modelLimits newapisync.ModelLimitsLifecycle,
	notifier notification.Notifier,
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
