package budget_test

import (
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
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

// --- fixture ---

type overrunFixture struct {
	stub     *mock.StubAdminClient
	notifier *testutil.RecordingNotifier
	overrun  *budget.OverrunService
	st       store.Store
}

func setupOverrun(t *testing.T) *overrunFixture {
	t.Helper()
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	notifier := &testutil.RecordingNotifier{}
	svc := budgetfix.NewOverrunService(t, cfg, st, stub, notifier)
	return &overrunFixture{stub: stub, notifier: notifier, overrun: svc, st: st}
}

// process marshals payload and calls ProcessOverrunPayload, failing on error.
func (f *overrunFixture) process(t *testing.T, payload map[string]any) {
	t.Helper()
	data, _ := json.Marshal(payload)
	if err := f.overrun.ProcessOverrunPayload(testutil.Ctx(), data); err != nil {
		t.Fatal(err)
	}
}

func (f *overrunFixture) assertKeyDisabled(t *testing.T) {
	t.Helper()
	if f.stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken (key disabled)")
	}
}

func (f *overrunFixture) assertKeyNotDisabled(t *testing.T) {
	t.Helper()
	if f.stub.UpdateTokenCalls != 0 {
		t.Fatal("did not expect keys to be disabled")
	}
}

func (f *overrunFixture) assertNotificationSent(t *testing.T) {
	t.Helper()
	for _, n := range f.notifier.Notifications {
		if n.EventType == types.NotificationEventOverrunBlocked {
			return
		}
	}
	t.Fatal("expected overrun_blocked notification")
}

func (f *overrunFixture) assertNoNotification(t *testing.T) {
	t.Helper()
	for _, n := range f.notifier.Notifications {
		if n.EventType == types.NotificationEventOverrunBlocked {
			t.Fatal("did not expect overrun notification")
		}
	}
}

// seedProjectOverrun sets up a project whose consumed > budget and wires
// platformKeyId to it via mapping. Returns the project ID for payload use.
func (f *overrunFixture) seedProjectOverrun(t *testing.T) uuid.UUID {
	t.Helper()
	ctx := testutil.Ctx()
	groups, err := f.st.Budget().Projects(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected project in seed")
	}
	gid := groups[0].ID
	budgetfix.SetProjectSnapshotConsumed(t, f.st, gid, groups[0].Budget+1)

	keys, err := f.st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range keys {
		if keys[i].ID == contract.IDPlatformKey1 {
			keys[i].ProjectID = &gid
			keys[i].Scope = types.PlatformKeyScopeProject
		}
	}
	if err := f.st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	tokenID := int64(99)
	if err := f.st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &tokenID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   "group-" + gid.String(),
	}); err != nil {
		t.Fatal(err)
	}
	return gid
}

// --- tests ---

func TestOverrunSkipsWhenLifecycleDisabled(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := budget.NewOverrunService(cfg, st, nil, notification.NewService(cfg, st, slog.Default()), slog.Default())

	data, _ := json.Marshal(map[string]any{"departmentId": contract.IDDept3})
	if err := svc.ProcessOverrunPayload(testutil.Ctx(), data); err != nil {
		t.Fatal(err)
	}
}

// PRD 3.2: 预警规则
// - "达到 100% 时阻断请求"
// - "通知发送失败不影响阻断逻辑"

func TestOverrunDepartmentAxis(t *testing.T) {
	t.Parallel()
	deptPayload := map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
	}

	t.Run("sends_notification_on_exhaustion", func(t *testing.T) {
		t.Parallel()
		fix := setupOverrun(t)
		budgetfix.SeedDeptOverrun(t, fix.st, contract.IDDept3, budgetfix.QuotaFromDisplay(25000))

		fix.process(t, deptPayload)
		fix.assertNotificationSent(t)
	})

	t.Run("does_not_fail_when_notification_fails", func(t *testing.T) {
		t.Parallel()
		stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
		cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
		svc := budgetfix.NewOverrunService(t, cfg, st, stub, &testutil.FailingNotifier{})
		budgetfix.SeedDeptOverrun(t, st, contract.IDDept3, budgetfix.QuotaFromDisplay(25000))

		data, _ := json.Marshal(deptPayload)
		if err := svc.ProcessOverrunPayload(testutil.Ctx(), data); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("skips_when_not_exhausted", func(t *testing.T) {
		t.Parallel()
		fix := setupOverrun(t)
		budgetfix.SeedDeptOverrun(t, fix.st, contract.IDDept3, 12500)

		fix.process(t, deptPayload)
		fix.assertKeyNotDisabled(t)
		fix.assertNoNotification(t)
	})
}

func TestOverrunMemberAxis(t *testing.T) {
	t.Parallel()

	t.Run("disables_key_and_sends_notification", func(t *testing.T) {
		t.Parallel()
		fix := setupOverrun(t)
		ctx := testutil.Ctx()

		newapisynctf.UpsertMapping(t, fix.st, newapisynctf.DefaultMappingOpts())
		if err := fix.st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, 100); err != nil {
			t.Fatal(err)
		}
		budgetfix.SetMemberSnapshotConsumed(t, fix.st, contract.IDMember1, 101)

		fix.process(t, map[string]any{
			"memberId":      contract.IDMember1,
			"departmentId":  contract.IDDept3,
			"platformKeyId": contract.IDPlatformKey1,
		})
		fix.assertKeyDisabled(t)
		fix.assertNotificationSent(t)
	})
}

