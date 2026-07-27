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

	logsBefore := testutil.NotificationLogs(st)

	if err := svc.Send(testutil.Ctx(), types.Notification{
		EventType: types.NotificationEventSyncThreshold,
		Payload:   map[string]any{"detail": "test"},
	}); err != nil {
		t.Fatal(err)
	}

	logsAfter := testutil.NotificationLogs(st)
	if len(logsAfter) != len(logsBefore)+1 {
		t.Fatalf("expected +1 notification log, before=%d after=%d", len(logsBefore), len(logsAfter))
	}
	// The new entry should be the last one (ordered by created_at).
	newest := logsAfter[len(logsAfter)-1]
	if newest.EventType != types.NotificationEventSyncThreshold {
		t.Fatalf("unexpected event type %s", newest.EventType)
	}
}

func TestWebhookNotTriggeredByFallbackChain(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, func(c *config.Config) {
		c.NotifyWebhookURL = "http://127.0.0.1:1/unreachable"
	})
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := notification.NewService(cfg, st, logger)

	logsBefore := testutil.NotificationLogs(st)

	err := svc.Send(testutil.Ctx(), types.Notification{
		EventType: types.NotificationEventOverrunBlocked,
		Payload:   map[string]any{"scope": "member"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// New dispatch behavior: normal priority only uses in_app channel.
	// Webhook is registered but not in the fallback chain, so it's never called.
	logsAfter := testutil.NotificationLogs(st)
	if len(logsAfter) != len(logsBefore)+1 {
		t.Fatalf("expected +1 notification log, before=%d after=%d", len(logsBefore), len(logsAfter))
	}
	newest := logsAfter[len(logsAfter)-1]
	if newest.Channel != types.NotificationChannelInApp {
		t.Fatalf("expected channel in_app, got %s", newest.Channel)
	}
}
