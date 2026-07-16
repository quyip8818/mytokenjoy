package store

import (
	"context"
	"errors"
)

const (
	NewAPILogTableName           = "logs"
	NewAPILogTypeConsume         = 2
	ReconcileStreamNewAPIConsume = "newapi_consume"
)

var ErrConsumeLogNotFound = errors.New("consume log not found")

type RawConsumeLog struct {
	ID               int64
	TokenID          int64
	Quota            int64
	ModelName        string
	CreatedAt        int64
	PromptTokens     int64
	CompletionTokens int64
	UseTime          int64
	Content          string
}

type LogStore interface {
	GetConsumeLogByID(ctx context.Context, logID int64) (*RawConsumeLog, error)
	GetConsumeLogsByIDs(ctx context.Context, logIDs []int64) ([]RawConsumeLog, error)
	ListConsumeLogIDsAfter(ctx context.Context, afterID int64, limit int) ([]int64, error)
	GetReconcileCursor(ctx context.Context, stream string) (int64, error)
	SetReconcileCursor(ctx context.Context, stream string, logID int64) error
	CountConsumeLogsAfter(ctx context.Context, afterID int64) (int64, error)
	IngestLagSeconds(ctx context.Context, afterID int64) (int64, error)
}

type noopLogStore struct{}

func (noopLogStore) GetConsumeLogByID(context.Context, int64) (*RawConsumeLog, error) {
	return nil, errors.New("log store not configured")
}

func (noopLogStore) GetConsumeLogsByIDs(context.Context, []int64) ([]RawConsumeLog, error) {
	return nil, errors.New("log store not configured")
}

func (noopLogStore) ListConsumeLogIDsAfter(context.Context, int64, int) ([]int64, error) {
	return nil, errors.New("log store not configured")
}

func (noopLogStore) GetReconcileCursor(context.Context, string) (int64, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) SetReconcileCursor(context.Context, string, int64) error {
	return errors.New("log store not configured")
}

func (noopLogStore) CountConsumeLogsAfter(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) IngestLagSeconds(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func NoopLogStore() LogStore {
	return noopLogStore{}
}
