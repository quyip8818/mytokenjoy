package budget_test

import (
	"encoding/json"
	"log/slog"
	"testing"

	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestOverrunDisablesDepartmentKeys(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, 25000)

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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when department overrun")
	}
}

func TestOverrunSkipsWhenLifecycleDisabled(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	overrun := budget.NewOverrunService(cfg, st, nil, notification.NewService(cfg, st, slog.Default()), slog.Default())
	ctx := testutil.Ctx()

	payload, err := json.Marshal(map[string]any{"departmentId": contract.IDDept3})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
}

func TestOverrunMemberAxisWhenOverQuota(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == contract.IDPlatformKey1 {
			keys[i].Used = 9999
			keys[i].Quota = 1000
		}
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	if err := st.Org().UpdateMemberPersonalQuota(ctx, contract.IDMember1, 100); err != nil {
		t.Fatal(err)
	}

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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when member overrun")
	}
}

func TestOverrunBudgetGroupAxis(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	groups, err := st.Budget().Groups(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected budget group")
	}
	groupID := groups[0].ID
	groups[0].Consumed = groups[0].Budget
	if err := st.Budget().SetGroups(ctx, groups); err != nil {
		t.Fatal(err)
	}
	groupIDCopy := groupID
	tokenID := int64(99)
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		DepartmentID:  contract.IDDept3,
		BudgetGroupID: &groupIDCopy,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "group-" + groupID,
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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when budget group overrun")
	}
}
