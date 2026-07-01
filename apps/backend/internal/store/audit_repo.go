package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type AuditRepository interface {
	Settings(ctx context.Context) (types.AuditSettings, error)
	SetSettings(ctx context.Context, settings types.AuditSettings) error
	OperationLogs(ctx context.Context) ([]types.OperationLog, error)
	AppendOperationLog(ctx context.Context, log types.OperationLog) error
	CallLogs(ctx context.Context) ([]types.CallLog, error)
}
