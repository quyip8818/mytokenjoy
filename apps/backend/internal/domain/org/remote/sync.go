package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *Service) TriggerSync(ctx context.Context) (types.ImportResult, error) {
	var result types.ImportResult
	err := s.runOrgSyncLocked(ctx, true, func(ctx context.Context) error {
		var syncErr error
		result, syncErr = s.syncFromProvider(ctx, types.SyncTypeManual)
		return syncErr
	})
	if err != nil {
		return types.ImportResult{}, err
	}
	return result, nil
}

func (s *Service) RunScheduledSync(ctx context.Context) error {
	due, err := s.dueForScheduledSync(ctx)
	if err != nil || !due {
		return err
	}
	return s.runOrgSyncLocked(ctx, false, func(ctx context.Context) error {
		_, syncErr := s.syncFromProvider(ctx, types.SyncTypeScheduled)
		return syncErr
	})
}

// FanoutScheduledSyncJobs enqueues org_sync for tenants due for scheduled sync.
func (s *Service) FanoutScheduledSyncJobs(ctx context.Context) error {
	return company.ForEachActiveCompany(ctx, s.d.Store.Company(), func(entryCtx context.Context, co store.Company) error {
		due, err := s.dueForScheduledSync(entryCtx)
		if err != nil || !due {
			return err
		}
		return jobs.InsertOrgSync(entryCtx, s.enqueuer, nil, co.ID)
	})
}

func (s *Service) dueForScheduledSync(ctx context.Context) (bool, error) {
	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return false, err
	}
	cfg := integration.ToSyncConfig()
	if !cfg.Enabled {
		return false, nil
	}
	return s.shouldRunScheduledSync(ctx, cfg), nil
}

func (s *Service) runOrgSyncLocked(ctx context.Context, conflictIfBusy bool, fn func(context.Context) error) error {
	release, acquired, err := s.acquireOrgSyncLock(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		if conflictIfBusy {
			return domain.Conflict("org sync already in progress")
		}
		return nil
	}
	defer release()
	return fn(ctx)
}

func (s *Service) shouldRunScheduledSync(ctx context.Context, cfg types.SyncConfig) bool {
	if cfg.FrequencyHours <= 0 {
		return false
	}
	lastRun := s.lastScheduledSyncTime(ctx)
	if lastRun != nil && time.Since(*lastRun) < time.Duration(cfg.FrequencyHours)*time.Hour {
		return false
	}
	if cfg.StartTime == "" {
		return true
	}
	parsed, err := time.Parse("15:04", cfg.StartTime)
	if err != nil {
		s.d.Logger.Warn("invalid sync start time", "start_time", cfg.StartTime, "error", err)
		return false
	}
	now := time.Now()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	return !now.Before(startToday)
}

func (s *Service) lastScheduledSyncTime(ctx context.Context) *time.Time {
	logs, err := s.d.Store.Org().SyncLogs(ctx)
	if err != nil {
		return nil
	}
	for _, entry := range logs {
		if entry.Type != types.SyncTypeScheduled {
			continue
		}
		parsed, err := time.Parse("2006-01-02 15:04", entry.Time)
		if err != nil {
			continue
		}
		return &parsed
	}
	return nil
}

func (s *Service) schedulerHolder() string {
	return fmt.Sprintf("worker-%d", time.Now().UnixNano())
}

func (s *Service) acquireOrgSyncLock(ctx context.Context) (release func(), acquired bool, err error) {
	companyID := company.CompanyID(ctx)
	if companyID == 0 {
		return func() {}, false, fmt.Errorf("org sync: company context required")
	}
	holder := s.schedulerHolder()
	lockName := types.OrgSyncLockName(companyID)
	acquired, err = s.d.Store.SchedulerLock().TryAcquire(ctx, lockName, holder, 15*time.Minute)
	if err != nil || !acquired {
		return func() {}, acquired, err
	}
	return func() {
		_ = s.d.Store.SchedulerLock().Release(ctx, lockName, holder)
	}, true, nil
}

func (s *Service) syncFromProvider(ctx context.Context, syncType string) (types.ImportResult, error) {
	provider, platform, err := s.providerForStored(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}

	remoteDepts, err := provider.ListDepartments(ctx)
	if err != nil {
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}
	remoteMembers, fetchFailures, err := provider.ListMembers(ctx)
	if err != nil {
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	localDeptsTree, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
	if err != nil {
		return types.ImportResult{}, err
	}
	localDepts := pkgorg.FlattenDepartmentTree(localDeptsTree)
	localMembers, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	diff := pkgorg.BuildSyncDiff(localDepts, localMembers, remoteDepts, remoteMembers)

	integration, err := s.d.Store.Org().Integration(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	cfg := integration.ToSyncConfig()
	if len(diff.RemoveMembers) > cfg.DeleteMemberThreshold {
		detail := fmt.Sprintf("member deletions %d exceed threshold %d", len(diff.RemoveMembers), cfg.DeleteMemberThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.d.Notifier, cfg, detail)
		_ = s.appendSyncLog(ctx, syncType, types.SyncResultFailure, detail)
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}
	if len(diff.RemoveDepartments) > cfg.DeleteDepartmentThreshold {
		detail := fmt.Sprintf("department deletions %d exceed threshold %d", len(diff.RemoveDepartments), cfg.DeleteDepartmentThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.d.Notifier, cfg, detail)
		_ = s.appendSyncLog(ctx, syncType, types.SyncResultFailure, detail)
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}

	result, applyErr := s.applySyncDiff(ctx, platform, diff)
	result.Failures = append(result.Failures, fetchFailures...)
	if applyErr != nil {
		_ = s.appendSyncLog(ctx, syncType, types.SyncResultFailure, applyErr.Error())
		return result, applyErr
	}

	syncResult := types.SyncResultSuccess
	if len(result.Failures) > 0 {
		syncResult = types.SyncResultPartial
	}
	detail := fmt.Sprintf(
		"成功 %d 人，%d 部门；失败 %d 人",
		result.SuccessMembers, result.SuccessDepartments, len(result.Failures),
	)
	_ = s.appendSyncLog(ctx, syncType, syncResult, detail)
	return result, nil
}

func (s *Service) appendSyncLog(ctx context.Context, syncType, result, detail string) error {
	logEntry := types.SyncLog{
		ID:     fmt.Sprintf("sync-%d", time.Now().UnixNano()),
		Time:   time.Now().Format("2006-01-02 15:04"),
		Type:   syncType,
		Result: result,
		Detail: detail,
	}
	return s.d.Store.Org().AppendSyncLog(ctx, logEntry)
}

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
	return s.d.Store.Org().SetIntegration(ctx, integration)
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