func TestOverrunProjectAxis(t *testing.T) {
	t.Parallel()

	t.Run("disables_key_and_sends_notification", func(t *testing.T) {
		t.Parallel()
		fix := setupOverrun(t)
		gid := fix.seedProjectOverrun(t)

		fix.process(t, map[string]any{
			"projectId":     gid,
			"departmentId":  contract.IDDept3,
			"platformKeyId": contract.IDPlatformKey1,
		})
		fix.assertKeyDisabled(t)
		fix.assertNotificationSent(t)
	})
}

func TestOverrunPlatformKeyAxis(t *testing.T) {
	t.Parallel()
	fix := setupOverrun(t)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, fix.st, newapisynctf.DefaultMappingOpts())
	keys, err := fix.st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var keyBudget int64
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			keyBudget = key.Budget
			break
		}
	}
	if keyBudget <= 0 {
		t.Fatal("expected plk-1 to have positive budget")
	}
	budgetfix.SetPlatformKeySnapshotConsumed(t, fix.st, contract.IDPlatformKey1, keyBudget+1)

	fix.process(t, map[string]any{
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey1,
	})
	fix.assertKeyDisabled(t)
}

func TestOverrunProjectMemberAxis(t *testing.T) {
	t.Parallel()
	fix := setupOverrun(t)
	ctx := testutil.Ctx()

	projects, err := fix.st.Budget().Projects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	memberBudget := int64(100)
	for i := range projects {
		if projects[i].ID == contract.IDProject1 {
			if projects[i].MemberBudgets == nil {
				projects[i].MemberBudgets = make(map[uuid.UUID]int64)
			}
			projects[i].MemberBudgets[contract.IDMember1] = memberBudget
			break
		}
	}
	if err := fix.st.Budget().SetProjects(ctx, projects); err != nil {
		t.Fatal(err)
	}
	budgetfix.SetPlatformKeySnapshotConsumed(t, fix.st, contract.IDPlatformKey6, memberBudget+1)

	projectID := contract.IDProject1
	memberID := contract.IDMember1
	tokenID := int64(88)
	if err := fix.st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey6,
		NewAPIKeyID:   &tokenID,
		MemberID:      &memberID,
		ProjectID:     &projectID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   "group-" + projectID.String(),
	}); err != nil {
		t.Fatal(err)
	}

	fix.process(t, map[string]any{
		"memberId":      contract.IDMember1,
		"projectId":     contract.IDProject1,
		"departmentId":  contract.IDDept3,
		"platformKeyId": contract.IDPlatformKey6,
	})
	fix.assertKeyDisabled(t)
}

func TestOverrunSkipsMemberAxisWhenProjectPresent(t *testing.T) {
	t.Parallel()
	fix := setupOverrun(t)
	ctx := testutil.Ctx()

	newapisynctf.UpsertMapping(t, fix.st, newapisynctf.DefaultMappingOpts())
	if err := fix.st.Org().UpdateMemberPersonalBudget(ctx, contract.IDMember1, 100); err != nil {
		t.Fatal(err)
	}
	budgetfix.SetMemberSnapshotConsumed(t, fix.st, contract.IDMember1, 101)

	groups, err := fix.st.Budget().Projects(ctx)
	if err != nil || len(groups) == 0 {
		t.Fatal("expected project")
	}
	groupID := groups[0].ID
	budgetfix.SetProjectSnapshotConsumed(t, fix.st, groupID, 0)

	fix.process(t, map[string]any{
		"memberId":     contract.IDMember1,
		"projectId":    groupID,
		"departmentId": contract.IDDept3,
	})
	fix.assertKeyNotDisabled(t)
}
