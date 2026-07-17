package org_test

import (
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestTransferMembersDoesNotBumpAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.TransferMembers(ctx, []string{contract.IDMember1.String()}, contract.IDDept4); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("expected authz revision unchanged after transfer, before=%d after=%d", before, after)
	}
}

func TestTransferMembersUpdatesPlatformKeyMapping(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	memberID := contract.IDMember1
	targetDept := contract.IDDept4
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		MemberID:      &memberID,
		DepartmentID:  contract.IDDept3,
		NewAPIGroup:   "default",
		SyncStatus:    store.MappingSyncStatusSynced,
	}); err != nil {
		t.Fatal(err)
	}

	if err := svc.TransferMembers(testutil.Ctx(), []string{memberID.String()}, targetDept); err != nil {
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

	mappings, err := st.PlatformKeyMappings().ListMappingsByMemberID(ctx, memberID)
	if err != nil {
		t.Fatal(err)
	}
	if len(mappings) == 0 {
		t.Fatal("expected platform key mapping")
	}
	if mappings[0].DepartmentID != targetDept {
		t.Fatalf("expected mapping department %s, got %s", targetDept, mappings[0].DepartmentID)
	}
}
