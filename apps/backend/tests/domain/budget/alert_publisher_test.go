package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestIndexMembersByRole(t *testing.T) {
	t.Parallel()
	m1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	m2 := uuid.MustParse("00000000-0000-7000-0000-000000000002")
	m3 := uuid.MustParse("00000000-0000-7000-0000-000000000003")
	m4 := uuid.MustParse("00000000-0000-7000-0000-000000000004")
	members := []types.Member{
		{ID: m1, Status: "active", Roles: []string{"super_admin", "org_admin"}},
		{ID: m2, Status: "active", Roles: []string{"org_admin"}},
		{ID: m3, Status: "inactive", Roles: []string{"super_admin"}},
		{ID: m4, Status: "active", Roles: []string{}},
	}
	result := budget.IndexMembersByRole(members)

	if len(result["super_admin"]) != 1 || result["super_admin"][0] != m1 {
		t.Errorf("super_admin = %v, want [%s]", result["super_admin"], m1)
	}
	if len(result["org_admin"]) != 2 {
		t.Errorf("org_admin = %v, want [%s, %s]", result["org_admin"], m1, m2)
	}
	if _, ok := result["member"]; ok {
		t.Errorf("expected no 'member' key, got %v", result["member"])
	}
}

func TestResolveRoleRecipients(t *testing.T) {
	t.Parallel()
	role1 := uuid.MustParse("00000000-0000-7000-0000-000000000a01")
	role2 := uuid.MustParse("00000000-0000-7000-0000-000000000a02")
	role3 := uuid.MustParse("00000000-0000-7000-0000-000000000a03")
	role999 := uuid.MustParse("00000000-0000-7000-0000-000000000a99")
	roleNameByID := map[uuid.UUID]string{
		role1: "super_admin",
		role2: "org_admin",
		role3: "member",
	}
	m1 := uuid.MustParse("00000000-0000-7000-0000-00000000aa01")
	m2 := uuid.MustParse("00000000-0000-7000-0000-00000000aa02")
	m3 := uuid.MustParse("00000000-0000-7000-0000-00000000aa03")
	m4 := uuid.MustParse("00000000-0000-7000-0000-00000000aa04")
	membersByRoleName := map[string][]uuid.UUID{
		"super_admin": {m1},
		"org_admin":   {m1, m2},
		"member":      {m3, m4},
	}

	tests := []struct {
		name    string
		roleIDs []uuid.UUID
		want    int
	}{
		{"single role", []uuid.UUID{role1}, 1},
		{"overlapping roles deduplicate", []uuid.UUID{role1, role2}, 2},
		{"unknown role ID skipped", []uuid.UUID{role999}, 0},
		{"empty", nil, 0},
		{"all roles", []uuid.UUID{role1, role2, role3}, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := budget.ResolveRoleRecipients(tc.roleIDs, roleNameByID, membersByRoleName)
			if len(got) != tc.want {
				t.Errorf("got %d recipients %v, want %d", len(got), got, tc.want)
			}
			seen := make(map[uuid.UUID]struct{})
			for _, id := range got {
				if _, dup := seen[id]; dup {
					t.Errorf("duplicate recipient: %s", id)
				}
				seen[id] = struct{}{}
			}
		})
	}
}
