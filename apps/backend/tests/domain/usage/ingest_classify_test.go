package usage_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestIsRecoverableIngestError(t *testing.T) {
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

func TestWebhookIngestResultFor(t *testing.T) {
	ok := domainusage.WebhookIngestResultFor(nil)
	if ok.Status != http.StatusOK || !ok.RecordNotify || ok.RecordFailure {
		t.Fatalf("unexpected ok result: %+v", ok)
	}
	notFound := domainusage.WebhookIngestResultFor(store.ErrConsumeLogNotFound)
	if notFound.Status != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", notFound.Status)
	}
	business := domainusage.WebhookIngestResultFor(domain.NotFound("mapping missing"))
	if business.Status != http.StatusOK || !business.RecordFailure || business.RecordNotify {
		t.Fatalf("unexpected business result: %+v", business)
	}
	temp := domainusage.WebhookIngestResultFor(domain.ServiceUnavailable("db busy"))
	if temp.Status != http.StatusInternalServerError || temp.RecordFailure || temp.RecordNotify {
		t.Fatalf("unexpected temp result: %+v", temp)
	}
}

func TestClassifyIngestError(t *testing.T) {
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
		{"tokenjoy temp", domain.ServiceUnavailable("tx"), domainusage.IngestTokenjoyTemp},
		{"generic", errors.New("network"), domainusage.IngestTokenjoyTemp},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := domainusage.ClassifyIngestError(tc.err); got != tc.want {
				t.Fatalf("ClassifyIngestError() = %v, want %v", got, tc.want)
			}
		})
	}
}
