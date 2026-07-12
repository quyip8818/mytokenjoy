package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type AuditOperationFilter struct {
	Action     string
	OperatorID string
	Keyword    string
	From       string
	To         string
	Page       int
	PageSize   int
}

type AuditRepository interface {
	Settings(ctx context.Context) (types.AuditSettings, error)
	SetSettings(ctx context.Context, settings types.AuditSettings) error
	OperationLogs(ctx context.Context) ([]types.OperationLog, error)
	ListOperationsPage(ctx context.Context, filter AuditOperationFilter) ([]types.OperationLog, int, error)
	OperationCountsByDay(ctx context.Context, filter AuditOperationFilter) ([]types.OperationDailyCount, error)
	AppendOperationLog(ctx context.Context, log types.OperationLog) error
}
