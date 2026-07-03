package budget_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget"
	relay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newOverrunService(t *testing.T, cfg config.Config, st store.Store, stub *mock.StubAdminClient) *budget.OverrunService {
	t.Helper()
	lifecycle := relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg))
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifier := notification.NewService(cfg, st, logger)
	return budget.NewOverrunService(cfg, st, lifecycle, notifier, logger)
}

func TestOverrunDisablesDepartmentKeys(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	overrun := budget.NewOverrunService(cfg, st, relay.NewTokenLifecycle(cfg, st, stub, nil, relay.NewChannelPolicy(cfg)), notification.NewService(cfg, st, slog.Default()), slog.Default())
	ctx := testutil.Ctx()

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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when department overrun")
	}
}

func TestOverrunSkipsWhenLifecycleDisabled(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	overrun := budget.NewOverrunService(cfg, st, nil, notification.NewService(cfg, st, slog.Default()), slog.Default())
	ctx := testutil.Ctx()

	payload, err := json.Marshal(map[string]any{"departmentId": seed.IDDept3})
	if err != nil {
		t.Fatal(err)
	}
	if err := overrun.ProcessOverrunPayload(ctx, payload); err != nil {
		t.Fatal(err)
	}
}

func TestOverrunMemberAxisWhenOverQuota(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	overrun := newOverrunService(t, cfg, st, stub)
	ctx := testutil.Ctx()

	relayfix.UpsertMapping(t, st, relayfix.DefaultMappingOpts())
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
	if err := st.Org().UpdateMemberPersonalQuota(ctx, seed.IDMember1, 100); err != nil {
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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when member overrun")
	}
}

func TestOverrunBudgetGroupAxis(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	cfg, st := testutil.NewMemoryStoreFromConfig(t, testutil.WithNewAPIEnabled(true))
	overrun := newOverrunService(t, cfg, st, stub)
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
	if stub.UpdateTokenCalls == 0 {
		t.Fatal("expected UpdateToken when budget group overrun")
	}
}
