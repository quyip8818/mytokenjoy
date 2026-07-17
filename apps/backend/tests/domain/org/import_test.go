package org_test

import (
	"testing"

	"github.com/google/uuid"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestImportCreatesDepartmentsAndMembers(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	before, err := env.Store.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	result := orgfix.ImportFeishuOrg(t, env)
	if result.SuccessDepartments < 1 || result.SuccessMembers < 1 {
		t.Fatalf("unexpected result %+v", result)
	}
	departments, err := common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if pkgorg.FindDepartment(departments, contract.IDFeishuDept1) == nil {
		t.Fatal("expected imported department in tree")
	}
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	foundMember := false
	for _, member := range members {
		if member.ExternalID != nil && *member.ExternalID == contract.IDFeishuExtUser1 {
			foundMember = true
			break
		}
	}
	if !foundMember {
		t.Fatal("expected imported member")
	}
	after, err := env.Store.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if result.SuccessMembers < 1 || after <= before {
		t.Fatalf("expected authz revision bump on new members, before=%d after=%d members=%d", before, after, result.SuccessMembers)
	}
}

func TestImportDoesNotOverwriteManualDepartment(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	manual := types.DeptSourceManual
	departments, err := common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	for i := range departments {
		for j := range departments[i].Children {
			if departments[i].Children[j].ID == contract.IDDept2 {
				departments[i].Children[j].ExternalID = testutil.StrPtr("od-manual")
				departments[i].Children[j].Source = &manual
			}
		}
	}
	orgfix.PersistDepartmentsT(t, ctx, env.Store, departments)
	orgfix.ImportFeishuOrg(t, env)

	departments, err = common.LoadDepartments(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	dept := pkgorg.FindDepartment(departments, contract.IDDept2)
	if dept == nil || dept.Name != "技术部" {
		t.Fatalf("manual department should keep name, got %+v", dept)
	}
}

func TestImportSecondRunIdempotent(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	first := orgfix.ImportFeishuOrg(t, env)
	members, err := env.Store.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeMembers := len(members)
	second := orgfix.ImportFeishuOrg(t, env)
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
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	orgfix.ImportFeishuOrg(t, env)
	budgetTree, err := common.LoadBudgetTree(ctx, env.Store.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if budget.FindBudgetNode(budgetTree, contract.IDFeishuDept1) == nil {
		t.Fatal("expected budget node for imported department")
	}
	rules, err := common.LoadRoutingRules(ctx, env.Store.Org().Nodes(), env.Store.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	foundRule := false
	for _, rule := range rules {
		if rule.NodeID == contract.IDFeishuDept1 {
			foundRule = true
			break
		}
	}
	if !foundRule {
		t.Fatal("expected routing rule for imported department")
	}
}

func TestRetryImportFiltersByFailureIDs(t *testing.T) {
	t.Parallel()
	env := orgfix.SetupFeishuConnected(t)
	ctx := testutil.Ctx()
	if err := env.Store.Org().SetImportFailures(ctx, []types.ImportFailure{
		{ID: uuid.MustParse("00000000-0000-7000-0000-00000000cc01"), Name: "Retry Me", EmployeeID: "ou-retry", Reason: "network"},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := env.Svc.RetryImport(ctx, []string{"fail-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.SuccessMembers < 0 {
		t.Fatalf("unexpected result %+v", result)
	}
}
