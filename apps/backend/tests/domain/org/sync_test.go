package org_test

import (
	"errors"
	"fmt"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSyncThresholdBlocksDeletion(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	importedExternalID := "ou-gone"
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	members = append(members, types.Member{
		ID: "m-feishu-ou-gone", Name: "Gone User", DepartmentID: seed.IDDept3, DepartmentName: "研发部",
		Status: "active", Roles: []string{"普通成员"}, Source: types.MemberSourceImported, ExternalID: &importedExternalID,
	})
	if err := env.Store.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 0, DeleteDepartmentThreshold: 5,
	})
	env.Cfg.FeishuBaseURL = env.ServerURL
	env.Svc = orgfix.NewService(t, env.Cfg, env.Store)

	_, err = env.Svc.TriggerSync(testutil.Ctx())
	if err == nil {
		t.Fatal("expected threshold error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %v", err)
	}
}

func TestSyncRenamesBudgetAndRouting(t *testing.T) {
	t.Parallel()
	deptName := "Mock Dept"
	server := testutil.StartMutableFeishuServer(t, &deptName, testutil.DefaultFeishuUsers())
	env := orgfix.SetupImportedFeishuOrgWithServer(t, server.URL)
	ctx := testutil.Ctx()
	deptName = "Renamed Dept"
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if _, err := env.Svc.TriggerSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	departments, err := common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	dept := pkgorg.FindDepartment(departments, seed.IDFeishuDept1)
	if dept == nil || dept.Name != "Renamed Dept" {
		t.Fatalf("expected renamed department, got %+v", dept)
	}
	budgetTree, err := common.LoadBudgetTree(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	node := budget.FindBudgetNode(budgetTree, seed.IDFeishuDept1)
	if node == nil || node.Name != "Renamed Dept" {
		t.Fatalf("expected renamed budget node, got %+v", node)
	}
	rules, err := common.LoadRoutingRules(ctx, env.Store.Org().Nodes(), env.Store.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		if rule.NodeID == seed.IDFeishuDept1 && rule.NodeName != "Renamed Dept" {
			t.Fatalf("expected renamed routing rule, got %+v", rule)
		}
	}
}

func TestSyncSoftDeletesBelowThreshold(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupImportedFeishuOrg(t)
	ctx := testutil.Ctx()
	externalID := "ou-gone"
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	members = append(members, types.Member{
		ID: "m-feishu-ou-gone", Name: "Gone User", DepartmentID: seed.IDDept3,
		Status: "active", Roles: []string{"普通成员"}, Source: types.MemberSourceImported, ExternalID: &externalID,
	})
	if err := env.Store.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	before, err := env.Store.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.Svc.TriggerSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	after, err := env.Store.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase after sync soft-delete, before=%d after=%d", before, after)
	}
	members, err = env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if member.ID == "m-feishu-ou-gone" && member.Status != "inactive" {
			t.Fatalf("expected soft-deleted member, got status %s", member.Status)
		}
	}
}

func TestSyncSkipsManualDepartmentDeletion(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupImportedFeishuOrg(t)
	ctx := testutil.Ctx()
	manual := types.DeptSourceManual
	departments, err := common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	for i := range departments {
		for j := range departments[i].Children {
			if departments[i].Children[j].ID == seed.IDDept2 {
				departments[i].Children[j].ExternalID = testutil.StrPtr("od-manual")
				departments[i].Children[j].Source = &manual
			}
		}
	}
	orgfix.PersistDepartmentsT(t, ctx, env.Store, departments)
	env = orgfix.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if _, err := env.Svc.TriggerSync(testutil.Ctx()); err != nil {
		t.Fatal(err)
	}
	departments, err = common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if pkgorg.FindDepartment(departments, seed.IDDept2) == nil {
		t.Fatal("manual department should remain after sync")
	}
}

func TestSyncConfigRoundTrip(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	cfg := types.SyncConfig{
		Enabled: true, StartTime: "02:30", FrequencyHours: 4,
		DeleteMemberThreshold: 3, DeleteDepartmentThreshold: 2,
	}
	if err := svc.UpdateSyncConfig(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetSyncConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.StartTime != "02:30" || got.FrequencyHours != 4 {
		t.Fatalf("unexpected config %+v", got)
	}
	if got.DeleteMemberThreshold != 3 || got.DeleteDepartmentThreshold != 2 {
		t.Fatalf("unexpected thresholds %+v", got)
	}
}

func TestListSyncLogsPagination(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	for i := 0; i < 3; i++ {
		if err := st.Org().AppendSyncLog(ctx, types.SyncLog{
			ID: fmt.Sprintf("log-%d", i), Time: "2026-06-01 10:00",
			Type: types.SyncTypeManual, Result: types.SyncResultSuccess, Detail: "ok",
		}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := svc.ListSyncLogs(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Total < 3 {
		t.Fatalf("unexpected page %+v", page)
	}
}
