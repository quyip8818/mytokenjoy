package org_test

import (
	"testing"

	"github.com/google/uuid"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestBatchInviteByIDs(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	pendingID := uuid.MustParse("00000000-0000-7000-0000-00000000ff99")
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range members {
		if members[i].ID == pendingID {
			continue
		}
		if members[i].Status == types.MemberStatusActive {
			members[i].Status = types.MemberStatusPending
		}
	}
	members = append(members, types.Member{
		ID: pendingID, Name: "Pending User", DepartmentID: contract.IDDept5,
		Status: types.MemberStatusPending, Roles: []string{"普通成员"},
	})
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
