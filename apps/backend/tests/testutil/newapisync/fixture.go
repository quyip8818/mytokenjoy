//go:build testhook

package newapisync

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

// PrepareIngestFixture upserts a platform key mapping and prepares budget snapshot
// headroom for a successful ingest. Use this instead of UpsertMapping when the
// test expects ingest to succeed without immediate overrun; demo seed dept-3 is intentionally
// over budget.
func PrepareIngestFixture(t *testing.T, st store.Store, opts MappingOpts, amount ...float64) {
	t.Helper()
	if opts.PlatformKeyID == uuid.Nil {
		opts = DefaultMappingOpts()
	}
	if opts.DepartmentID == uuid.Nil {
		opts.DepartmentID = contract.IDDept3
	}

	UpsertMapping(t, st, opts)

	ingestAmount := testutil.DefaultConsumeLogQuota()
	if len(amount) > 0 && amount[0] > 0 {
		ingestAmount = amount[0]
	}

	fixture := testutil.IngestBudgetFixture{
		DepartmentID:  opts.DepartmentID,
		PlatformKeyID: opts.PlatformKeyID,
		Amount:        ingestAmount,
	}
	if !opts.NoMember {
		fixture.MemberID = opts.MemberID
		if fixture.MemberID == uuid.Nil {
			fixture.MemberID = contract.IDMember1
		}
	}
	testutil.PrepareIngestBudgetHeadroom(t, st, fixture)
}
