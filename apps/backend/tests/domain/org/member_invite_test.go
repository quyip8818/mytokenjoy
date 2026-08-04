package org_test

import (
	"testing"

	"github.com/google/uuid"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestBatchInviteByIDs(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Use an existing seeded member as the pending target (avoid needing to insert user row).
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) == 0 {
		t.Fatal("no seed members")
	}
	// Pick the first active member and flip its status to pending.
	pendingID := members[0].ID
	members[0].Status = types.MemberStatusPending
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchInvite(ctx, []uuid.UUID{pendingID}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Sent != 1 {
		t.Fatalf("expected sent=1, got %d", result.Sent)
	}
}

func TestBatchInviteAllPending(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Set all members to pending so BatchInvite targets them all.
	for i := range members {
		members[i].Status = types.MemberStatusPending
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchInvite(ctx, nil, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Sent != len(members) {
		t.Fatalf("expected sent=%d, got %d", len(members), result.Sent)
	}
}

func TestBatchInviteReusesExistingInvite(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) == 0 {
		t.Fatal("no seed members")
	}
	// Make first member pending.
	pendingID := members[0].ID
	members[0].Status = types.MemberStatusPending
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}

	// First call — creates invite.
	result1, err := svc.BatchInvite(ctx, []uuid.UUID{pendingID}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result1.Sent != 1 {
		t.Fatalf("first call: expected sent=1, got %d", result1.Sent)
	}

	// Get the invite that was created.
	invite1, err := st.Invite().GetInviteByMemberID(ctx, pendingID)
	if err != nil || invite1 == nil {
		t.Fatal("expected invite to be created after first BatchInvite")
	}

	// Second call — should reuse same invite (not create a new one).
	result2, err := svc.BatchInvite(ctx, []uuid.UUID{pendingID}, uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	if result2.Sent != 1 {
		t.Fatalf("second call: expected sent=1, got %d", result2.Sent)
	}

	invite2, err := st.Invite().GetInviteByMemberID(ctx, pendingID)
	if err != nil || invite2 == nil {
		t.Fatal("expected invite to still exist")
	}
	if invite1.ID != invite2.ID {
		t.Fatalf("expected invite to be reused (same ID), got %s vs %s", invite1.ID, invite2.ID)
	}
	if invite1.InviteCode != invite2.InviteCode {
		t.Fatalf("expected same invite code on reuse, got %s vs %s", invite1.InviteCode, invite2.InviteCode)
	}
}
