package org_test

import (
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateMemberPersists(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	member, err := svc.CreateMember(ctx, types.Member{
		Name: "Phase0 User", Phone: "13900001111", Email: "phase0@example.com",
		DepartmentID: "dept-5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.ID == "" {
		t.Fatal("expected member id")
	}

	page, err := svc.ListMembers(ctx, "dept-5", "", true, 1, 200)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range page.Items {
		if item.ID == member.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created member not found in list")
	}
}

func TestCreateMemberUnknownDepartment404(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	_, err := svc.CreateMember(testutil.Ctx(), types.Member{
		Name: "Ghost", Phone: "13900002222", Email: "ghost@example.com",
		DepartmentID: "missing-dept",
	})
	asDomainError(t, err)
}

func TestDeleteMembersRejectsSelf(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()
	err := svc.DeleteMembers(ctx, []string{contract.IDMember1}, contract.IDMember1)
	de := asDomainError(t, err)
	if de.Status != domain.StatusBadRequest {
		t.Fatalf("expected 400, got %d", de.Status)
	}
}

func TestDeleteMembersDisablesKeys(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	if err := svc.DeleteMembers(testutil.Ctx(), []string{contract.IDMember1}, ""); err != nil {
		t.Fatal(err)
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if member.ID == contract.IDMember1 && member.Status != "inactive" {
			t.Fatalf("expected inactive status, got %s", member.Status)
		}
	}

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == contract.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}

func TestUpdateMemberStatusDisablesKeys(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	if err := svc.UpdateMemberStatus(testutil.Ctx(), []string{contract.IDMember1}, "inactive"); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == contract.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}

func TestListMembersDirectOnly(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	ctx := testutil.Ctx()

	allPage, err := svc.ListMembers(ctx, "dept-2", "", false, 1, 200)
	if err != nil {
		t.Fatal(err)
	}
	directPage, err := svc.ListMembers(ctx, "dept-2", "", true, 1, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(directPage.Items) >= len(allPage.Items) {
		t.Fatalf("directOnly should return fewer members: direct=%d all=%d", len(directPage.Items), len(allPage.Items))
	}
	if directPage.Total >= allPage.Total {
		t.Fatalf("directOnly total should be smaller: direct=%d all=%d", directPage.Total, allPage.Total)
	}
}
