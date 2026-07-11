package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

type recordingSyncService struct {
	companyIDs []int64
}

func (s *recordingSyncService) RunScheduledSync(ctx context.Context) error {
	s.companyIDs = append(s.companyIDs, company.CompanyID(ctx))
	return nil
}

func TestOrgSyncRunsForEveryActiveCompany(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	if err := st.Company().Create(ctx, store.Company{
		ID: 1_000_001, Slug: "co-b", Name: "Company B", Status: store.CompanyStatusActive,
		BillingCurrency: "CNY", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Company().Create(ctx, store.Company{
		ID: 1_000_002, Slug: "co-suspended", Name: "Suspended", Status: store.CompanyStatusSuspended,
		BillingCurrency: "CNY", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	recording := &recordingSyncService{}
	if err := company.ForEachActiveCompany(ctx, st.Company(), func(companyCtx context.Context, _ store.Company) error {
		return recording.RunScheduledSync(companyCtx)
	}); err != nil {
		t.Fatal(err)
	}
	if len(recording.companyIDs) != 3 {
		t.Fatalf("expected sync for tokenjoy + default + co-b, got %v", recording.companyIDs)
	}
	seen := map[int64]bool{}
	for _, id := range recording.companyIDs {
		seen[id] = true
	}
	for _, want := range []int64{contract.TokenJoyCompanyID, contract.DefaultCompanyID, 1_000_001} {
		if !seen[want] {
			t.Fatalf("expected company %d in sync list, got %v", want, recording.companyIDs)
		}
	}
	if seen[1_000_002] {
		t.Fatalf("suspended company should not sync, got %v", recording.companyIDs)
	}
}
