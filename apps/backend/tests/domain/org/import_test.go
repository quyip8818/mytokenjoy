package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestImportCreatesDepartmentsAndMembers(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	result := testutil.ImportFeishuOrg(t, env)
	if result.SuccessDepartments < 1 || result.SuccessMembers < 1 {
		t.Fatalf("unexpected result %+v", result)
	}
	if orgutil.FindDepartment(env.Store.Org().Departments(), seed.IDFeishuDept1) == nil {
		t.Fatal("expected imported department in tree")
	}
	foundMember := false
	for _, member := range env.Store.Org().Members() {
		if member.ExternalID != nil && *member.ExternalID == seed.IDFeishuExtUser1 {
			foundMember = true
			break
		}
	}
	if !foundMember {
		t.Fatal("expected imported member")
	}
}

func TestImportDoesNotOverwriteManualDepartment(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
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
	testutil.ImportFeishuOrg(t, env)

	dept := orgutil.FindDepartment(env.Store.Org().Departments(), seed.IDDept2)
	if dept == nil || dept.Name != "技术部" {
		t.Fatalf("manual department should keep name, got %+v", dept)
	}
}

func TestImportSecondRunIdempotent(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	first := testutil.ImportFeishuOrg(t, env)
	beforeMembers := len(env.Store.Org().Members())
	second := testutil.ImportFeishuOrg(t, env)
	if len(env.Store.Org().Members()) != beforeMembers {
		t.Fatalf("expected member count unchanged, before=%d after=%d", beforeMembers, len(env.Store.Org().Members()))
	}
	if second.SuccessMembers > first.SuccessMembers {
		t.Fatalf("expected no duplicate member imports, first=%+v second=%+v", first, second)
	}
}

func TestImportProvisionsBudgetAndRouting(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	testutil.ImportFeishuOrg(t, env)
	if budgetutil.FindBudgetNode(env.Store.Budget().Tree(), seed.IDFeishuDept1) == nil {
		t.Fatal("expected budget node for imported department")
	}
	foundRule := false
	for _, rule := range env.Store.Models().RoutingRules() {
		if rule.NodeID == seed.IDFeishuDept1 {
			foundRule = true
			break
		}
	}
	if !foundRule {
		t.Fatal("expected routing rule for imported department")
	}
}
