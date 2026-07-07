package budget_test

import (
	"encoding/json"
	"log/slog"
	"testing"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	mock "github.com/tokenjoy/backend/tests/testutil/mock"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

// PRD 3.2: 预警规则
// - "达到 100% 时阻断请求"
// - "通知发送失败不影响阻断逻辑"

func TestOverrunNotifiesOnDepartmentExhaustion(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.RecordingNotifier{}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := domainbudget.NewOverrunService(cfg, st, lifecycle, notifier, slog.Default())

	// Set budget fully consumed
	tree, _ := common.LoadBudgetTree(ctx, st.Org().Nodes())
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 25000) // seed budget is 25000
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	if len(notifier.Notifications) == 0 {
		t.Fatal("expected overrun notification to be sent")
	}
	if notifier.Notifications[0].EventType != types.NotificationEventOverrunBlocked {
		t.Errorf("expected event %s, got %s", types.NotificationEventOverrunBlocked, notifier.Notifications[0].EventType)
	}
}

func TestOverrunBlocksEvenIfNotificationFails(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.FailingNotifier{}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := domainbudget.NewOverrunService(cfg, st, lifecycle, notifier, slog.Default())

	tree, _ := common.LoadBudgetTree(ctx, st.Org().Nodes())
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 25000)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	// Should still succeed — notification failure does NOT block overrun processing
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
}

func TestOverrunSkipsWhenBudgetNotExhausted(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.RecordingNotifier{}
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := domainbudget.NewOverrunService(cfg, st, lifecycle, notifier, slog.Default())

	// Budget at 50% — NOT exhausted
	tree, _ := common.LoadBudgetTree(ctx, st.Org().Nodes())
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 12500) // 50% of 25000
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	if len(notifier.Notifications) > 0 {
		t.Error("should NOT notify when budget is not exhausted")
	}
}
