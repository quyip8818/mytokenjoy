package org_test

import (
	"testing"

	"github.com/google/uuid"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateMemberBumpsAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMember(ctx, types.CreateMemberInput{
		User:   types.CreateMemberUserInput{Name: "Revision User", Phone: "13900003333", Email: "revision@example.com"},
		Member: types.CreateMemberData{DepartmentID: contract.IDDept5},
	}, uuid.Nil); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase, before=%d after=%d", before, after)
	}
}

func TestUpdateMemberRolesBumpsAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateMember(ctx, contract.IDMember1, types.UpdateMemberInput{
		Roles: []string{"预算审批员"},
	}); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase after role change, before=%d after=%d", before, after)
	}
}

func TestBatchImportBumpsAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.BatchImport(ctx, []types.BatchImportRow{
		{Name: "CSV User", Phone: "13900004444", Email: "csv@example.com", DepartmentName: "测试组"},
	}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d", result.Imported)
	}
	after, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase, before=%d after=%d", before, after)
	}
}
