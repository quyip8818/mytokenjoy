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

func TestOutcomeForReconcileAdvancesCursor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		err     error
		advance bool
	}{
		{"ok", nil, true},
		{"log not found", store.ErrConsumeLogNotFound, true},
		{"business", domain.NotFound("mapping"), true},
		{"log db temp", errors.New("pg"), false},
		{"retryable", domain.ServiceUnavailable("tx"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainusage.OutcomeFor(tc.err).ReconcileAdvancesCursor(); got != tc.advance {
				t.Fatalf("advance = %v, want %v", got, tc.advance)
			}
		})
	}
}

func TestOutcomeShouldRecordFailure(t *testing.T) {
	t.Parallel()
	if !domainusage.OutcomeFor(domain.BadRequest("x")).ShouldRecordFailure() {
		t.Fatal("expected business failure to record")
	}
	if domainusage.OutcomeFor(nil).ShouldRecordFailure() {
		t.Fatal("expected ok not to record failure")
	}
	if domainusage.OutcomeFor(store.ErrConsumeLogNotFound).ShouldRecordFailure() {
		t.Fatal("expected log not found not to record failure")
	}
}

func TestOutcomeRetryDisposition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		att  int
		want domainusage.RetryDisposition
	}{
		{"ok", nil, 0, domainusage.RetryDone},
		{"log not found", store.ErrConsumeLogNotFound, 0, domainusage.RetryScheduleBackoff},
		{"recoverable business", domain.NotFound("mapping"), 0, domainusage.RetryScheduleBackoff},
		{"permanent business", domain.BadRequest("invalid"), 0, domainusage.RetryDead},
		{"max attempts", store.ErrConsumeLogNotFound, store.IngestJobMaxAttempts - 1, domainusage.RetryDead},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainusage.OutcomeFor(tc.err).Retry(tc.att); got != tc.want {
				t.Fatalf("retry = %v, want %v", got, tc.want)
			}
		})
	}
}
