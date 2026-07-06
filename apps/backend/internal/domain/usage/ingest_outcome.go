package usage

import (
	"math"
	"net/http"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

type IngestOutcome struct {
	kind IngestErrorKind
	err  error
}

func OutcomeFor(err error) IngestOutcome {
	return IngestOutcome{kind: ClassifyIngestError(err), err: err}
}

type RetryDisposition int

const (
	RetryDone RetryDisposition = iota
	RetryDead
	RetryScheduleBackoff
)

func (o IngestOutcome) Webhook() WebhookIngestResult {
	switch o.kind {
	case IngestOK:
		return WebhookIngestResult{
			Status:       http.StatusOK,
			RecordNotify: true,
			Message:      "ok",
		}
	case IngestLogNotFound:
		return WebhookIngestResult{
			Status:  http.StatusServiceUnavailable,
			Message: "consume log not visible",
		}
	case IngestBusiness:
		return WebhookIngestResult{
			Status:        http.StatusOK,
			RecordFailure: true,
			Message:       "accepted",
		}
	default:
		return WebhookIngestResult{
			Status:  http.StatusInternalServerError,
			Message: "ingest failed",
		}
	}
}

func (o IngestOutcome) ReconcileAdvancesCursor() bool {
	switch o.kind {
	case IngestOK, IngestBusiness, IngestLogNotFound:
		return true
	default:
		return false
	}
}

func (o IngestOutcome) ShouldRecordFailure() bool {
	return o.kind == IngestBusiness
}

func (o IngestOutcome) Retry(attempts int) RetryDisposition {
	switch o.kind {
	case IngestOK:
		return RetryDone
	case IngestBusiness:
		if IsRecoverableIngestError(o.err) {
			return retryBackoffOrDead(attempts)
		}
		return RetryDead
	default:
		return retryBackoffOrDead(attempts)
	}
}

func retryBackoffOrDead(attempts int) RetryDisposition {
	if attempts+1 >= store.IngestFailureMaxAttempts {
		return RetryDead
	}
	return RetryScheduleBackoff
}

func IngestBackoff(attempts int) time.Duration {
	seconds := math.Min(300, math.Pow(2, float64(attempts)))
	return time.Duration(seconds) * time.Second
}
