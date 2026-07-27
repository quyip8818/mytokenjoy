package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

// PRD: 成员状态: pending→active(注册), active⇄disabled, pending→硬删, active/disabled→disabled(软删)

func TestMemberStatusTransition_ActiveToDisabled(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	if err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "disabled"); err != nil {
		t.Fatal(err)
	}

	members, _ := st.Org().Members(ctx)
	for _, m := range members {
		if m.ID == contract.IDMember1 && m.Status != "disabled" {
			t.Fatalf("expected disabled, got %s", m.Status)
		}
	}
}

func TestMemberStatusTransition_DisabledToActive(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// First disable
	svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "disabled")
	// Then re-enable
	if err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "active"); err != nil {
		t.Fatal(err)
	}

	members, _ := st.Org().Members(ctx)
	for _, m := range members {
		if m.ID == contract.IDMember1 && m.Status != "active" {
			t.Fatalf("expected active, got %s", m.Status)
		}
	}
}

func TestMemberStatusTransition_PendingToActiveRejected(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Set member to pending
	members, _ := st.Org().Members(ctx)
	for i := range members {
		if members[i].ID == contract.IDMember1 {
			members[i].Status = types.MemberStatusPending
		}
	}
	st.Org().SetMembers(ctx, members)

	// Attempt to manually activate should fail
	err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "active")
	if err == nil {
		t.Fatal("expected error when activating pending member manually")
	}
}

func TestMemberDisableDisablesAllKeys(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Add a second key for the same member
	keys, _ := st.Keys().PlatformKeys(ctx)
	memberID := contract.IDMember1
	keys = append(keys, types.PlatformKey{
		ID: uuid.MustParse("00000000-0000-7000-0000-00000000ff88"), Name: "Extra Key", Status: "active", MemberID: &memberID,
	})
	st.Keys().SetPlatformKeys(ctx, keys)

	// Disable member
	if err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "disabled"); err != nil {
		t.Fatal(err)
	}

	// ALL keys belonging to this member should be disabled
	keys, _ = st.Keys().PlatformKeys(ctx)
	for _, key := range keys {
		if key.MemberID != nil && *key.MemberID == contract.IDMember1 {
			if key.Status != "disabled" {
				t.Errorf("key %s should be disabled, got %s", key.ID, key.Status)
			}
		}
	}
}

func TestMemberDeleteSetsDisabled(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// DeleteMembers for active member → soft delete (set disabled)
	if err := svc.DeleteMembers(ctx, []uuid.UUID{contract.IDMember1}, uuid.Nil); err != nil {
		t.Fatal(err)
	}

	members, _ := st.Org().Members(ctx)
	for _, m := range members {
		if m.ID == contract.IDMember1 && m.Status != "disabled" {
			t.Fatalf("expected disabled after delete, got %s", m.Status)
		}
	}
}
