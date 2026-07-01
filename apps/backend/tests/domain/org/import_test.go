package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestImportCreatesDepartmentsAndMembers(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	result := testutil.ImportFeishuOrg(t, env)
	if result.SuccessDepartments < 1 || result.SuccessMembers < 1 {
		t.Fatalf("unexpected result %+v", result)
	}
	departments, err := env.Store.Org().Departments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pkgorg.FindDepartment(departments, seed.IDFeishuDept1) == nil {
		t.Fatal("expected imported department in tree")
	}
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	foundMember := false
	for _, member := range members {
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
	ctx := testutil.Ctx()
	manual := types.DeptSourceManual
	departments, err := env.Store.Org().Departments(ctx)
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
	if err := env.Store.Org().SetDepartments(ctx, departments); err != nil {
		t.Fatal(err)
	}
	testutil.ImportFeishuOrg(t, env)

	departments, err = env.Store.Org().Departments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	dept := pkgorg.FindDepartment(departments, seed.IDDept2)
	if dept == nil || dept.Name != "技术部" {
		t.Fatalf("manual department should keep name, got %+v", dept)
	}
}

func TestImportSecondRunIdempotent(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	first := testutil.ImportFeishuOrg(t, env)
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeMembers := len(members)
	second := testutil.ImportFeishuOrg(t, env)
	members, err = env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != beforeMembers {
		t.Fatalf("expected member count unchanged, before=%d after=%d", beforeMembers, len(members))
	}
	if second.SuccessMembers > first.SuccessMembers {
		t.Fatalf("expected no duplicate member imports, first=%+v second=%+v", first, second)
	}
}

func TestImportProvisionsBudgetAndRouting(t *testing.T) {
	env := testutil.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	testutil.ImportFeishuOrg(t, env)
	budgetTree, err := env.Store.Budget().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if budget.FindBudgetNode(budgetTree, seed.IDFeishuDept1) == nil {
		t.Fatal("expected budget node for imported department")
	}
	rules, err := env.Store.Models().RoutingRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	foundRule := false
	for _, rule := range rules {
		if rule.NodeID == seed.IDFeishuDept1 {
			foundRule = true
			break
		}
	}
	if !foundRule {
		t.Fatal("expected routing rule for imported department")
	}
}
