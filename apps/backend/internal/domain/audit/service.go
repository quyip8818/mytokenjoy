package audit

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSettings(ctx context.Context) (types.AuditSettings, error)
	UpdateSettings(ctx context.Context, settings types.AuditSettings) (types.AuditSettings, error)
	ListOperations(ctx context.Context, params types.AuditOperationsQueryParams) (types.PageResult[types.OperationLog], error)
	ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error)
	OperationsTimeline(ctx context.Context, params types.AuditOperationsQueryParams) ([]types.OperationDailyCount, error)
	CallsSummary(ctx context.Context, params types.AuditCallsQueryParams) (types.CallsSummary, error)
}

// Store is the narrow store surface the audit domain needs.
type Store interface {
	Audit() store.AuditRepository
	Ledger() store.LedgerRepository
}

type service struct {
	cfg    config.Config
	store  Store
	reader domainusage.ReadModel
}

func NewService(cfg config.Config, st Store, reader domainusage.ReadModel) Service {
	return &service{cfg: cfg, store: st, reader: reader}
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
		return types.AuditSettings{}, err
	}
	return current, nil
}

func (s *service) ListOperations(ctx context.Context, params types.AuditOperationsQueryParams) (types.PageResult[types.OperationLog], error) {
	page, pageSize := types.NormalizePageParams(params.Page, params.PageSize)
	items, total, err := s.store.Audit().ListOperationsPage(ctx, store.AuditOperationFilter{
		Action:     params.Action,
		OperatorID: params.OperatorID,
		Keyword:    params.Keyword,
		From:       params.From,
		To:         params.To,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return types.PageResult[types.OperationLog]{}, err
	}
	return types.PageResult[types.OperationLog]{
		Items: items, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func (s *service) ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error) {
	return s.reader.ListCalls(ctx, params)
}

func (s *service) OperationsTimeline(ctx context.Context, params types.AuditOperationsQueryParams) ([]types.OperationDailyCount, error) {
	return s.store.Audit().OperationCountsByDay(ctx, store.AuditOperationFilter{
		Action:     params.Action,
		OperatorID: params.OperatorID,
		Keyword:    params.Keyword,
		From:       params.From,
		To:         params.To,
	})
}

func (s *service) CallsSummary(ctx context.Context, params types.AuditCallsQueryParams) (types.CallsSummary, error) {
	return s.store.Ledger().CallsSummary(ctx, store.LedgerCallFilter{
		Model:    params.Model,
		Status:   params.Status,
		CallerID: params.CallerID,
		Keyword:  params.Keyword,
		From:     params.From,
		To:       params.To,
	})
}
