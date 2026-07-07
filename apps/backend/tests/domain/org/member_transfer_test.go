package org_test

import (
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestTransferMembersDoesNotBumpAuthzRevision(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	before, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.TransferMembers(ctx, []string{seed.IDMember1}, "dept-4"); err != nil {
		t.Fatal(err)
	}
	after, err := st.Company().GetAuthzRevision(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("expected authz revision unchanged after transfer, before=%d after=%d", before, after)
	}
}

func TestTransferMembersUpdatesRelayMapping(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
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
