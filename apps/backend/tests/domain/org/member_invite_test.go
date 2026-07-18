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

	result, err := svc.BatchInvite(ctx, []uuid.UUID{pendingID})
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
	pendingCount := 0
	for i := range members {
		if members[i].Status == types.MemberStatusPending || members[i].Status == types.MemberStatusInactive {
			pendingCount++
			continue
		}
		members[i].Status = types.MemberStatusInactive
		pendingCount++
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchInvite(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Sent != pendingCount {
		t.Fatalf("expected sent=%d, got %d", pendingCount, result.Sent)
	}
}
