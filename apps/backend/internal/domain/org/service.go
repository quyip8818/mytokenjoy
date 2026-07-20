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
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
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
	delayer common.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
	enqueuer remote.JobEnqueuer,
) Service {
	deps := core.NewDeps(cfg, st, factory, notifier, delayer, logger, grants)
	return &service{
		LocalService: structure.New(deps),
		Service:      remote.New(deps, enqueuer),
	}
}
