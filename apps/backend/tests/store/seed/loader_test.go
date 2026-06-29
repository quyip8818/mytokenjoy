package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store/seed"
)

func TestBuildMembersAnchorIDs(t *testing.T) {
	members := seed.BuildMembers()
	if len(members) < 120 {
		t.Fatalf("expected at least 120 members, got %d", len(members))
	}

	for _, id := range []string{"m-admin", "m-1", "m-2"} {
		if _, ok := org.FindMemberByID(members, id); !ok {
			t.Fatalf("expected anchor member %s", id)
		}
	}
}

func TestLoadSnapshot(t *testing.T) {
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("NEW_API_ENABLED", "false")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	snapshot := seed.Load(cfg)
	if len(snapshot.Departments) == 0 {
		t.Fatal("expected departments in snapshot")
	}
	if len(snapshot.Roles) != 6 {
		t.Fatalf("expected 6 roles, got %d", len(snapshot.Roles))
	}
	if len(snapshot.BudgetTree) == 0 || snapshot.BudgetTree[0].ID != "dept-1" {
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
	if len(snapshot.CallLogs) == 0 {
		t.Fatal("expected call logs")
	}
	if !snapshot.AuditSettings.ContentRetentionEnabled {
		t.Fatal("expected audit content retention enabled")
	}
}
