package org

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type service struct {
	cfg         config.Config
	store       store.Store
	factory     datasource.Factory
	modelLimits relay.ModelLimitsEnqueuer
	notifier    notification.Notifier
	delayer     common.Delayer
	cryptoKey   []byte
	logger      *slog.Logger
}

func NewService(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	modelLimits relay.ModelLimitsEnqueuer,
	notifier notification.Notifier,
	delayer common.Delayer,
	logger *slog.Logger,
) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		cfg:         cfg,
		store:       st,
		factory:     factory,
		modelLimits: modelLimits,
		notifier:    notifier,
		delayer:     delayer,
		logger:      logger,
	}
}

func (s *service) recalcRoleMemberCounts(ctx context.Context, roles []types.Role) error {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	for i := range roles {
		roles[i].MemberCount = pkgorg.CountMembersByRole(members, roles[i].Name)
	}
	return nil
}

func (s *service) ListPermissions(ctx context.Context) ([]types.Permission, error) {
	return s.store.Org().Permissions(ctx)
}

func (s *service) GetSyncConfig(ctx context.Context) (types.SyncConfig, error) {
	integration, err := s.store.Org().Integration(ctx)
	if err != nil {
		return types.SyncConfig{}, err
	}
	return integration.ToSyncConfig(), nil
}

func (s *service) UpdateSyncConfig(ctx context.Context, cfg types.SyncConfig) error {
	integration, err := s.store.Org().Integration(ctx)
	if err != nil {
		return err
	}
	integration.ApplySyncConfig(cfg)
	return s.store.Org().SetIntegration(ctx, integration)
}

func (s *service) ListSyncLogs(ctx context.Context, page, pageSize int) (types.PageResult[types.SyncLog], error) {
	logs, err := s.store.Org().SyncLogs(ctx)
	if err != nil {
		return types.PageResult[types.SyncLog]{}, err
	}
	items, total, safePage, safeSize := common.Paginate(logs, page, pageSize)
	return types.PageResult[types.SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}, nil
}
