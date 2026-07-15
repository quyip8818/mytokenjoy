package budget

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestIndexMembersByRole(t *testing.T) {
	t.Parallel()
	members := []types.Member{
		{ID: "m-1", Status: "active", Roles: []string{"super_admin", "org_admin"}},
		{ID: "m-2", Status: "active", Roles: []string{"org_admin"}},
		{ID: "m-3", Status: "inactive", Roles: []string{"super_admin"}}, // inactive → excluded
		{ID: "m-4", Status: "active", Roles: []string{}},                // no roles
	}
	result := indexMembersByRole(members)

	if len(result["super_admin"]) != 1 || result["super_admin"][0] != "m-1" {
		t.Errorf("super_admin = %v, want [m-1]", result["super_admin"])
	}
	if len(result["org_admin"]) != 2 {
		t.Errorf("org_admin = %v, want [m-1, m-2]", result["org_admin"])
	}
	if _, ok := result["member"]; ok {
		t.Errorf("expected no 'member' key, got %v", result["member"])
	}
}

func TestResolveRoleRecipients(t *testing.T) {
	t.Parallel()
	roleNameByID := map[string]string{
		"role-1": "super_admin",
		"role-2": "org_admin",
		"role-3": "member",
	}
	membersByRoleName := map[string][]string{
		"super_admin": {"m-1"},
		"org_admin":   {"m-1", "m-2"},
		"member":      {"m-3", "m-4"},
	}

	tests := []struct {
		name    string
		roleIDs []string
		want    int // expected unique recipients
	}{
		{"single role", []string{"role-1"}, 1},
		{"overlapping roles deduplicate", []string{"role-1", "role-2"}, 2}, // m-1 appears only once
		{"unknown role ID skipped", []string{"role-999"}, 0},
		{"empty", nil, 0},
		{"all roles", []string{"role-1", "role-2", "role-3"}, 4}, // m-1, m-2, m-3, m-4
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveRoleRecipients(tc.roleIDs, roleNameByID, membersByRoleName)
			if len(got) != tc.want {
				t.Errorf("got %d recipients %v, want %d", len(got), got, tc.want)
			}
			// Verify no duplicates.
			seen := make(map[string]struct{})
			for _, id := range got {
				if _, dup := seen[id]; dup {
					t.Errorf("duplicate recipient: %s", id)
				}
				seen[id] = struct{}{}
			}
		})
	}
}
