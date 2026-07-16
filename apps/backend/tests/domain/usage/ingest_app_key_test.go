package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestIngestAppKeyIncrementsPlatformKeyConsumed(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 77, RemainQuota: 1000}}
	runner, st, ingest := riverfix.NewIngestRuntime(t, stub)
	ctx := testutil.Ctx()

	fullKey := "sk-app-key-test"
	if err := st.Keys().SetPlatformKeys(ctx, []types.PlatformKey{{
		ID:        "plk-3",
		Name:      "App Key",
		KeyPrefix: "sk-app",
		Scope:     types.PlatformKeyScopeProject,
		FullKey:   &fullKey,
		Status:    "active",
		CreatedAt: "2026-06-19",
	}}); err != nil {
		t.Fatal(err)
	}

	newapisynctf.PrepareIngestFixture(t, st, newapisynctf.MappingOpts{
		PlatformKeyID: "plk-3",
		NewAPIKeyID:   77,
		NoMember:      true,
		DepartmentID:  contract.IDDept3,
	})

	before := budgetfix.PlatformKeySnapshotConsumed(t, st, "plk-3")

	testutil.SeedConsumeLog(t, st, testutil.DefaultConsumeLog(98002, 77))
	if err := ingest.IngestByLogID(ctx, 98002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}
	runner.RunOnce(t, ctx)

	after := budgetfix.PlatformKeySnapshotConsumed(t, st, "plk-3")
	if after <= before {
		t.Fatalf("expected platform key consumed increase for app key, before=%v after=%v", before, after)
	}
}
