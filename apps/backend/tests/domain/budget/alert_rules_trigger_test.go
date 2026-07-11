package budget_test

import (
	"encoding/json"
	"testing"

	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestOverrunDepartmentThresholdSendsNotification(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, testutil.DisplayPoints(25000))

	payload, err := json.Marshal(map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}

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
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	if err := st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, 100); err != nil {
		t.Fatal(err)
	}
	testutil.SetMemberSnapshotConsumed(t, st, contract.IDMember1, 100.01)

	payload, err := json.Marshal(map[string]any{
		"memberId":      contract.IDMember1,
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
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
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, 100)

	payload, err := json.Marshal(map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
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
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	groups, err := st.Budget().Groups(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected budget groups in seed")
	}
	groupID := groups[0].ID
	testutil.SetGroupSnapshotConsumed(t, st, groupID, groups[0].Budget+0.01)
	groupIDCopy := groupID
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == contract.IDPlatformKey1 {
			keys[i].BudgetGroupID = &groupIDCopy
		}
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	tokenID := int64(99)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   "group-" + groupID,
	}); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"budgetGroupId": groupIDCopy,
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
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
