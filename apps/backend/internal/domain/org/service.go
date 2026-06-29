package org

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/notification"
	"github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/roleutil"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type service struct {
	cfg       config.Config
	store     store.Store
	factory   datasource.Factory
	lifecycle relay.Lifecycle
	notifier  notification.Notifier
	delayer   simulate.Delayer
	cryptoKey []byte
	logger    *slog.Logger
}

func NewService(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	lifecycle relay.Lifecycle,
	notifier notification.Notifier,
	delayer simulate.Delayer,
	logger *slog.Logger,
) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		cfg:       cfg,
		store:     st,
		factory:   factory,
		lifecycle: lifecycle,
		notifier:  notifier,
		delayer:   delayer,
		logger:    logger,
	}
}

func (s *service) recalcRoleMemberCounts(roles []Role) {
	members := s.store.Org().Members()
	for i := range roles {
		roles[i].MemberCount = roleutil.CountMembersByRole(members, roles[i].Name)
	}
}

func (s *service) ListPermissions() []Permission {
	return s.store.Org().Permissions()
}

func (s *service) GetSyncConfig() SyncConfig {
	return s.store.Org().SyncConfig()
}

func (s *service) UpdateSyncConfig(cfg SyncConfig) error {
	return s.store.Org().SetSyncConfig(cfg)
}

func (s *service) ListSyncLogs(page, pageSize int) types.PageResult[SyncLog] {
	logs := s.store.Org().SyncLogs()
	items, total, safePage, safeSize := pkg.Paginate(logs, page, pageSize)
	return types.PageResult[SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}
}
