package usage_test

import (
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestClassifyIngestError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want domainusage.IngestErrorKind
	}{
		{"nil", nil, domainusage.IngestOK},
		{"log not found", store.ErrConsumeLogNotFound, domainusage.IngestLogNotFound},
		{"mapping missing", domain.NotFound("mapping"), domainusage.IngestBusiness},
		{"bad request", domain.BadRequest("invalid"), domainusage.IngestBusiness},
		{"unprocessable", domain.Validation("pricing"), domainusage.IngestBusiness},
		{"retryable service", domain.ServiceUnavailable("tx"), domainusage.IngestRetryable},
		{"generic", errors.New("network"), domainusage.IngestRetryable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainusage.ClassifyIngestError(tc.err); got != tc.want {
				t.Fatalf("ClassifyIngestError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsRecoverableIngestError(t *testing.T) {
	t.Parallel()
	if !domainusage.IsRecoverableIngestError(domain.NotFound("mapping not found for token 1")) {
		t.Fatal("expected mapping not found to be recoverable")
	}
	if domainusage.IsRecoverableIngestError(domain.BadRequest("invalid")) {
		t.Fatal("expected bad request to be permanent")
	}
	if domainusage.IsRecoverableIngestError(store.ErrConsumeLogNotFound) {
		t.Fatal("expected log not found to use retry path via classify, not recoverable business")
	}
	if domainusage.IsRecoverableIngestError(errors.New("db down")) {
		t.Fatal("expected generic error to be non-recoverable business")
	}
}
