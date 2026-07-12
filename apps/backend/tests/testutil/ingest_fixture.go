//go:build testhook

package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

// IngestBudgetFixture describes budget axes checked by ingest enforceBudgetCap.
type IngestBudgetFixture struct {
	DepartmentID  string
	PlatformKeyID string
	MemberID      string // empty when ingest mapping has no member quota axis
	Amount        float64
}

func DefaultConsumeLogQuota() float64 {
	return 500_000
}

// PrepareIngestBudgetHeadroom sets snapshot consumed values so enforceBudgetCap
// allows an ingest of fixture.Amount. Demo seed may exceed department budgets by
// design; tests that expect successful ingest must call this explicitly.
func PrepareIngestBudgetHeadroom(t *testing.T, st store.Store, fixture IngestBudgetFixture) {
	t.Helper()
	if fixture.DepartmentID == "" {
		t.Fatal("PrepareIngestBudgetHeadroom: departmentID required")
	}
	if fixture.Amount <= 0 {
		fixture.Amount = DefaultConsumeLogQuota()
	}

	ctx := Ctx()
	limit, found, err := st.Org().Nodes().GetNodeBudget(ctx, fixture.DepartmentID)
	if err != nil {
		t.Fatal(err)
	}
	if found && limit > 0 {
		budgetfix.SetDeptSnapshotConsumed(t, st, fixture.DepartmentID, ingestHeadroomConsumed(limit, fixture.Amount))
	}

	if fixture.MemberID != "" {
		quota, memberFound, err := st.Org().MemberPersonalBudget(ctx, fixture.MemberID)
		if err != nil {
			t.Fatal(err)
		}
		if memberFound {
			budgetfix.SetMemberSnapshotConsumed(t, st, fixture.MemberID, ingestHeadroomConsumed(quota, fixture.Amount))
		}
	}

	if fixture.PlatformKeyID == "" {
		return
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID != fixture.PlatformKeyID || key.Budget <= 0 {
			continue
		}
		budgetfix.SetPlatformKeySnapshotUsed(t, st, fixture.PlatformKeyID, ingestHeadroomConsumed(key.Budget, fixture.Amount))
		break
	}
}

func ingestHeadroomConsumed(limit, amount float64) float64 {
	consumed := limit - amount
	if consumed < 0 {
		return 0
	}
	return consumed
}
