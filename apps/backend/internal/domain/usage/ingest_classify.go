package usage

import (
	"errors"

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

type WebhookIngestResult struct {
	Status        int
	RecordFailure bool
	RecordNotify  bool
	Message       string
}

func WebhookIngestResultFor(err error) WebhookIngestResult {
	return OutcomeFor(err).Webhook()
}
