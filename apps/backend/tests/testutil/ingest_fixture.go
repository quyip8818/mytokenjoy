//go:build testhook

package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

// IngestBudgetFixture describes budget axes used when preparing ingest headroom in tests.
type IngestBudgetFixture struct {
	DepartmentID  string
	PlatformKeyID string
	MemberID      string // empty when ingest mapping has no member quota axis
	Amount        float64
}

func DefaultConsumeLogQuota() float64 {
	return 500_000
}

// PrepareIngestBudgetHeadroom sets snapshot consumed values so ingest projections
// do not immediately trigger overrun disable for the seeded demo budgets.
func PrepareIngestBudgetHeadroom(t *testing.T, st store.Store, fixture IngestBudgetFixture) {
	t.Helper()
	if fixture.DepartmentID == "" {
		t.Fatal("PrepareIngestBudgetHeadroom: departmentID required")
	}
	if fixture.Amount <= 0 {
		fixture.Amount = DefaultConsumeLogQuota()
	}

	ctx := Ctx()
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
		budgetfix.SetPlatformKeySnapshotConsumed(t, st, fixture.PlatformKeyID, ingestHeadroomConsumed(key.Budget, fixture.Amount))
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
