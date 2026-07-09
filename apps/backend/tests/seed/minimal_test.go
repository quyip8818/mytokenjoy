package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/filler"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestLoadMinimalSnapshot(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig(testutil.WithBootstrapMode(config.BootstrapMinimal))
	snapshot := seed.LoadMinimal(cfg)
	if len(snapshot.Members) != len(filler.BuildAnchorMembers()) {
		t.Fatalf("expected %d anchor members, got %d", len(filler.BuildAnchorMembers()), len(snapshot.Members))
	}
	if len(snapshot.PlatformKeys) != 1 || snapshot.PlatformKeys[0].ID != contract.IDPlatformKey1 {
		t.Fatalf("expected single anchor platform key, got %+v", snapshot.PlatformKeys)
	}
	if len(snapshot.BudgetGroups) != 1 || snapshot.BudgetGroups[0].ID != contract.IDBudgetGroup1 {
		t.Fatalf("expected minimal budget group bg-1, got %+v", snapshot.BudgetGroups)
	}
	if len(snapshot.UsageLedger) != 0 {
		t.Fatalf("expected no usage ledger in minimal seed, got %d", len(snapshot.UsageLedger))
	}
}

func TestMinimalSeedStore(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithBootstrapMode(config.BootstrapMinimal))
	members, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(members) < 8 {
		t.Fatalf("expected at least 8 members, got %d", len(members))
	}
	if len(members) > 15 {
		t.Fatalf("expected minimal seed member count, got %d", len(members))
	}
}

func TestLoadMinimalFromConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", config.DefaultDatabaseURL)
	t.Setenv("COMPANY_NAME", "Demo Company")
	t.Setenv("NEW_API_ENABLED", "false")
	t.Setenv("SESSION_SECRET", "test-session-secret")
	t.Setenv("DATA_SOURCE_CREDENTIAL_KEY", testutil.DefaultTestCredentialKey)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.BootstrapMode = config.BootstrapMinimal
	snapshot := seed.LoadMinimal(cfg)
	if snapshot.Company.ID != contract.DefaultCompanyID {
		t.Fatalf("expected default company, got %+v", snapshot.Company)
	}
}
