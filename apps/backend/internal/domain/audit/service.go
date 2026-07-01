package audit

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSettings(ctx context.Context) (types.AuditSettings, error)
	UpdateSettings(ctx context.Context, settings types.AuditSettings) (types.AuditSettings, error)
	ListOperations(ctx context.Context, params types.AuditOperationsQueryParams) (types.PageResult[types.OperationLog], error)
	ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error)
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

func (s *service) GetSettings(ctx context.Context) (types.AuditSettings, error) {
	return s.store.Audit().Settings(ctx)
}

func (s *service) UpdateSettings(ctx context.Context, settings types.AuditSettings) (types.AuditSettings, error) {
	current, err := s.store.Audit().Settings(ctx)
	if err != nil {
		return types.AuditSettings{}, err
	}
	current.ContentRetentionEnabled = settings.ContentRetentionEnabled
	if err := s.store.Audit().SetSettings(ctx, current); err != nil {
		return types.AuditSettings{}, fmt.Errorf("persist audit settings: %w", err)
	}
	return current, nil
}

func (s *service) ListOperations(ctx context.Context, params types.AuditOperationsQueryParams) (types.PageResult[types.OperationLog], error) {
	items, err := s.store.Audit().OperationLogs(ctx)
	if err != nil {
		return types.PageResult[types.OperationLog]{}, err
	}
	items = common.FilterByEquals(items, params.Action, func(item types.OperationLog) string {
		return item.Action
	})
	items = common.FilterByEquals(items, params.OperatorID, func(item types.OperationLog) string {
		return item.OperatorID
	})
	items = common.FilterByDateRangeCreatedAt(items, params.From, params.To, func(item types.OperationLog) string {
		return item.CreatedAt
	})
	items = common.FilterByKeyword(items, params.Keyword, []func(types.OperationLog) string{
		func(item types.OperationLog) string { return item.Detail },
		func(item types.OperationLog) string { return item.Target },
		func(item types.OperationLog) string { return item.Operator },
	})
	paged, total, page, pageSize := common.Paginate(items, params.Page, params.PageSize)
	return types.PageResult[types.OperationLog]{
		Items: paged, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func (s *service) ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error) {
	items, err := s.store.Audit().CallLogs(ctx)
	if err != nil {
		return types.PageResult[types.CallLog]{}, err
	}
	items = common.FilterByEquals(items, params.Model, func(item types.CallLog) string {
		return item.Model
	})
	items = common.FilterByEquals(items, params.Status, func(item types.CallLog) string {
		return item.Status
	})
	items = common.FilterByEquals(items, params.CallerID, func(item types.CallLog) string {
		return item.CallerID
	})
	items = common.FilterByDateRangeCreatedAt(items, params.From, params.To, func(item types.CallLog) string {
		return item.CreatedAt
	})
	items = common.FilterByKeyword(items, params.Keyword, []func(types.CallLog) string{
		func(item types.CallLog) string { return item.InputPreview },
		func(item types.CallLog) string { return item.OutputPreview },
		func(item types.CallLog) string { return item.Caller },
		func(item types.CallLog) string { return item.Model },
	})
	paged, total, page, pageSize := common.Paginate(items, params.Page, params.PageSize)
	return types.PageResult[types.CallLog]{
		Items: paged, Total: total, Page: page, PageSize: pageSize,
	}, nil
}
