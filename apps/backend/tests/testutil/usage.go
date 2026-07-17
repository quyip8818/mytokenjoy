//go:build testhook

package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

type UsageBucketOpts struct {
	BucketStart  time.Time
	DepartmentID uuid.UUID
	MemberID     uuid.UUID
	Model        string
	Cost         float64
	DisplayCost  float64
	CallCount    int
}

func DefaultUsageBucketOpts() UsageBucketOpts {
	return UsageBucketOpts{
		BucketStart:  time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC),
		DepartmentID: contract.IDDept3,
		MemberID:     contract.IDMember1,
		Model:        "gpt-4o",
		Cost:         1,
		DisplayCost:  1,
		CallCount:    1,
	}
}

func SeedUsageBucket(t *testing.T, st store.Store, opts UsageBucketOpts) {
	t.Helper()
	def := DefaultUsageBucketOpts()
	if opts.BucketStart.IsZero() {
		opts.BucketStart = def.BucketStart
	}
	if opts.DepartmentID == "" {
		opts.DepartmentID = def.DepartmentID
	}
	if opts.MemberID == "" {
		opts.MemberID = def.MemberID
	}
	if opts.Model == "" {
		opts.Model = def.Model
	}
	if opts.CallCount == 0 {
		opts.CallCount = def.CallCount
	}
	if opts.DisplayCost == 0 && opts.Cost != 0 {
		opts.DisplayCost = opts.Cost
	}
	if err := st.Usage().UpsertBucket(Ctx(), types.UsageBucketRow{
		BucketStart: opts.BucketStart, DepartmentID: opts.DepartmentID, MemberID: opts.MemberID,
		Model: opts.Model, Cost: opts.Cost, DisplayCost: opts.DisplayCost, CallCount: opts.CallCount,
	}); err != nil {
		t.Fatal(err)
	}
}

func UsageBucketCount(st store.Store) int {
	return len(UsageBucketRows(st))
}

func HasLedgerLogID(st store.Store, logID int64) (bool, error) {
	entries, _, err := st.Ledger().ListCallSettledPage(Ctx(), store.LedgerCallFilter{Page: 1, PageSize: 10000})
	if err != nil {
		return false, err
	}
	key := domainusage.NewAPIIdempotencyKey(logID)
	for _, entry := range entries {
		if entry.IdempotencyKey == key {
			return true, nil
		}
	}
	return false, nil
}

func AssertUsageBucketCount(t *testing.T, st store.Store, want int) {
	t.Helper()
	got := UsageBucketCount(st)
	if got != want {
		t.Fatalf("expected %d usage buckets, got %d", want, got)
	}
}

func AdminDashboardScope() domainusage.SessionScope {
	return domainusage.SessionScope{
		MemberID: contract.IDMemberAdmin, Permissions: []string{permission.DashboardCost, "*"},
	}
}
