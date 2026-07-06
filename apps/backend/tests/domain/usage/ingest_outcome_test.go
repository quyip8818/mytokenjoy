package usage_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestOutcomeForWebhook(t *testing.T) {
	cases := []struct {
		name           string
		err            error
		wantStatus     int
		wantNotify     bool
		wantRecordFail bool
	}{
		{"ok", nil, http.StatusOK, true, false},
		{"log not found", store.ErrConsumeLogNotFound, http.StatusServiceUnavailable, false, false},
		{"business", domain.NotFound("mapping"), http.StatusOK, false, true},
		{"temp", domain.ServiceUnavailable("db"), http.StatusInternalServerError, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			webhook := domainusage.OutcomeFor(tc.err).Webhook()
			if webhook.Status != tc.wantStatus || webhook.RecordNotify != tc.wantNotify || webhook.RecordFailure != tc.wantRecordFail {
				t.Fatalf("webhook = %+v", webhook)
			}
		})
	}
}

func TestOutcomeForReconcileAdvancesCursor(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		advance bool
	}{
		{"ok", nil, true},
		{"log not found", store.ErrConsumeLogNotFound, true},
		{"business", domain.NotFound("mapping"), true},
		{"log db temp", errors.New("pg"), false},
		{"tokenjoy temp", domain.ServiceUnavailable("tx"), false},
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
		{"max attempts", store.ErrConsumeLogNotFound, store.IngestFailureMaxAttempts - 1, domainusage.RetryDead},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainusage.OutcomeFor(tc.err).Retry(tc.att); got != tc.want {
				t.Fatalf("retry = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWebhookIngestResultForMatchesOutcome(t *testing.T) {
	errs := []error{nil, store.ErrConsumeLogNotFound, domain.NotFound("m"), domain.ServiceUnavailable("x")}
	for _, err := range errs {
		if domainusage.WebhookIngestResultFor(err) != domainusage.OutcomeFor(err).Webhook() {
			t.Fatalf("mismatch for err %v", err)
		}
	}
}
