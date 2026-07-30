package seed_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/filler"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestLoadMinimalSnapshot(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	snapshot := seed.LoadMinimal(cfg)
	if len(snapshot.Members) != len(filler.BuildAnchorMembers()) {
		t.Fatalf("expected %d anchor members, got %d", len(filler.BuildAnchorMembers()), len(snapshot.Members))
	}
	if len(snapshot.PlatformKeys) != 1 || snapshot.PlatformKeys[0].ID != contract.IDPlatformKey1 {
		t.Fatalf("expected single anchor platform key, got %+v", snapshot.PlatformKeys)
	}
	if len(snapshot.Projects) != 1 || snapshot.Projects[0].ID != contract.IDProject1 {
		t.Fatalf("expected minimal project proj-1, got %+v", snapshot.Projects)
	}
	if len(snapshot.UsageLedger) != 0 {
		t.Fatalf("expected no usage ledger in minimal seed, got %d", len(snapshot.UsageLedger))
	}
}

func TestMinimalSeedStore(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	members, err := st.Org().Members(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(members) < 8 {
		t.Fatalf("expected at least 8 members, got %d", len(members))
	}
}

func TestLoadMinimalFromConfig(t *testing.T) {
	cfg := testutil.TestConfig()
	snapshot := seed.LoadMinimal(cfg)
	if snapshot.Company.ID != contract.DefaultCompanyID {
		t.Fatalf("expected default company, got %+v", snapshot.Company)
	}
}
