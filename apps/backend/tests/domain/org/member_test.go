package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateMemberPersists(t *testing.T) {
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
	svc := newTestOrgService(t)
	_, err := svc.CreateMember(testutil.Ctx(), types.Member{
		Name: "Ghost", Phone: "13900002222", Email: "ghost@example.com",
		DepartmentID: "missing-dept",
	})
	asDomainError(t, err)
}

func TestDeleteMembersDisablesKeys(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	if err := svc.DeleteMembers(testutil.Ctx(), []string{seed.IDMember1}); err != nil {
		t.Fatal(err)
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if member.ID == seed.IDMember1 && member.Status != "inactive" {
			t.Fatalf("expected inactive status, got %s", member.Status)
		}
	}

	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == seed.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}

func TestTransferMembersUpdatesRelayMapping(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	memberID := seed.IDMember1
	targetDept := "dept-4"
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		MemberID:      &memberID,
		DepartmentID:  seed.IDDept3,
		RelayGroup:    "default",
		SyncStatus:    store.RelaySyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}

	if err := svc.TransferMembers(testutil.Ctx(), []string{memberID}, targetDept); err != nil {
		t.Fatal(err)
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if member.ID != memberID {
			continue
		}
		if member.DepartmentID != targetDept {
			t.Fatalf("expected department %s, got %s", targetDept, member.DepartmentID)
		}
	}

	mappings, err := st.Relay().ListMappingsByMemberID(ctx, memberID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) == 0 {
		t.Fatal("expected relay mapping")
	}
	if mappings[0].DepartmentID != targetDept {
		t.Fatalf("expected mapping department %s, got %s", targetDept, mappings[0].DepartmentID)
	}
}

func TestUpdateMemberStatusDisablesKeys(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	if err := svc.UpdateMemberStatus(testutil.Ctx(), []string{seed.IDMember1}, "inactive"); err != nil {
		t.Fatal(err)
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == seed.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}

func TestCreateMemberBumpsAuthzRevision(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateMember(ctx, types.Member{
		Name: "Revision User", Phone: "13900003333", Email: "revision@example.com",
		DepartmentID: "dept-5",
	}); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase, before=%d after=%d", before, after)
	}
}

func TestUpdateMemberRolesBumpsAuthzRevision(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var target types.Member
	for _, member := range members {
		if member.ID == seed.IDMember1 {
			target = member
			break
		}
	}
	target.Roles = []string{"组织管理员"}
	if _, err := svc.UpdateMember(ctx, seed.IDMember1, target); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase after role change, before=%d after=%d", before, after)
	}
}

func TestBatchImportBumpsAuthzRevision(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.BatchImport(ctx, []types.BatchImportRow{
		{Name: "CSV User", Phone: "13900004444", Email: "csv@example.com", DepartmentName: "测试组"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected 1 imported, got %d", result.Imported)
	}
	after, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after <= before {
		t.Fatalf("expected authz revision to increase, before=%d after=%d", before, after)
	}
}
