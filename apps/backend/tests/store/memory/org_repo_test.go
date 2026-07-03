package memory_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOrgMembersPersonalQuota(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()
	if err := st.Org().UpdateMemberPersonalQuota(ctx, seed.IDMember1, 9999); err != nil {
		t.Fatal(err)
	}
	quota, ok, err := st.Org().MemberPersonalQuota(ctx, seed.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || quota != 9999 {
		t.Fatalf("expected quota 9999, got %f ok=%v", quota, ok)
	}
}

func TestOrgMembersFilterByDepartment(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, member := range members {
		if member.DepartmentID == seed.IDDept3 {
			count++
		}
	}
	if count == 0 {
		t.Fatal("expected members in dept-3 from seed")
	}
}

func TestOrgRolesRoundTrip(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()
	roles, err := st.Org().Roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	roles = append(roles, types.Role{
		ID: "role-test", Name: "Test Role", Type: "custom",
		Permissions: []string{"p-1"}, MemberCount: 0,
	})
	if err := st.Org().SetRoles(ctx, roles); err != nil {
		t.Fatal(err)
	}
	got, err := st.Org().Roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, role := range got {
		if role.ID == "role-test" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected persisted role")
	}
}

func TestOrgIntegrationCredentialIndependent(t *testing.T) {
	_, st := testutil.NewMemoryStoreFromConfig(t)
	ctx := testutil.Ctx()
	integration := types.OrgIntegration{
		Enabled: true, StartTime: "05:00", FrequencyHours: 6,
	}
	if err := st.Org().SetIntegration(ctx, integration); err != nil {
		t.Fatal(err)
	}
	if err := st.Org().SaveIntegrationCredential(ctx, types.PlatformFeishu, []byte("encrypted")); err != nil {
		t.Fatal(err)
	}
	got, err := st.Org().Integration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.StartTime != "05:00" {
		t.Fatalf("sync config changed: %+v", got)
	}
}
