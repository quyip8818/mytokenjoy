//go:build testhook && integration

package worker_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func TestWorkerProcessesRebalanceQueue(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()

	tokenID := int64(42)
	if err := fix.st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID: contract.DefaultCompanyID, PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID: &tokenID, MemberID: &contract.IDMember1,
		DepartmentID: contract.IDDept3, SyncStatus: store.MappingSyncStatusSynced,
		NewAPIGroup: "dept-dept-3",
	}); err != nil {
		t.Fatal(err)
	}
	if err := jobs.InsertRebalance(ctx, fix.rt.Enqueuer, nil, contract.DefaultCompanyID, store.RebalanceAxisMember, contract.IDMember1); err != nil {
		t.Fatal(err)
	}

	fix.runRiver(t)

	// Rebalance no longer calls NewAPI (tokens are unlimited_quota=true).
	// Verify the job was drained and combined_key_remain was recomputed.
	if riverfix.PendingRebalanceCount(fix.st, contract.DefaultCompanyID) != 0 {
		t.Fatal("expected rebalance queue drained")
	}
	summaries, err := fix.st.CombinedKeySummaries().ListByPlatformKeyIDs(ctx, []uuid.UUID{contract.IDPlatformKey1})
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) == 0 {
		t.Fatal("expected combined_key_summary to be computed after rebalance")
	}
}

func TestWorkerCompanyRebalanceSetsLastRebalancedPeriod(t *testing.T) {
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 42, RemainQuota: 1000}}
	fix := newWorkerFixture(t, stub)
	ctx := testutil.Ctx()
	current := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, fix.rt.Registry.Config.Clock()).String()

	if err := jobs.InsertRebalance(ctx, fix.rt.Enqueuer, nil, contract.DefaultCompanyID, store.RebalanceAxisCompany, contract.DefaultCompanyID); err != nil {
		t.Fatal(err)
	}

	fix.runRiver(t)

	tbs, err := fix.st.TenantBackgroundState().Get(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if tbs == nil || tbs.LastRebalancedPeriod != current {
		t.Fatalf("expected last_rebalanced_period=%q, got %+v", current, tbs)
	}
}
