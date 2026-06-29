package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/internal/store"
)

type UsageBucketOpts struct {
	BucketStart  time.Time
	DepartmentID string
	MemberID     string
	Model        string
	CostCNY      float64
	CallCount    int
}

func DefaultUsageBucketOpts() UsageBucketOpts {
	return UsageBucketOpts{
		BucketStart:  time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC),
		DepartmentID: seed.IDDept3,
		MemberID:     seed.IDMember1,
		Model:        "gpt-4o",
		CostCNY:      1,
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
	if err := st.Usage().UpsertBucket(context.Background(), types.UsageBucketRow{
		BucketStart: opts.BucketStart, DepartmentID: opts.DepartmentID, MemberID: opts.MemberID,
		Model: opts.Model, CostCNY: opts.CostCNY, CallCount: opts.CallCount,
	}); err != nil {
		t.Fatal(err)
	}
}

func UsageBucketCount(st store.Store) int {
	mem, ok := st.(*store.Memory)
	if !ok {
		return -1
	}
	return len(mem.UsageBucketRows())
}

func AssertUsageBucketCount(t *testing.T, st store.Store, want int) {
	t.Helper()
	got := UsageBucketCount(st)
	if got < 0 {
		t.Fatal("usage bucket count requires memory store in unit tests")
	}
	if got != want {
		t.Fatalf("expected %d usage buckets, got %d", want, got)
	}
}

func AdminDashboardScope() domainusage.SessionScope {
	return domainusage.SessionScope{
		MemberID: seed.IDMemberAdmin, Permissions: []string{permission.DashboardCost, "*"},
	}
}
