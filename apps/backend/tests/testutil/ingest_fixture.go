//go:build testhook

package testutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

// IngestBudgetFixture describes budget axes used when preparing ingest headroom in tests.
type IngestBudgetFixture struct {
	DepartmentID  uuid.UUID
	PlatformKeyID uuid.UUID
	MemberID      uuid.UUID // empty when ingest mapping has no member quota axis
	Amount        int64
}

func DefaultConsumeLogQuota() int64 {
	return 500_000
}

// PrepareIngestBudgetHeadroom sets snapshot consumed values so ingest projections
// do not immediately trigger overrun disable for the seeded demo budgets.
func PrepareIngestBudgetHeadroom(t *testing.T, st store.Store, fixture IngestBudgetFixture) {
	t.Helper()
	if fixture.DepartmentID == uuid.Nil {
		t.Fatal("PrepareIngestBudgetHeadroom: departmentID required")
	}
	if fixture.Amount <= 0 {
		fixture.Amount = DefaultConsumeLogQuota()
	}

	ctx := Ctx()
	if fixture.MemberID != uuid.Nil {
		quota, memberFound, err := st.Org().MemberPersonalBudget(ctx, fixture.MemberID)
		if err != nil {
			t.Fatal(err)
		}
		if memberFound {
			budgetfix.SetMemberSnapshotConsumed(t, st, fixture.MemberID, ingestHeadroomConsumed(quota, fixture.Amount))
		}
	}

	if fixture.PlatformKeyID == uuid.Nil {
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

func ingestHeadroomConsumed(limit, amount int64) int64 {
	consumed := limit - amount
	if consumed < 0 {
		return 0
	}
	return consumed
}
