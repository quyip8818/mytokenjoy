package budget_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

// --- Alert rule CRUD and threshold tests ---

func TestCreateAlertRuleWithMultipleThresholds(t *testing.T) {
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:        seed.IDDept3,
		NodeName:      "后端组",
		Thresholds:    []int{80, 90, 100},
		NotifyRoleIDs: []string{"role-admin"},
		Enabled:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rule.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(rule.Thresholds) != 3 {
		t.Fatalf("expected 3 thresholds, got %d", len(rule.Thresholds))
	}
	if rule.Thresholds[0] != 80 || rule.Thresholds[1] != 90 || rule.Thresholds[2] != 100 {
		t.Fatalf("unexpected thresholds: %v", rule.Thresholds)
	}
}

func TestDisabledAlertRuleDoesNotTrigger(t *testing.T) {
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:        seed.IDDept3,
		NodeName:      "后端组",
		Thresholds:    []int{80, 90, 100},
		NotifyRoleIDs: []string{"role-admin"},
		Enabled:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Disable the rule
	updated, err := svc.UpdateAlert(ctx, rule.ID, types.AlertRule{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Enabled {
		t.Fatal("expected rule to be disabled")
	}

	// Verify the rule is persisted as disabled
	rules, err := svc.ListAlerts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		if r.ID == rule.ID && r.Enabled {
			t.Fatal("expected persisted rule to be disabled")
		}
	}
}

func TestUpdateAlertRuleThresholds(t *testing.T) {
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:     seed.IDDept3,
		NodeName:   "后端组",
		Thresholds: []int{80},
		Enabled:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Update to have multiple thresholds
	updated, err := svc.UpdateAlert(ctx, rule.ID, types.AlertRule{
		Thresholds: []int{80, 90, 100},
		Enabled:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Thresholds) != 3 {
		t.Fatalf("expected 3 thresholds after update, got %d", len(updated.Thresholds))
	}
}

func TestDeleteAlertRule(t *testing.T) {
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:     seed.IDDept3,
		NodeName:   "后端组",
		Thresholds: []int{80, 90, 100},
		Enabled:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteAlert(ctx, rule.ID); err != nil {
		t.Fatal(err)
	}

	rules, err := svc.ListAlerts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rules {
		if r.ID == rule.ID {
			t.Fatal("expected rule to be deleted")
		}
	}
}

// --- Overrun threshold notification tests ---

func TestOverrunDepartmentThresholdSendsNotification(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	ctx := testutil.Ctx()

	// Set consumed >= budget to trigger the 100% threshold
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 25000)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, err := json.Marshal(map[string]any{
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	// Verify notification was sent
	logs := testutil.NotificationLogs(st)
	if len(logs) == 0 {
		t.Fatal("expected notification log for overrun")
	}
	found := false
	for _, log := range logs {
		if log.EventType == types.NotificationEventOverrunBlocked {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected overrun_blocked notification event")
	}
}

func TestOverrunMemberThresholdSendsNotification(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	ctx := testutil.Ctx()

	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
	// Set member personal quota below used
	if err := st.Org().UpdateMemberPersonalQuota(ctx, seed.IDMember1, 100); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == seed.IDPlatformKey1 {
			keys[i].Used = 9999
			keys[i].Quota = 1000
		}
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"memberId":      seed.IDMember1,
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	logs := testutil.NotificationLogs(st)
	found := false
	for _, log := range logs {
		if log.EventType == types.NotificationEventOverrunBlocked {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected overrun_blocked notification for member quota breach")
	}
}

func TestOverrunDoesNotNotifyWhenBelowBudget(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	ctx := testutil.Ctx()

	// Keep consumed well below budget
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	testutil.SetDeptConsumed(t, tree, seed.IDDept3, 100)
	orgfix.PersistBudgetTreeT(t, ctx, st, tree)
	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())

	payload, err := json.Marshal(map[string]any{
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	logs := testutil.NotificationLogs(st)
	for _, log := range logs {
		if log.EventType == types.NotificationEventOverrunBlocked {
			t.Fatal("did not expect overrun notification when below budget")
		}
	}
	if stub.UpdateTokenCalls != 0 {
		t.Fatal("did not expect keys to be disabled when below budget")
	}
}

func TestOverrunBudgetGroupSendsNotification(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	overrun := budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
	ctx := testutil.Ctx()

	// Set up budget group at 100% consumed
	groups, err := st.Budget().Groups(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected budget groups in seed")
	}
	groupID := groups[0].ID
	groups[0].Consumed = groups[0].Budget
	if err := st.Budget().SetGroups(ctx, groups); err != nil {
		t.Fatal(err)
	}
	groupIDCopy := groupID
	tokenID := int64(99)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		DepartmentID:  seed.IDDept3,
		BudgetGroupID: &groupIDCopy,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "group-" + groupID,
	}); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"budgetGroupId": groupIDCopy,
		"departmentId":  seed.IDDept3,
		"platformKeyId": seed.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

	logs := testutil.NotificationLogs(st)
	found := false
	for _, log := range logs {
		if log.EventType == types.NotificationEventOverrunBlocked {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected overrun_blocked notification for budget group breach")
	}
}

// Ensure type assertion is used (suppress unused import)
var _ config.Config

