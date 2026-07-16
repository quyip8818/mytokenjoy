//go:build testhook && integration

package worker_test

import (
	"strings"
	"testing"

	riverfix "github.com/tokenjoy/backend/tests/testutil/river"

	"github.com/tokenjoy/backend/internal/domain/newapisync/outbox"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestProcessUnknownNewAPISyncOutboxKindFails(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	if err := jobs.InsertNewAPISync(ctx, fix.rt.Enqueuer, nil, jobs.NewAPISyncArgs{
		CompanyID: contract.DefaultCompanyID,
		SubKind:   "unknown_kind",
	}); err != nil {
		t.Fatal(err)
	}

	fix.runRiver(t)

	pool := postgres.MainPool(fix.st)
	var state string
	var lastErr *string
	if err := pool.QueryRow(ctx, `
		SELECT state::text,
			CASE WHEN cardinality(errors) > 0 THEN (errors[cardinality(errors)]->>'error') ELSE NULL END
		FROM river_job WHERE kind = $1 ORDER BY id DESC LIMIT 1
	`, jobs.KindNewAPISync).Scan(&state, &lastErr); err != nil {
		t.Fatal(err)
	}
	if state != "cancelled" && state != "discarded" {
		t.Fatalf("expected cancelled/discarded state, got %q", state)
	}
	if lastErr == nil || !strings.Contains(*lastErr, "unknown newapi sync sub kind") {
		t.Fatalf("expected unknown kind error recorded, got %v", lastErr)
	}
}

func TestProcessNewAPISyncOutbox(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, Key: "sk-worker", RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()
	if err := fix.st.Company().UpdateNewAPIWalletUserID(ctx, contract.DefaultCompanyID, 501); err != nil {
		t.Fatal(err)
	}

	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID: "plk-worker", Name: "worker-key", Scope: types.PlatformKeyScopeMember, MemberID: &memberID,
		Status: "active", Budget: 1000, ModelWhitelist: []int64{contract.IDModel1},
		CreatedAt: "2026-06-19",
	}
	keys, err := fix.st.Keys().PlatformKeys(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := fix.st.Keys().SetPlatformKeys(testutil.Ctx(), keys); err != nil {
		t.Fatal(err)
	}

	if err := fix.newAPISync.SyncCreatePlatformKey(ctx, key, contract.IDDept3); err != nil {
		t.Fatal(err)
	}
	if riverfix.ListPendingNewAPISync(t, fix.st, outbox.KindCreateKey, 100) == 0 {
		t.Fatal("expected pending create_key outbox before RunOnce")
	}

	fix.runRiver(t)

	if stub.CreateTokenCalls < 1 {
		t.Fatalf("expected CreateToken to be called, got %d", stub.CreateTokenCalls)
	}
	if riverfix.ListPendingNewAPISync(t, fix.st, outbox.KindCreateKey, 100) != 0 {
		t.Fatal("expected newapi sync outbox done after RunOnce")
	}
}
