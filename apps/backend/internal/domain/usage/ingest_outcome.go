package usage

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

// IngestErrorKind classifies ingest errors for retry decisions.
type IngestErrorKind int

const (
	IngestOK IngestErrorKind = iota
	IngestBusiness
	IngestLogNotFound
	IngestLogDBTemp
	IngestRetryable
)

// ClassifyIngestError determines the error kind for River retry/cancel decisions.
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
			return IngestRetryable
		default:
			return IngestBusiness
		}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return IngestLogDBTemp
	}
	return IngestRetryable
}

// IsRecoverableIngestError reports whether a business error is recoverable
// (e.g. mapping not found yet — may appear after replication lag).
func IsRecoverableIngestError(err error) bool {
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Status == domain.StatusNotFound
}
