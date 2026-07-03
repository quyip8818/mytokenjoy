package audit

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSettings(ctx context.Context) (types.AuditSettings, error)
	UpdateSettings(ctx context.Context, settings types.AuditSettings) (types.AuditSettings, error)
	ListOperations(ctx context.Context, params types.AuditOperationsQueryParams) (types.PageResult[types.OperationLog], error)
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
