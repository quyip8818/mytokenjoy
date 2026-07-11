package org

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/domain/org/structure"
	"github.com/tokenjoy/backend/internal/infra/notification"
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
	modelLimits newapisync.ModelLimitsLifecycle,
	notifier notification.Notifier,
	delayer common.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
) Service {
	deps := core.NewDeps(cfg, st, factory, modelLimits, notifier, delayer, logger, grants)
	return &service{
		LocalService: structure.New(deps),
		Service:      remote.New(deps),
	}
}
