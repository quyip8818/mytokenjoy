package org_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSyncThresholdBlocksDeletion(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	importedExternalID := "ou-gone"
	members := env.Store.Org().Members()
	members = append(members, types.Member{
		ID: "m-feishu-ou-gone", Name: "Gone User", DepartmentID: seed.IDDept3, DepartmentName: "研发部",
		Status: "active", Roles: []string{"普通成员"}, Source: types.MemberSourceImported, ExternalID: &importedExternalID,
	})
	_ = env.Store.Org().SetMembers(members)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, StartTime: "00:00", FrequencyHours: 1,
		DeleteMemberThreshold: 0, DeleteDepartmentThreshold: 5,
	})
	env.Cfg.FeishuBaseURL = env.ServerURL
	env.Svc = testutil.NewOrgService(t, env.Cfg, env.Store)

	_, err := env.Svc.TriggerSync(context.Background())
	if err == nil {
		t.Fatal("expected threshold error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %v", err)
	}
}

func TestSyncRenamesBudgetAndRouting(t *testing.T) {
	deptName := "Mock Dept"
	server := testutil.StartMutableFeishuServer(t, &deptName, testutil.DefaultFeishuUsers())
	env := testutil.SetupImportedFeishuOrgWithServer(t, server.URL)
	deptName = "Renamed Dept"
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if _, err := env.Svc.TriggerSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	dept := orgutil.FindDepartment(env.Store.Org().Departments(), seed.IDFeishuDept1)
	if dept == nil || dept.Name != "Renamed Dept" {
		t.Fatalf("expected renamed department, got %+v", dept)
	}
	node := budgetutil.FindBudgetNode(env.Store.Budget().Tree(), seed.IDFeishuDept1)
	if node == nil || node.Name != "Renamed Dept" {
		t.Fatalf("expected renamed budget node, got %+v", node)
	}
	for _, rule := range env.Store.Models().RoutingRules() {
		if rule.NodeID == seed.IDFeishuDept1 && rule.NodeName != "Renamed Dept" {
			t.Fatalf("expected renamed routing rule, got %+v", rule)
		}
	}
}

func TestSyncSoftDeletesBelowThreshold(t *testing.T) {
	env := testutil.SetupImportedFeishuOrg(t)
	externalID := "ou-gone"
	members := env.Store.Org().Members()
	members = append(members, types.Member{
		ID: "m-feishu-ou-gone", Name: "Gone User", DepartmentID: seed.IDDept3,
		Status: "active", Roles: []string{"普通成员"}, Source: types.MemberSourceImported, ExternalID: &externalID,
	})
	_ = env.Store.Org().SetMembers(members)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if _, err := env.Svc.TriggerSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, member := range env.Store.Org().Members() {
		if member.ID == "m-feishu-ou-gone" && member.Status != "inactive" {
			t.Fatalf("expected soft-deleted member, got status %s", member.Status)
		}
	}
}

func TestSyncSkipsManualDepartmentDeletion(t *testing.T) {
	env := testutil.SetupImportedFeishuOrg(t)
	manual := types.DeptSourceManual
	departments := env.Store.Org().Departments()
	for i := range departments {
		for j := range departments[i].Children {
			if departments[i].Children[j].ID == seed.IDDept2 {
				departments[i].Children[j].ExternalID = testutil.StrPtr("od-manual")
				departments[i].Children[j].Source = &manual
			}
		}
	}
	_ = env.Store.Org().SetDepartments(departments)
	env = testutil.WithSyncConfig(t, env, types.SyncConfig{
		Enabled: true, DeleteMemberThreshold: 10, DeleteDepartmentThreshold: 5,
	})

	if _, err := env.Svc.TriggerSync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if orgutil.FindDepartment(env.Store.Org().Departments(), seed.IDDept2) == nil {
		t.Fatal("manual department should remain after sync")
	}
}
