package audit

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkg "github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/auditfilter"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSettings() types.AuditSettings
	UpdateSettings(settings types.AuditSettings) types.AuditSettings
	ListOperations(params types.AuditOperationsQueryParams) types.PageResult[types.OperationLog]
	ListCalls(params types.AuditCallsQueryParams) types.PageResult[types.CallLog]
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

func (s *service) GetSettings() types.AuditSettings {
	return s.store.Audit().Settings()
}

func (s *service) UpdateSettings(settings types.AuditSettings) types.AuditSettings {
	current := s.store.Audit().Settings()
	current.ContentRetentionEnabled = settings.ContentRetentionEnabled
	s.store.Audit().SetSettings(current)
	return current
}

func (s *service) ListOperations(params types.AuditOperationsQueryParams) types.PageResult[types.OperationLog] {
	items := s.store.Audit().OperationLogs()
	if params.Action != "" {
		filtered := make([]types.OperationLog, 0)
		for _, item := range items {
			if item.Action == params.Action {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	if params.OperatorID != "" {
		filtered := make([]types.OperationLog, 0)
		for _, item := range items {
			if item.OperatorID == params.OperatorID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	items = auditfilter.FilterByDateRangeCreatedAt(items, params.From, params.To, func(item types.OperationLog) string {
		return item.CreatedAt
	})
	items = auditfilter.FilterByKeyword(items, params.Keyword, []func(types.OperationLog) string{
		func(item types.OperationLog) string { return item.Detail },
		func(item types.OperationLog) string { return item.Target },
		func(item types.OperationLog) string { return item.Operator },
	})
	paged, total, page, pageSize := pkg.Paginate(items, params.Page, params.PageSize)
	return types.PageResult[types.OperationLog]{
		Items: paged, Total: total, Page: page, PageSize: pageSize,
	}
}

func (s *service) ListCalls(params types.AuditCallsQueryParams) types.PageResult[types.CallLog] {
	items := s.store.Audit().CallLogs()
	if params.Model != "" {
		filtered := make([]types.CallLog, 0)
		for _, item := range items {
			if item.Model == params.Model {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	if params.Status != "" {
		filtered := make([]types.CallLog, 0)
		for _, item := range items {
			if item.Status == params.Status {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	if params.CallerID != "" {
		filtered := make([]types.CallLog, 0)
		for _, item := range items {
			if item.CallerID == params.CallerID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	items = auditfilter.FilterByDateRangeCreatedAt(items, params.From, params.To, func(item types.CallLog) string {
		return item.CreatedAt
	})
	items = auditfilter.FilterByKeyword(items, params.Keyword, []func(types.CallLog) string{
		func(item types.CallLog) string { return item.InputPreview },
		func(item types.CallLog) string { return item.OutputPreview },
		func(item types.CallLog) string { return item.Caller },
		func(item types.CallLog) string { return item.Model },
	})
	paged, total, page, pageSize := pkg.Paginate(items, params.Page, params.PageSize)
	return types.PageResult[types.CallLog]{
		Items: paged, Total: total, Page: page, PageSize: pageSize,
	}
}
