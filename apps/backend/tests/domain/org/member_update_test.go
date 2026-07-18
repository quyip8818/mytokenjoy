package org_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestUpdateMemberMergeSemanticsPreservesUnchangedFields(t *testing.T) {
	t.Parallel()
	svc, st := newTestOrgServiceWithStore(t)
	ctx := testutil.Ctx()

	// Verify initial state of seeded member1
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var original types.Member
	for _, m := range members {
		if m.ID == contract.IDMember1 {
			original = m
			break
		}
	}
	if original.ID == uuid.Nil {
		t.Fatal("seed member1 not found")
	}
	// Sanity: member1 has roles, source, companyID set
	if len(original.Roles) == 0 {
		t.Fatal("expected seeded member to have roles")
	}
	if original.Source == "" {
		t.Fatal("expected seeded member to have source")
	}
	if original.CompanyID == uuid.Nil {
		t.Fatal("expected seeded member to have companyID")
	}

	// Update only name and phone (partial update from frontend)
	updated, err := svc.UpdateMember(ctx, contract.IDMember1, types.Member{
		Name:  "张三改名",
		Phone: "13999999999",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Name and phone should be updated
	if updated.Name != "张三改名" {
		t.Fatalf("expected name '张三改名', got %q", updated.Name)
	}
	if updated.Phone != "13999999999" {
		t.Fatalf("expected phone '13999999999', got %q", updated.Phone)
	}

	// Roles, Status, Source, CompanyID must be preserved
	if !slices.Equal(updated.Roles, original.Roles) {
		t.Fatalf("roles lost: expected %v, got %v", original.Roles, updated.Roles)
	}
	if updated.Status != original.Status {
		t.Fatalf("status lost: expected %q, got %q", original.Status, updated.Status)
	}
	if updated.Source != original.Source {
		t.Fatalf("source lost: expected %q, got %q", original.Source, updated.Source)
	}
	if updated.CompanyID != original.CompanyID {
		t.Fatalf("companyID lost: expected %d, got %d", original.CompanyID, updated.CompanyID)
	}
	if updated.Email != original.Email {
		t.Fatalf("email lost: expected %q, got %q", original.Email, updated.Email)
	}
}

func TestUpdateMemberRolesUpdatedWhenProvided(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	newRoles := []string{permission.RoleMember}
	updated, err := svc.UpdateMember(ctx, contract.IDMember1, types.Member{
		Roles: newRoles,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated.Roles, newRoles) {
		t.Fatalf("expected roles %v, got %v", newRoles, updated.Roles)
	}
}
