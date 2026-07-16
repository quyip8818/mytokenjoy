package notification_test

import (
	"testing"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/infra/notification"
)

func TestRendererExtractsTitleFromPayload(t *testing.T) {
	t.Parallel()
	r := notification.NewRenderer()

	event := domainnotification.Event{
		EventType: "budget_alert_reached",
		Payload: map[string]any{
			"title": "Budget Exceeded",
			"body":  "Your department exceeded 90% of budget",
		},
	}

	msg := r.Render(event)
	if msg.Title != "Budget Exceeded" {
		t.Fatalf("expected title 'Budget Exceeded', got %q", msg.Title)
	}
	if msg.Body != "Your department exceeded 90% of budget" {
		t.Fatalf("expected body about budget, got %q", msg.Body)
	}
}

func TestRendererUsesDefaultTitle(t *testing.T) {
	t.Parallel()
	r := notification.NewRenderer()

	event := domainnotification.Event{
		EventType: domainnotification.EventBudgetAlertReached,
		Payload:   map[string]any{},
	}

	msg := r.Render(event)
	if msg.Title != "Budget Alert" {
		t.Fatalf("expected default title 'Budget Alert', got %q", msg.Title)
	}
}

func TestRendererEnrichesPayloadWithEventType(t *testing.T) {
	t.Parallel()
	r := notification.NewRenderer()

	event := domainnotification.Event{
		EventType: "key_expired",
		Payload:   map[string]any{"keyId": "k-123"},
	}

	msg := r.Render(event)
	if msg.Payload["eventType"] != "key_expired" {
		t.Fatalf("expected eventType in payload, got %v", msg.Payload["eventType"])
	}
	if msg.Payload["keyId"] != "k-123" {
		t.Fatal("expected original payload fields preserved")
	}
}

func TestRendererFallsBackToMessageField(t *testing.T) {
	t.Parallel()
	r := notification.NewRenderer()

	event := domainnotification.Event{
		EventType: domainnotification.EventOverrunBlocked,
		Payload:   map[string]any{"message": "overrun details here"},
	}

	msg := r.Render(event)
	if msg.Body != "overrun details here" {
		t.Fatalf("expected body from message field, got %q", msg.Body)
	}
}
