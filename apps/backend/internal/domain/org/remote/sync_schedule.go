package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *Service) ensureScheduledOrgSync(ctx context.Context) error {
	companyID := company.CompanyID(ctx)
	if companyID == uuid.Nil {
		return fmt.Errorf("org sync: company context required")
	}
	cfg, err := s.GetSyncConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	tbs, err := s.d.Store.TenantBackgroundState().Get(ctx, companyID)
	if err != nil {
		return err
	}
	now := clock.NowUTC(s.d.Cfg.Clock())
	if tbs != nil && tbs.NextOrgSyncAt != nil && tbs.NextOrgSyncAt.After(now) {
		hasPending, err := s.d.Store.RiverJob().HasActiveOrgSync(ctx, companyID)
		if err != nil {
			return err
		}
		if hasPending {
			return nil
		}
	}
	hasPending, err := s.d.Store.RiverJob().HasActiveOrgSync(ctx, companyID)
	if err != nil {
		return err
	}
	if hasPending {
		return nil
	}
	return s.rescheduleOrgSync(ctx, cfg, tbs, now)
}

func (s *Service) recordSyncSuccess(ctx context.Context, syncedAt time.Time) error {
	companyID := company.CompanyID(ctx)
	cfg, err := s.GetSyncConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	next := ComputeNextOrgSync(cfg, &syncedAt, s.d.Cfg.Clock())
	if err := s.d.Store.TenantBackgroundState().UpsertOrgSchedule(ctx, companyID, next, &syncedAt); err != nil {
		return err
	}
	if err := s.enqueuer.CancelPendingOrgSync(ctx, companyID); err != nil {
		return err
	}
	return s.enqueuer.InsertOrgSync(ctx, companyID, &next)
}

func (s *Service) rescheduleOrgSync(ctx context.Context, cfg types.SyncConfig, tbs *store.TenantBackgroundState, now time.Time) error {
	companyID := company.CompanyID(ctx)
	if err := s.d.Store.TenantBackgroundState().EnsureRow(ctx, companyID); err != nil {
		return err
	}
	var last *time.Time
	if tbs != nil {
		last = tbs.LastOrgSyncAt
	}
	next := ComputeNextOrgSync(cfg, last, s.d.Cfg.Clock())
	if next.Before(now) {
		next = now
	}
	if err := s.d.Store.TenantBackgroundState().UpsertOrgSchedule(ctx, companyID, next, last); err != nil {
		return err
	}
	if err := s.enqueuer.CancelPendingOrgSync(ctx, companyID); err != nil {
		return err
	}
	return s.enqueuer.InsertOrgSync(ctx, companyID, &next)
}
