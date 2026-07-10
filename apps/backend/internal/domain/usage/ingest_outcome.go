package usage

import (
	"errors"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestErrorKind int

const (
	IngestOK IngestErrorKind = iota
	IngestBusiness
	IngestLogNotFound
	IngestLogDBTemp
	IngestTokenjoyTemp
)

func ClassifyIngestError(err error) IngestErrorKind {
	if err == nil {
		return IngestOK
	}
	if errors.Is(err, store.ErrConsumeLogNotFound) {
		return IngestLogNotFound
	}
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Status {
		case domain.StatusNotFound, domain.StatusUnprocessable, domain.StatusBadRequest:
			return IngestBusiness
		case domain.StatusServiceUnavailable:
			return IngestTokenjoyTemp
		default:
			return IngestBusiness
		}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return IngestLogDBTemp
	}
	return IngestTokenjoyTemp
}

func IsRecoverableIngestError(err error) bool {
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Status == domain.StatusNotFound
}

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
	if attempts+1 >= store.IngestJobMaxAttempts {
		return RetryDead
	}
	return RetryScheduleBackoff
}

func IngestBackoff(attempts int) time.Duration {
	seconds := math.Min(300, math.Pow(2, float64(attempts)))
	return time.Duration(seconds) * time.Second
}
