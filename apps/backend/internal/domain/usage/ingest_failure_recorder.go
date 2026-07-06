package usage

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type FailureRecorder interface {
	RecordFailure(ctx context.Context, logID int64, source string, err error) error
	ApplyRetry(ctx context.Context, failure store.IngestFailure, ingestErr error) error
}

type failureRecorder struct {
	logStore store.LogStore
	logger   *slog.Logger
}

func NewFailureRecorder(logStore store.LogStore, logger *slog.Logger) FailureRecorder {
	if logger == nil {
		logger = slog.Default()
	}
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	return &failureRecorder{logStore: logStore, logger: logger}
}

func (r *failureRecorder) RecordFailure(ctx context.Context, logID int64, source string, err error) error {
	if err == nil {
		return nil
	}
	return r.logStore.UpsertFailure(ctx, store.IngestFailureFromError(logID, source, err))
}

func (r *failureRecorder) ApplyRetry(ctx context.Context, failure store.IngestFailure, ingestErr error) error {
	errMsg := ""
	if ingestErr != nil {
		errMsg = ingestErr.Error()
	}
	switch OutcomeFor(ingestErr).Retry(failure.Attempts) {
	case RetryDone:
		return r.logStore.MarkFailureDone(ctx, failure.ID)
	case RetryDead:
		return r.logStore.MarkFailureDead(ctx, failure.ID, errMsg)
	case RetryScheduleBackoff:
		next := time.Now().Add(IngestBackoff(failure.Attempts + 1))
		return r.logStore.MarkFailureRetry(ctx, failure.ID, next, errMsg)
	default:
		return nil
	}
}
