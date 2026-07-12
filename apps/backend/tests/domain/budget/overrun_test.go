package budget_test

import (
	"encoding/json"
	"log/slog"
	"testing"

	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

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

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	if err := st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, 100); err != nil {
		t.Fatal(err)
	}
	budgetfix.SetMemberSnapshotConsumed(t, st, contract.IDMember1, 100.01)

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

func TestOverrunProjectAxis(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	groups, err := st.Budget().Projects(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected project")
	}
	groupID := groups[0].ID
	budgetfix.SetProjectSnapshotConsumed(t, st, groupID, groups[0].Budget+0.01)
	groupIDCopy := groupID
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == contract.IDPlatformKey1 {
			keys[i].ProjectID = &groupIDCopy
			keys[i].Scope = types.PlatformKeyScopeProject
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
		"projectId":     groupIDCopy,
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
		t.Fatal("expected UpdateToken when project overrun")
	}
}

func TestOverrunPlatformKeyAxisWhenOverQuota(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var keyBudget float64
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			keyBudget = key.Budget
			break
		}
	}
	if keyBudget <= 0 {
		t.Fatal("expected plk-1 to have positive budget")
	}
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, contract.IDPlatformKey1, keyBudget+0.01)

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
		t.Fatal("expected UpdateToken when platform key overrun")
	}
}

func TestOverrunProjectMemberAxisWhenOverQuota(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	projects, err := st.Budget().Projects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	memberBudget := 100.0
	for i := range projects {
		if projects[i].ID == contract.IDProject1 {
			if projects[i].MemberBudgets == nil {
				projects[i].MemberBudgets = make(map[string]float64)
			}
			projects[i].MemberBudgets[contract.IDMember1] = memberBudget
			break
		}
	}
	if err := st.Budget().SetProjects(ctx, projects); err != nil {
		t.Fatal(err)
	}
	budgetfix.SetPlatformKeySnapshotConsumed(t, st, "plk-bg-1", memberBudget+0.01)

	projectID := contract.IDProject1
	memberID := contract.IDMember1
	tokenID := int64(88)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: "plk-bg-1",
		NewAPIKeyID:   &tokenID,
		MemberID:      &memberID,
		ProjectID:     &projectID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   "group-" + projectID,
	}); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"memberId":      contract.IDMember1,
		"projectId":     contract.IDProject1,
		"departmentId":  contract.IDDept3,
		"platformKeyId": "plk-bg-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when project_member sub-budget overrun")
	}
}

func TestOverrunSkipsMemberAxisWhenProjectPresent(t *testing.T) {
	t.Parallel()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	overrun := budgetfix.NewOverrunService(t, cfg, st, stub, nil)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, st, newapisynctf.DefaultMappingOpts())
	if err := st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, 100); err != nil {
		t.Fatal(err)
	}
	budgetfix.SetMemberSnapshotConsumed(t, st, contract.IDMember1, 100.01)

	groups, err := st.Budget().Projects(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected project")
	}
	groupID := groups[0].ID
	budgetfix.SetProjectSnapshotConsumed(t, st, groupID, 0)

	payload, err := json.Marshal(map[string]any{
		"memberId":     contract.IDMember1,
		"projectId":    groupID,
		"departmentId": contract.IDDept3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
	if stub.UpdateTokenCalls != 0 {
		t.Fatal("expected member axis to be skipped when project is present")
	}
}
