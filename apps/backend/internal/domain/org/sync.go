package org

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) TriggerSync(ctx context.Context) (types.ImportResult, error) {
	return s.syncFromProvider(ctx, types.SyncTypeManual)
}

func (s *service) RunScheduledSync(ctx context.Context) error {
	cfg, err := s.store.Org().Integration(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	if !s.shouldRunScheduledSync(ctx, cfg.ToSyncConfig()) {
		return nil
	}

	holder := s.schedulerHolder()
	acquired, err := s.store.SchedulerLock().TryAcquire(
		ctx, types.SchedulerLockOrgSync, holder, 15*time.Minute,
	)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() {
		_ = s.store.SchedulerLock().Release(ctx, types.SchedulerLockOrgSync, holder)
	}()

	_, syncErr := s.syncFromProvider(ctx, types.SyncTypeScheduled)
	return syncErr
}

func (s *service) shouldRunScheduledSync(ctx context.Context, cfg types.SyncConfig) bool {
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
		s.logger.Warn("invalid sync start time", "start_time", cfg.StartTime, "error", err)
		return false
	}
	now := time.Now()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	return !now.Before(startToday)
}

func (s *service) lastScheduledSyncTime(ctx context.Context) *time.Time {
	logs, err := s.store.Org().SyncLogs(ctx)
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

func (s *service) schedulerHolder() string {
	return fmt.Sprintf("worker-%d", time.Now().UnixNano())
}

func (s *service) syncFromProvider(ctx context.Context, syncType string) (types.ImportResult, error) {
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

	localDeptsTree, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.ImportResult{}, err
	}
	localDepts := pkgorg.FlattenDepartmentTree(localDeptsTree)
	localMembers, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	diff := buildSyncDiff(localDepts, localMembers, remoteDepts, remoteMembers)

	integration, err := s.store.Org().Integration(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	cfg := integration.ToSyncConfig()
	if len(diff.removeMembers) > cfg.DeleteMemberThreshold {
		detail := fmt.Sprintf("member deletions %d exceed threshold %d", len(diff.removeMembers), cfg.DeleteMemberThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.notifier, cfg, detail)
		_ = s.appendSyncLog(ctx, syncType, types.SyncResultFailure, detail)
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}
	if len(diff.removeDepartment) > cfg.DeleteDepartmentThreshold {
		detail := fmt.Sprintf("department deletions %d exceed threshold %d", len(diff.removeDepartment), cfg.DeleteDepartmentThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.notifier, cfg, detail)
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

func (s *service) appendSyncLog(ctx context.Context, syncType, result, detail string) error {
	logEntry := types.SyncLog{
		ID:     fmt.Sprintf("sync-%d", time.Now().UnixNano()),
		Time:   time.Now().Format("2006-01-02 15:04"),
		Type:   syncType,
		Result: result,
		Detail: detail,
	}
	return s.store.Org().AppendSyncLog(ctx, logEntry)
}
