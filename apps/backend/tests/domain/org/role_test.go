package org_test

import (
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpdateRolePersistsAndBumpsAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	role, err := svc.CreateRole(ctx, "Editable", []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	before, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateRole(ctx, role.ID, "Renamed Role", []string{"p-1", "p-2"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed Role" || len(updated.Permissions) != 2 {
		t.Fatalf("unexpected role %+v", updated)
	}
	after, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision bump, before=%d after=%d", before, after)
	}
}

func TestUpdateRolePresetNotFound(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.UpdateRole(testutil.Ctx(), "missing-role", "X", []string{"p-1"})
	if err == nil {
		t.Fatal("expected error for missing role")
	}
}
