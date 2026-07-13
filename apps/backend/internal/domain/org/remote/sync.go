package remote

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *Service) GetSyncConfig(ctx context.Context) (types.SyncConfig, error) {
	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return types.SyncConfig{}, err
	}
	return integration.ToSyncConfig(), nil
}

func (s *Service) UpdateSyncConfig(ctx context.Context, cfg types.SyncConfig) error {
	if cfg.FrequencyHours < 1 {
		return domain.Validation("frequencyHours must be at least 1")
	}
	if cfg.DeleteMemberThreshold < 0 {
		return domain.Validation("deleteMemberThreshold must not be negative")
	}
	if cfg.DeleteDepartmentThreshold < 0 {
		return domain.Validation("deleteDepartmentThreshold must not be negative")
	}

	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return err
	}
	integration.ApplySyncConfig(cfg)
	if err := s.d.Store.Org().SetIntegration(ctx, integration); err != nil {
		return err
	}

	companyID := company.CompanyID(ctx)
	if companyID == 0 {
		return fmt.Errorf("org sync config: company context required")
	}
	if err := s.d.Store.TenantBackgroundState().EnsureRow(ctx, companyID); err != nil {
		return err
	}
	if !cfg.Enabled {
		return s.enqueuer.CancelPendingOrgSync(ctx, companyID)
	}
	tbs, err := s.d.Store.TenantBackgroundState().Get(ctx, companyID)
	if err != nil {
		return err
	}
	return s.rescheduleOrgSync(ctx, cfg, tbs, clock.NowUTC(s.d.Cfg.Clock()))
}

func (s *Service) ListSyncLogs(ctx context.Context, page, pageSize int) (types.PageResult[types.SyncLog], error) {
	logs, err := s.d.Store.Org().SyncLogs(ctx)
	if err != nil {
		return types.PageResult[types.SyncLog]{}, err
	}
	items, total, safePage, safeSize := common.Paginate(logs, page, pageSize)
	return types.PageResult[types.SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}, nil
}
