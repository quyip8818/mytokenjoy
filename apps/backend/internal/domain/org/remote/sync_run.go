package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *Service) TriggerSync(ctx context.Context) (types.ImportResult, error) {
	var result types.ImportResult
	err := s.runOrgSyncLocked(ctx, true, func(ctx context.Context) error {
		var syncErr error
		result, syncErr = s.syncFromProvider(ctx, types.SyncTypeManual)
		if syncErr != nil {
			return syncErr
		}
		return s.recordSyncSuccess(ctx, time.Now().UTC())
	})
	if err != nil {
		return types.ImportResult{}, err
	}
	return result, nil
}

func (s *Service) RunScheduledSync(ctx context.Context) error {
	cfg, err := s.GetSyncConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	if err := s.ensureScheduledOrgSync(ctx); err != nil {
		return err
	}
	return s.runOrgSyncLocked(ctx, false, func(ctx context.Context) error {
		_, syncErr := s.syncFromProvider(ctx, types.SyncTypeScheduled)
		if syncErr != nil {
			return syncErr
		}
		return s.recordSyncSuccess(ctx, time.Now().UTC())
	})
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
		core.NotifySyncThresholdExceeded(ctx, s.d.Notifier, cfg, detail)
		_ = s.appendSyncLog(ctx, syncType, types.SyncResultFailure, detail)
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}
	if len(diff.RemoveDepartments) > cfg.DeleteDepartmentThreshold {
		detail := fmt.Sprintf("department deletions %d exceed threshold %d", len(diff.RemoveDepartments), cfg.DeleteDepartmentThreshold)
		core.NotifySyncThresholdExceeded(ctx, s.d.Notifier, cfg, detail)
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
