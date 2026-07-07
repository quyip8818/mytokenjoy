package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateAlertRuleWithMultipleThresholds(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:        contract.IDDept3,
		NodeName:      "后端组",
		Thresholds:    []int{80, 90, 100},
		NotifyRoleIDs: []string{"role-1"},
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
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:        contract.IDDept3,
		NodeName:      "后端组",
		Thresholds:    []int{80, 90, 100},
		NotifyRoleIDs: []string{"role-1"},
		Enabled:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateAlert(ctx, rule.ID, types.AlertRule{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Enabled {
		t.Fatal("expected rule to be disabled")
	}

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
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:     contract.IDDept3,
		NodeName:   "后端组",
		Thresholds: []int{80},
		Enabled:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

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
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()

	rule, err := svc.CreateAlert(ctx, types.AlertRule{
		NodeID:     contract.IDDept3,
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
