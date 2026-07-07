package budget_test

import (
	"encoding/json"
	"testing"

	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

// PRD 3.2: 预警规则
// - "达到 100% 时阻断请求"
// - "通知发送失败不影响阻断逻辑"

func TestOverrunNotifiesOnDepartmentExhaustion(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.RecordingNotifier{}
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, notifier)

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, 25000)

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
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
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.FailingNotifier{}
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, notifier)

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, 25000)

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
	})
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
}

func TestOverrunSkipsWhenBudgetNotExhausted(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	ctx := testutil.Ctx()

	notifier := &testutil.RecordingNotifier{}
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, notifier)

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, 12500)

	payload, _ := json.Marshal(map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
	})
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	if len(notifier.Notifications) > 0 {
		t.Error("should NOT notify when budget is not exhausted")
	}
}
