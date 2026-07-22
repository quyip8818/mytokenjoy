package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

// PRD 2.2 成员状态: [创建]→启用, [邀请]→未激活→启用, 启用⇄停用, 停用/启用→删除

func TestMemberStatusTransition_ActiveToInactive(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Active member goes inactive
	if err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "inactive"); err != nil {
		t.Fatal(err)
	}

	members, _ := st.Org().Members(ctx)
	for _, m := range members {
		if m.ID == contract.IDMember1 && m.Status != "inactive" {
			t.Fatalf("expected inactive, got %s", m.Status)
		}
	}
}

func TestMemberStatusTransition_InactiveToActive(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// First disable
	svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "inactive")
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
	if err := svc.UpdateMemberStatus(ctx, []uuid.UUID{contract.IDMember1}, "inactive"); err != nil {
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

func TestMemberDeleteSetsInactive(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// DeleteMembers is implemented as UpdateMemberStatus to "inactive"
	if err := svc.DeleteMembers(ctx, []uuid.UUID{contract.IDMember1}, uuid.Nil); err != nil {
		t.Fatal(err)
	}

	members, _ := st.Org().Members(ctx)
	for _, m := range members {
		if m.ID == contract.IDMember1 && m.Status != "inactive" {
			t.Fatalf("expected inactive after delete, got %s", m.Status)
		}
	}
}

func TestBatchStatusChangeMultipleMembers(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	members, _ := st.Org().Members(ctx)
	ids := make([]uuid.UUID, 0)
	for _, m := range members {
		if m.Status == "active" && len(ids) < 3 {
			ids = append(ids, m.ID)
		}
	}
	if len(ids) < 2 {
		t.Skip("not enough active members for batch test")
	}

	// Batch disable
	if err := svc.UpdateMemberStatus(ctx, ids, "inactive"); err != nil {
		t.Fatal(err)
	}

	members, _ = st.Org().Members(ctx)
	for _, m := range members {
		for _, id := range ids {
			if m.ID == id && m.Status != "inactive" {
				t.Errorf("member %s should be inactive", m.ID)
			}
		}
	}
}

func TestCreateMemberDefaultsToActive(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	member, err := svc.CreateMember(ctx, types.Member{
		Alias: "New Person", Phone: "13900009999", Email: "new@example.com",
		DepartmentID: contract.IDDept3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.Status != "active" {
		t.Errorf("new member should be active, got %s", member.Status)
	}
}

func TestBatchInviteSetsStatus(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := orgfix.NewService(t, cfg, st)
	ctx := testutil.Ctx()

	// Create a member first
	member, err := svc.CreateMember(ctx, types.Member{
		Alias: "Invite Target", Phone: "13900001111", Email: "invite@example.com",
		DepartmentID: contract.IDDept3,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := svc.BatchInvite(ctx, []uuid.UUID{member.ID})
	if err != nil {
		t.Fatal(err)
	}
	if result.Sent == 0 {
		t.Error("expected at least 1 invite sent")
	}
}
