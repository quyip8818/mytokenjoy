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

	IngestJobStatusPending = "pending"
	IngestJobStatusDead    = "dead"

	IngestJobMaxAttempts = 20

	ingestJobClaimLease = 5 * time.Minute
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

type IngestJob struct {
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
	GetConsumeLogsByIDs(ctx context.Context, logIDs []int64) ([]RawConsumeLog, error)
	ListConsumeLogIDsAfter(ctx context.Context, afterID int64, limit int) ([]int64, error)
	GetReconcileCursor(ctx context.Context, stream string) (int64, error)
	SetReconcileCursor(ctx context.Context, stream string, logID int64) error
	EnqueuePending(ctx context.Context, logID int64, source string) error
	UpsertJob(ctx context.Context, job IngestJob) error
	ClaimPendingJobs(ctx context.Context, limit int) ([]IngestJob, error)
	MarkJobDone(ctx context.Context, id string) error
	MarkJobRetry(ctx context.Context, id string, delay time.Duration, errMsg string) error
	MarkJobDead(ctx context.Context, id string, errMsg string) error
	CountConsumeLogsAfter(ctx context.Context, afterID int64) (int64, error)
	CountPendingIngestJobs(ctx context.Context) (int, error)
	IngestLagSeconds(ctx context.Context, afterID int64) (int64, error)
}

func IngestJobID(logID int64) string {
	return fmt.Sprintf("ij-%d", logID)
}

func IngestJobFromError(logID int64, source string, err error) IngestJob {
	return IngestJob{
		ID:     IngestJobID(logID),
		LogID:  logID,
		Source: source,
		Error:  err.Error(),
	}
}

func IngestJobClaimLease() time.Duration {
	return ingestJobClaimLease
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

func (noopLogStore) EnqueuePending(context.Context, int64, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) UpsertJob(context.Context, IngestJob) error {
	return errors.New("log store not configured")
}

func (noopLogStore) ClaimPendingJobs(context.Context, int) ([]IngestJob, error) {
	return nil, errors.New("log store not configured")
}

func (noopLogStore) MarkJobDone(context.Context, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) MarkJobRetry(context.Context, string, time.Duration, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) MarkJobDead(context.Context, string, string) error {
	return errors.New("log store not configured")
}

func (noopLogStore) CountConsumeLogsAfter(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) CountPendingIngestJobs(context.Context) (int, error) {
	return 0, errors.New("log store not configured")
}

func (noopLogStore) IngestLagSeconds(context.Context, int64) (int64, error) {
	return 0, errors.New("log store not configured")
}

func NoopLogStore() LogStore {
	return noopLogStore{}
}
