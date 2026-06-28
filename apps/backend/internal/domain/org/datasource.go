package org

import (
	"context"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
)

func (s *service) GetDataSourceStatus() DataSourceStatus {
	return s.store.Org().DataSourceStatus()
}

func (s *service) TestDataSource(ctx context.Context) (DataSourceTestResult, error) {
	if err := s.delayer.Wait(ctx, time.Second); err != nil {
		return DataSourceTestResult{}, err
	}
	return DataSourceTestResult{Success: true}, nil
}

func (s *service) UpdateDataSource() DataSourceStatus {
	platform := PlatformFeishu
	status := s.store.Org().DataSourceStatus()
	status.Connected = true
	status.Platform = &platform
	_ = s.store.Org().SetDataSourceStatus(status)
	return status
}

func (s *service) SearchDataSource(keyword string) DataSourceSearchResult {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return DataSourceSearchResult{}
	}

	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	for _, member := range members {
		if strings.Contains(member.Name, trimmed) ||
			strings.Contains(member.Phone, trimmed) ||
			strings.Contains(member.Email, trimmed) {
			department := member.DepartmentName
			if path := orgutil.GetDeptPath(departments, member.DepartmentID); path != nil {
				department = *path
			}
			return DataSourceSearchResult{
				Name: member.Name, Department: department, MappingOK: true,
			}
		}
	}
	return DataSourceSearchResult{}
}

func (s *service) ImportDataSource(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 2*time.Second); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 120, SuccessDepartments: 5,
		Failures: s.store.Org().ImportFailures(),
	}, nil
}

func (s *service) RetryImport(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 1, SuccessDepartments: 0, Failures: []ImportFailure{},
	}, nil
}

func (s *service) GetSyncConfig() SyncConfig {
	return s.store.Org().SyncConfig()
}

func (s *service) UpdateSyncConfig(cfg SyncConfig) {
	_ = s.store.Org().SetSyncConfig(cfg)
}

func (s *service) TriggerSync(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 1500*time.Millisecond); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 3, SuccessDepartments: 0, Failures: []ImportFailure{},
	}, nil
}

func (s *service) ListSyncLogs(page, pageSize int) types.PageResult[SyncLog] {
	logs := s.store.Org().SyncLogs()
	items, total, safePage, safeSize := pkg.Paginate(logs, page, pageSize)
	return types.PageResult[SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}
}
