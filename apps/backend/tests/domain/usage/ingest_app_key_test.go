package usage_test

import (
	"testing"

	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestAppKeyRollsUpDepartment(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ingest := testutil.NewIngestService(t, cfg, st)
	ctx := testutil.Ctx()

	fullKey := "sk-app-key-test"
	if err := st.Keys().SetPlatformKeys(ctx, []types.PlatformKey{{
		ID:        "plk-3",
		Name:      "App Key",
		KeyPrefix: "sk-app",
		FullKey:   &fullKey,
		Status:    "active",
		CreatedAt: "2026-06-19",
	}}); err != nil {
		t.Fatal(err)
	}

	newapisynctf.UpsertMapping(t, st, newapisynctf.MappingOpts{
		PlatformKeyID: "plk-3",
		NewAPIKeyID:   77,
		NoMember:      true,
		DepartmentID:  contract.IDDept3,
	})

	before := testutil.Dept3SnapshotConsumed(t, st)

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(98002, 77))
	if err := ingest.IngestByLogID(ctx, 98002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	after := testutil.Dept3SnapshotConsumed(t, st)
	if after <= before {
		t.Fatalf("expected department rollup for app key, before=%v after=%v", before, after)
	}
}
