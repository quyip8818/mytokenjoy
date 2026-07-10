package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

type queue struct {
	logStore store.LogStore
}

func NewQueue(logStore store.LogStore) Queue {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	return &queue{logStore: logStore}
}

func (q *queue) Enqueue(ctx context.Context, logID int64, source string) error {
	return q.logStore.EnqueuePending(ctx, logID, source)
}

func (q *queue) RecordFailure(ctx context.Context, logID int64, source string, err error) error {
	if err == nil {
		return nil
	}
	return q.logStore.UpsertJob(ctx, store.IngestJobFromError(logID, source, err))
}

func (q *queue) ApplyRetry(ctx context.Context, job store.IngestJob, ingestErr error) error {
	errMsg := ""
	if ingestErr != nil {
		errMsg = ingestErr.Error()
	}
	switch OutcomeFor(ingestErr).Retry(job.Attempts) {
	case RetryDone:
		return q.logStore.MarkJobDone(ctx, job.ID)
	case RetryDead:
		return q.logStore.MarkJobDead(ctx, job.ID, errMsg)
	default:
		return q.logStore.MarkJobRetry(ctx, job.ID, IngestBackoff(job.Attempts+1), errMsg)
	}
}
