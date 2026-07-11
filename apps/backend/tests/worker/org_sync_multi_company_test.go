package worker_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

type recordingSyncService struct {
	companyIDs []int64
}

func (s *recordingSyncService) GetSyncConfig(context.Context) (types.SyncConfig, error) {
	return types.SyncConfig{}, nil
}

func (s *recordingSyncService) UpdateSyncConfig(context.Context, types.SyncConfig) error {
	return nil
}

func (s *recordingSyncService) TriggerSync(context.Context) (types.ImportResult, error) {
	return types.ImportResult{}, nil
}

func (s *recordingSyncService) RunScheduledSync(ctx context.Context) error {
	s.companyIDs = append(s.companyIDs, company.CompanyID(ctx))
	return nil
}

func (s *recordingSyncService) ListSyncLogs(context.Context, int, int) (types.PageResult[types.SyncLog], error) {
	return types.PageResult[types.SyncLog]{}, nil
}

func TestOrgSyncRunsForEveryActiveCompany(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
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
	stub := &mock.StubAdminClient{Token: newapi.Token{ID: 99, RemainQuota: 1000}}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg, err := app.BuildRegistry(cfg, logger, st, app.WithAdminClient(stub), app.WithOrgSync(recording))
	if err != nil {
		t.Fatal(err)
	}
	runner := reg.WorkerRunner(logger)

	if err := runner.RunOrgSyncOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if len(recording.companyIDs) != 3 {
		t.Fatalf("expected sync for tokenjoy + default + co-b, got %v", recording.companyIDs)
	}
	seen := map[int64]bool{}
	for _, id := range recording.companyIDs {
		seen[id] = true
	}
	if !seen[contract.TokenJoyCompanyID] || !seen[contract.DefaultCompanyID] || !seen[int64(1_000_001)] {
		t.Fatalf("expected tokenjoy, default company and co-b, got %v", recording.companyIDs)
	}
	if seen[int64(1_000_002)] {
		t.Fatal("suspended company should not be synced")
	}
}
