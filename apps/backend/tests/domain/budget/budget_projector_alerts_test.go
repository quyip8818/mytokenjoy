package budget_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverTF "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestBudgetProjectorFiresHighestCrossedAlertThreshold(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	budgetAmt, found, err := st.Org().Nodes().GetNodeBudget(ctx, contract.IDDept3)
	if err != nil || !found || budgetAmt <= 0 {
		t.Fatalf("expected dept-3 budget, found=%v err=%v", found, err)
	}
	current, err := st.Ledger().SumAmountByDepartment(ctx, contract.IDDept3, contract.DemoBudgetPeriod)
	if err != nil {
		t.Fatal(err)
	}
	// Bring spend into [90%, 100%) so the highest crossed threshold is 90.
	target := budgetAmt*0.95 - current
	if target > 0 {
		budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, target)
	}

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(7601, 99))
	if err := ingest.IngestByLogID(ctx, 7601, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	notifier := &testutil.RecordingNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projector := budget.NewAsync(
		cfg, st,
		riverTF.NewBudgetInsertOnlyEnqueuer(t, cfg, st),
		budget.NoopGatewaySoftCache,
		logger,
		budget.WithProjectorNotifier(notifier),
	).Projector

	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	var alert *types.Notification
	for i := range notifier.Notifications {
		if notifier.Notifications[i].EventType == types.NotificationEventBudgetAlertReached {
			alert = &notifier.Notifications[i]
			break
		}
	}
	if alert == nil {
		t.Fatal("expected budget_alert_reached notification")
	}
	if alert.Recipient != "department:"+contract.IDDept3 {
		t.Fatalf("recipient = %q, want department:%s", alert.Recipient, contract.IDDept3)
	}
	threshold, ok := alert.Payload["threshold"].(int)
	if !ok {
		// JSON numbers may round-trip as int via map[string]any when set directly.
		if f, fok := alert.Payload["threshold"].(float64); fok {
			threshold = int(f)
			ok = true
		}
	}
	if !ok || threshold != 90 {
		t.Fatalf("threshold = %v, want 90 (highest crossed below 100)", alert.Payload["threshold"])
	}

	firstCount := countBudgetAlerts(notifier)
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(7602, 99))
	if err := ingest.IngestByLogID(ctx, 7602, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if got := countBudgetAlerts(notifier); got != firstCount {
		t.Fatalf("alert dedup failed: before=%d after=%d", firstCount, got)
	}
}

func TestBudgetProjectorSkipsDisabledAlertRules(t *testing.T) {
	t.Parallel()
	stub := defaultBudgetIngestStub()
	cfg, st, ingest, _ := newBudgetIngestFixture(t, stub)
	ctx := testutil.Ctx()

	rules, err := st.Budget().AlertRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range rules {
		if rules[i].NodeID == contract.IDDept3 {
			rules[i].Enabled = false
		}
	}
	if err := st.Budget().SetAlertRules(ctx, rules); err != nil {
		t.Fatal(err)
	}

	budgetAmt, found, err := st.Org().Nodes().GetNodeBudget(ctx, contract.IDDept3)
	if err != nil || !found {
		t.Fatal("expected dept-3 budget")
	}
	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, budgetAmt)

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.DefaultMappingOpts())
	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(7611, 99))
	if err := ingest.IngestByLogID(ctx, 7611, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	notifier := &testutil.RecordingNotifier{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	projector := budget.NewAsync(
		cfg, st,
		riverTF.NewBudgetInsertOnlyEnqueuer(t, cfg, st),
		budget.NoopGatewaySoftCache,
		logger,
		budget.WithProjectorNotifier(notifier),
	).Projector
	if _, err := projector.RunBatch(ctx, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}
	if countBudgetAlerts(notifier) != 0 {
		t.Fatalf("disabled rules must not notify, got %d", countBudgetAlerts(notifier))
	}
}

func countBudgetAlerts(n *testutil.RecordingNotifier) int {
	count := 0
	for _, item := range n.Notifications {
		if item.EventType == types.NotificationEventBudgetAlertReached {
			count++
		}
	}
	return count
}
