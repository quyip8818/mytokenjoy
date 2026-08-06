package org

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/domain/org/structure"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/clock"
	"github.com/tokenjoy/backend/internal/support/simulate"
)

type service struct {
	*structure.LocalService
	*remote.Service
}

func NewService(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	notifier types.Notifier,
	sender core.DirectSender,
	delayer simulate.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
	enqueuer remote.JobEnqueuer,
	clk clock.Clock,
) Service {
	deps := core.NewDeps(cfg, st, factory, notifier, sender, delayer, logger, grants, clk)
	return &service{
		LocalService: structure.New(deps),
		Service:      remote.New(deps, enqueuer),
	}
}
