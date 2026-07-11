package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/filler"
)

func TestBuildMembersAnchorIDs(t *testing.T) {
	t.Parallel()
	members := filler.BuildMembers()
	if len(members) < 35 {
		t.Fatalf("expected at least 35 members, got %d", len(members))
	}

	for _, id := range []string{"m-admin", "m-1", "m-2"} {
		if _, ok := org.FindMemberByID(members, id); !ok {
			t.Fatalf("expected anchor member %s", id)
		}
	}
}

func TestLoadSnapshot(t *testing.T) {
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Demo Company")
	t.Setenv("NEW_API_ENABLED", "true")
	t.Setenv("NEW_API_BASE_URL", "http://127.0.0.1:3000")
	t.Setenv("NEW_API_ADMIN_TOKEN", "admin-token")
	t.Setenv("SESSION_SECRET", "test-session-secret")
	t.Setenv("DATA_SOURCE_CREDENTIAL_KEY", "dGV2LWNyZWRlbnRpYWwta2V5LWZvci1sb2NhbC1kZXY=")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := seed.Load(cfg)
	if snapshot.Company.ID != contract.DefaultCompanyID || snapshot.Company.Slug != "default" {
		t.Fatalf("expected default company, got %+v", snapshot.Company)
	}
	if len(snapshot.OrgNodes) == 0 {
		t.Fatal("expected departments in snapshot")
	}
	if len(snapshot.Roles) != 6 {
		t.Fatalf("expected 6 roles, got %d", len(snapshot.Roles))
	}
	if len(types.OrgNodesToBudgetTree(snapshot.OrgNodes)) == 0 || types.OrgNodesToBudgetTree(snapshot.OrgNodes)[0].ID != "dept-1" {
		t.Fatal("expected budget tree root")
	}
	if len(snapshot.ProviderKeys) < 8 {
		t.Fatalf("expected provider keys, got %d", len(snapshot.ProviderKeys))
	}
	if len(snapshot.PlatformKeys) == 0 {
		t.Fatal("expected platform keys")
	}
	foundPending := false
	for _, approval := range snapshot.Approvals {
		if approval.ID == "apv-1" {
			foundPending = true
			break
		}
	}
	if !foundPending {
		t.Fatal("expected pending approval apv-1")
	}
	if len(snapshot.OperationLogs) == 0 {
		t.Fatal("expected operation logs")
	}
	if len(snapshot.UsageLedger) == 0 {
		t.Fatal("expected usage ledger entries")
	}
	if !snapshot.AuditSettings.ContentRetentionEnabled {
		t.Fatal("expected audit content retention enabled")
	}
}
