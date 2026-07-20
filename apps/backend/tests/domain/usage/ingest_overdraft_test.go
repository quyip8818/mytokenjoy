package usage_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/adapter"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func TestIngestNotifiesOnOverdraftExpansion(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ctx := testutil.Ctx()
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || co == nil {
		t.Fatal("expected default company")
	}
	if co.WalletRemain > 0 {
		if _, err := billinglot.ConsumeLots(ctx, st, contract.DefaultCompanyID, co.WalletRemain); err != nil {
			t.Fatal(err)
		}
	}

	notifier := &testutil.RecordingNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	budgetOps := adapter.NewUsageBudgetOps(nil, nil, logger)
	lotConsumer := adapter.NewUsageLotConsumer()
	ingest := usage.NewIngestService(
		cfg, st, st.Logs(), logger,
		adapter.NewUsageIngestEnqueuer(jobs.NoopEnqueuer{}),
		notifier,
		budgetOps, lotConsumer,
	)

	const logID int64 = 7101
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(logID, 99))
	if err := ingest.IngestByLogID(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	if len(notifier.Notifications) != 1 {
		t.Fatalf("expected 1 overdraft notification, got %d", len(notifier.Notifications))
	}
	n := notifier.Notifications[0]
	if n.EventType != types.NotificationEventOverdraftExpanded {
		t.Fatalf("event = %q, want %q", n.EventType, types.NotificationEventOverdraftExpanded)
	}
	delta, ok := n.Payload["overdraftDelta"].(int64)
	if !ok || delta <= 0 {
		t.Fatalf("expected positive overdraftDelta, got %v", n.Payload["overdraftDelta"])
	}

	if err := ingest.IngestByLogID(ctx, logID, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if len(notifier.Notifications) != 1 {
		t.Fatalf("idempotent retry must not re-notify, got %d", len(notifier.Notifications))
	}
}

func TestIngestSkipsOverdraftNotificationWhenNotifierNil(t *testing.T) {
	t.Parallel()
	cfg, st := newIngestStore(t)
	ctx := testutil.Ctx()
	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())

	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || co == nil {
		t.Fatal("expected default company")
	}
	if co.WalletRemain > 0 {
		if _, err := billinglot.ConsumeLots(ctx, st, contract.DefaultCompanyID, co.WalletRemain); err != nil {
			t.Fatal(err)
		}
	}

	ingest := testutil.NewIngestService(t, cfg, st)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(7102, 99))
	if err := ingest.IngestByLogID(ctx, 7102, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
}
