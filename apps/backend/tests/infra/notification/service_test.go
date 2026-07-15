package notification_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestNotifierWritesLogEntry(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := notification.NewService(cfg, st, logger)

	if err := svc.Send(testutil.Ctx(), types.Notification{
		EventType: types.NotificationEventSyncThreshold,
		Recipient: "ops",
		Payload:   map[string]any{"detail": "test"},
	}); err != nil {
		t.Fatal(err)
	}

	logs := testutil.NotificationLogs(st)
	if len(logs) != 1 {
		t.Fatalf("expected 1 notification log, got %d", len(logs))
	}
	if logs[0].EventType != types.NotificationEventSyncThreshold {
		t.Fatalf("unexpected event type %s", logs[0].EventType)
	}
}

func TestWebhookNotTriggeredByFallbackChain(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, func(c *config.Config) {
		c.NotifyWebhookURL = "http://127.0.0.1:1/unreachable"
	})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := notification.NewService(cfg, st, logger)

	err := svc.Send(testutil.Ctx(), types.Notification{
		EventType: types.NotificationEventOverrunBlocked,
		Recipient: "ops",
		Payload:   map[string]any{"scope": "member"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// New dispatch behavior: normal priority only uses in_app channel.
	// Webhook is registered but not in the fallback chain, so it's never called.
	logs := testutil.NotificationLogs(st)
	if len(logs) != 1 {
		t.Fatalf("expected 1 notification log (in_app), got %d", len(logs))
	}
	if logs[0].Channel != types.NotificationChannelInApp {
		t.Fatalf("expected channel in_app, got %s", logs[0].Channel)
	}
}
