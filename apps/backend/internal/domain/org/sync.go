package org

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/notification"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/store"
)

type syncDiff struct {
	addDepartments    []datasource.RemoteDepartment
	updateDepartments []datasource.RemoteDepartment
	removeDepartment  []types.Department
	addMembers        []datasource.RemoteMember
	updateMembers     []datasource.RemoteMember
	removeMembers     []types.Member
}

func (s *service) TriggerSync(ctx context.Context) (ImportResult, error) {
	return s.syncFromProvider(ctx, types.SyncTypeManual)
}

func (s *service) RunScheduledSync(ctx context.Context) error {
	cfg := s.store.Org().SyncConfig()
	if !cfg.Enabled {
		return nil
	}
	if !s.shouldRunScheduledSync(cfg) {
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

func (s *service) shouldRunScheduledSync(cfg SyncConfig) bool {
	if cfg.FrequencyHours <= 0 {
		return false
	}
	lastRun := s.lastScheduledSyncTime()
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

func (s *service) lastScheduledSyncTime() *time.Time {
	logs := s.store.Org().SyncLogs()
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

func (s *service) syncFromProvider(ctx context.Context, syncType string) (ImportResult, error) {
	provider, platform, err := s.providerForStored()
	if err != nil {
		return ImportResult{}, err
	}

	remoteDepts, err := provider.ListDepartments(ctx)
	if err != nil {
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}
	remoteMembers, fetchFailures, err := provider.ListMembers(ctx)
	if err != nil {
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	localDepts := orgutil.FlattenDepartmentTree(s.store.Org().Departments())
	localMembers := s.store.Org().Members()
	diff := buildSyncDiff(localDepts, localMembers, remoteDepts, remoteMembers)

	cfg := s.store.Org().SyncConfig()
	if len(diff.removeMembers) > cfg.DeleteMemberThreshold {
		detail := fmt.Sprintf("member deletions %d exceed threshold %d", len(diff.removeMembers), cfg.DeleteMemberThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.notifier, cfg, detail)
		_ = s.appendSyncLog(syncType, types.SyncResultFailure, detail)
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}
	if len(diff.removeDepartment) > cfg.DeleteDepartmentThreshold {
		detail := fmt.Sprintf("department deletions %d exceed threshold %d", len(diff.removeDepartment), cfg.DeleteDepartmentThreshold)
		notification.NotifySyncThresholdExceeded(ctx, s.notifier, cfg, detail)
		_ = s.appendSyncLog(syncType, types.SyncResultFailure, detail)
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, detail)
	}

	result, applyErr := s.applySyncDiff(ctx, platform, diff)
	result.Failures = append(result.Failures, fetchFailures...)
	if applyErr != nil {
		_ = s.appendSyncLog(syncType, types.SyncResultFailure, applyErr.Error())
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
	_ = s.appendSyncLog(syncType, syncResult, detail)
	return result, nil
}

func (s *service) appendSyncLog(syncType, result, detail string) error {
	logEntry := types.SyncLog{
		ID:     fmt.Sprintf("sync-%d", time.Now().UnixNano()),
		Time:   time.Now().Format("2006-01-02 15:04"),
		Type:   syncType,
		Result: result,
		Detail: detail,
	}
	return s.store.Org().AppendSyncLog(logEntry)
}

func buildSyncDiff(
	localDepts []types.Department,
	localMembers []types.Member,
	remoteDepts []datasource.RemoteDepartment,
	remoteMembers []datasource.RemoteMember,
) syncDiff {
	remoteDeptMap := make(map[string]datasource.RemoteDepartment, len(remoteDepts))
	for _, dept := range remoteDepts {
		remoteDeptMap[dept.ExternalID] = dept
	}
	remoteMemberMap := make(map[string]datasource.RemoteMember, len(remoteMembers))
	for _, member := range remoteMembers {
		remoteMemberMap[member.ExternalID] = member
	}

	diff := syncDiff{}
	localImportedDepts := make(map[string]types.Department)
	for _, dept := range localDepts {
		if dept.ExternalID == nil || isManualDeptSource(dept.Source) {
			continue
		}
		localImportedDepts[*dept.ExternalID] = dept
		remote, ok := remoteDeptMap[*dept.ExternalID]
		if !ok {
			diff.removeDepartment = append(diff.removeDepartment, dept)
			continue
		}
		if remote.Name != dept.Name {
			diff.updateDepartments = append(diff.updateDepartments, remote)
		}
	}
	for _, remote := range remoteDepts {
		if _, ok := localImportedDepts[remote.ExternalID]; !ok {
			diff.addDepartments = append(diff.addDepartments, remote)
		}
	}

	for _, member := range localMembers {
		if member.ExternalID == nil || isManualMemberSource(member.Source) {
			continue
		}
		remote, ok := remoteMemberMap[*member.ExternalID]
		if !ok {
			diff.removeMembers = append(diff.removeMembers, member)
			continue
		}
		if remote.Name != member.Name || remote.Email != member.Email || remote.Mobile != member.Phone {
			diff.updateMembers = append(diff.updateMembers, remote)
		}
	}
	for _, remote := range remoteMembers {
		found := false
		for _, member := range localMembers {
			if member.ExternalID != nil && *member.ExternalID == remote.ExternalID {
				found = true
				break
			}
		}
		if !found {
			diff.addMembers = append(diff.addMembers, remote)
		}
	}
	return diff
}

func (s *service) applySyncDiff(ctx context.Context, platform types.Platform, diff syncDiff) (ImportResult, error) {
	remoteDepts := append([]datasource.RemoteDepartment{}, diff.addDepartments...)
	remoteDepts = append(remoteDepts, diff.updateDepartments...)
	remoteMembers := append([]datasource.RemoteMember{}, diff.addMembers...)
	remoteMembers = append(remoteMembers, diff.updateMembers...)

	result := ImportResult{}
	if len(remoteDepts) > 0 || len(remoteMembers) > 0 {
		importResult, err := s.importRemoteSnapshot(ctx, platform, remoteDepts, remoteMembers)
		if err != nil {
			return result, err
		}
		result.SuccessDepartments += importResult.SuccessDepartments
		result.SuccessMembers += importResult.SuccessMembers
	}

	if len(diff.removeMembers) == 0 && len(diff.removeDepartment) == 0 {
		return result, nil
	}

	err := s.store.WithTx(ctx, func(st store.Store) error {
		members := st.Org().Members()
		for _, removed := range diff.removeMembers {
			for i := range members {
				if members[i].ID != removed.ID {
					continue
				}
				members[i].Status = "inactive"
				result.SuccessMembers++
			}
		}

		state := &ProvisionState{
			Departments: st.Org().Departments(),
			BudgetTree:  st.Budget().Tree(),
			Rules:       st.Models().RoutingRules(),
			Models:      st.Models().Models(),
		}
		for _, removed := range diff.removeDepartment {
			if err := DeprovisionDepartment(state, removed.ID); err != nil {
				return err
			}
			result.SuccessDepartments++
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(state.Rules); err != nil {
			return err
		}
		return st.Org().SetMembers(members)
	})
	return result, err
}

func (s *service) importRemoteSnapshot(
	ctx context.Context,
	platform types.Platform,
	remoteDepts []datasource.RemoteDepartment,
	remoteMembers []datasource.RemoteMember,
) (ImportResult, error) {
	provider := &fixedProvider{departments: remoteDepts, members: remoteMembers}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}
