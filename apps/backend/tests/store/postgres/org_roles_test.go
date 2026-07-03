//go:build integration

package postgres_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestRoleMembersRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	roles, err := st.Org().Roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var custom types.Role
	for _, role := range roles {
		if role.Type == "custom" {
			custom = role
			break
		}
	}
	if custom.ID == "" {
		custom = types.Role{
			ID: "role-pg-test", Name: "PG Test", Type: "custom",
			Permissions: []string{"p-1"}, MemberCount: 0,
		}
		roles = append(roles, custom)
		if err := st.Org().SetRoles(ctx, roles); err != nil {
			t.Fatal(err)
		}
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	targetID := seed.IDMember3
	for i := range members {
		if members[i].ID != targetID {
			continue
		}
		found := false
		for _, roleName := range members[i].Roles {
			if roleName == custom.Name {
				found = true
				break
			}
		}
		if !found {
			members[i].Roles = append(members[i].Roles, custom.Name)
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			t.Fatal(err)
		}
		break
	}

	reloaded, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range reloaded {
		if member.ID != targetID {
			continue
		}
		hasRole := false
		for _, roleName := range member.Roles {
			if roleName == custom.Name {
				hasRole = true
				break
			}
		}
		if !hasRole {
			t.Fatalf("expected member %s to have role %s", targetID, custom.Name)
		}
	}
}

func TestMembersPersistByDepartment(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	deptCount := map[string]int{}
	for _, member := range members {
		deptCount[member.DepartmentID]++
	}
	if deptCount[seed.IDDept3] == 0 {
		t.Fatal("expected members in seed dept-3")
	}

	restarted := testPostgresStore(t)
	reloaded, err := restarted.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded) != len(members) {
		t.Fatalf("expected %d members after restart, got %d", len(members), len(reloaded))
	}
}
