package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	NewAPILogTableName           = "logs"
	NewAPILogTypeConsume         = 2
	ReconcileStreamNewAPIConsume = "newapi_consume"

	IngestFailureStatusPending = "pending"
	IngestFailureStatusDead    = "dead"

	IngestFailureMaxAttempts = 20

	ingestFailureClaimLease = 5 * time.Minute
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

type IngestFailure struct {
	ID        string
	LogID     int64
	Source    string
	Error     string
	Status    string
	Attempts  int
	NextRetry time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LogStore interface {
	GetConsumeLogByID(ctx context.Context, logID int64) (*RawConsumeLog, error)
	ListConsumeLogIDsAfter(ctx context.Context, afterID int64, limit int) ([]int64, error)
	GetReconcileCursor(ctx context.Context, stream string) (int64, error)
	SetReconcileCursor(ctx context.Context, stream string, logID int64) error
	UpsertFailure(ctx context.Context, f IngestFailure) error
	ClaimPendingFailures(ctx context.Context, limit int) ([]IngestFailure, error)
	MarkFailureDone(ctx context.Context, id string) error
	MarkFailureRetry(ctx context.Context, id string, next time.Time, errMsg string) error
	MarkFailureDead(ctx context.Context, id string, errMsg string) error
	CountConsumeLogsAfter(ctx context.Context, afterID int64) (int64, error)
	CountPendingIngestFailures(ctx context.Context) (int, error)
	IngestLagSeconds(ctx context.Context, afterID int64) (int64, error)
}

func IngestFailureID(logID int64) string {
	return fmt.Sprintf("if-%d", logID)
}

func IngestFailureFromError(logID int64, source string, err error) IngestFailure {
	return IngestFailure{
		ID:     IngestFailureID(logID),
		LogID:  logID,
		Source: source,
		Error:  err.Error(),
	}
}

func FailureClaimLease() time.Duration {
	return ingestFailureClaimLease
}

type noopLogStore struct{}

func (noopLogStore) GetConsumeLogByID(context.Context, int64) (*RawConsumeLog, error) {
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

func (noopLogStore) UpsertFailure(context.Context, IngestFailure) error {
	return errors.New("log store not configured")
}

func (noopLogStore) ClaimPendingFailures(context.Context, int) ([]IngestFailure, error) {
	return nil, errors.New("log store not configured")
}

func (noopLogStore) MarkFailureDone(context.Context, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) MarkFailureRetry(context.Context, string, time.Time, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) MarkFailureDead(context.Context, string, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) CountConsumeLogsAfter(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) CountPendingIngestFailures(context.Context) (int, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) IngestLagSeconds(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func NoopLogStore() LogStore {
	return noopLogStore{}
}
