package audit

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkg "github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/auditfilter"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSettings() types.AuditSettings
	UpdateSettings(settings types.AuditSettings) (types.AuditSettings, error)
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

func (s *service) UpdateSettings(settings types.AuditSettings) (types.AuditSettings, error) {
	current := s.store.Audit().Settings()
	current.ContentRetentionEnabled = settings.ContentRetentionEnabled
	if err := s.store.Audit().SetSettings(current); err != nil {
		return types.AuditSettings{}, fmt.Errorf("persist audit settings: %w", err)
	}
	return current, nil
}

func (s *service) ListOperations(params types.AuditOperationsQueryParams) types.PageResult[types.OperationLog] {
	items := s.store.Audit().OperationLogs()
	items = auditfilter.FilterByEquals(items, params.Action, func(item types.OperationLog) string {
		return item.Action
	})
	items = auditfilter.FilterByEquals(items, params.OperatorID, func(item types.OperationLog) string {
		return item.OperatorID
	})
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
	items = auditfilter.FilterByEquals(items, params.Model, func(item types.CallLog) string {
		return item.Model
	})
	items = auditfilter.FilterByEquals(items, params.Status, func(item types.CallLog) string {
		return item.Status
	})
	items = auditfilter.FilterByEquals(items, params.CallerID, func(item types.CallLog) string {
		return item.CallerID
	})
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
