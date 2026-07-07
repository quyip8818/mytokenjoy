package usage_test

import (
	"testing"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestIngestAppKeyRollsUpDepartment(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithIngestEnabled(true))
	ingest := testutil.NewIngestService(t, cfg, st)
	ctx := testutil.Ctx()

	relayfix.UpsertMapping(t, st, relayfix.MappingOpts{
		PlatformKeyID: "plk-3",
		NewAPITokenID: 77,
		NoMember:      true,
		DepartmentID:  seed.IDDept3,
	})

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	before := testutil.Dept3Consumed(t, tree)

	testutil.SeedConsumeLog(t, st, store.RawConsumeLog{
		ID: 8001, TokenID: 77, Quota: 500000, ModelName: "gpt-4o-mini", CreatedAt: 1,
	})
	if err := ingest.IngestByLogID(ctx, 8001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	tree, err = common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	after := testutil.Dept3Consumed(t, tree)
	if after <= before {
		t.Fatalf("expected department rollup for app key, before=%v after=%v", before, after)
	}
}
