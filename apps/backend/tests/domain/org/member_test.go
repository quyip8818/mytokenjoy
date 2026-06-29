package org_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateMemberPersists(t *testing.T) {
	svc := newTestOrgService(t)

	member, err := svc.CreateMember(org.Member{
		Name: "Phase0 User", Phone: "13900001111", Email: "phase0@example.com",
		DepartmentID: "dept-5",
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.ID == "" {
		t.Fatal("expected member id")
	}

	found := false
	for _, item := range svc.ListMembers("dept-5", "", true, 1, 200).Items {
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
	_, err := svc.CreateMember(org.Member{
		Name: "Ghost", Phone: "13900002222", Email: "ghost@example.com",
		DepartmentID: "missing-dept",
	})
	asDomainError(t, err)
}

func TestDeleteMembersDisablesKeys(t *testing.T) {
	svc := newTestOrgService(t)
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc = testutil.NewOrgService(t, cfg, st)

	if err := svc.DeleteMembers(context.Background(), []string{seed.IDMember1}); err != nil {
		t.Fatal(err)
	}

	members := st.Org().Members()
	for _, member := range members {
		if member.ID == seed.IDMember1 && member.Status != "inactive" {
			t.Fatalf("expected inactive status, got %s", member.Status)
		}
	}

	keys := st.Keys().PlatformKeys()
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == seed.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}

func TestTransferMembersUpdatesRelayMapping(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := testutil.NewOrgService(t, cfg, st)

	memberID := seed.IDMember1
	targetDept := "dept-4"
	if err := st.Relay().UpsertMapping(store.RelayMapping{
		PlatformKeyID: seed.IDPlatformKey1,
		MemberID:      &memberID,
		DepartmentID:  seed.IDDept3,
		RelayGroup:    "default",
		SyncStatus:    store.RelaySyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}

	if err := svc.TransferMembers(context.Background(), []string{memberID}, targetDept); err != nil {
		t.Fatal(err)
	}

	members := st.Org().Members()
	for _, member := range members {
		if member.ID != memberID {
			continue
		}
		if member.DepartmentID != targetDept {
			t.Fatalf("expected department %s, got %s", targetDept, member.DepartmentID)
		}
	}

	mappings, err := st.Relay().ListMappingsByMemberID(memberID)
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

	if err := svc.UpdateMemberStatus(context.Background(), []string{seed.IDMember1}, "inactive"); err != nil {
		t.Fatal(err)
	}
	for _, key := range st.Keys().PlatformKeys() {
		if key.MemberID != nil && *key.MemberID == seed.IDMember1 && key.Status != "disabled" {
			t.Fatalf("expected disabled key, got %s", key.Status)
		}
	}
}
