package core

import (
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Deps struct {
	Cfg         config.Config
	Store       store.Store
	Factory     datasource.Factory
	ModelLimits relay.ModelLimitsEnqueuer
	Notifier    notification.Notifier
	Delayer     common.Delayer
	Logger      *slog.Logger
	cryptoKey   []byte
}

func NewDeps(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	modelLimits relay.ModelLimitsEnqueuer,
	notifier notification.Notifier,
	delayer common.Delayer,
	logger *slog.Logger,
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
	}
}

func (d *Deps) BudgetPeriod() string {
	if len(d.Cfg.DemoToday) >= 7 {
		return d.Cfg.DemoToday[:7]
	}
	return time.Now().Format("2006-01")
}
