package audit_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newAuditService(t *testing.T) audit.Service {
	t.Helper()
	cfg, st := testutil.NewTestStoreWithRuntimeSeed(t)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	return audit.NewService(cfg, st, reader)
}

func TestListOperationsPaginationAndActionFilter(t *testing.T) {
	t.Parallel()
	svc := newAuditService(t)
	ctx := testutil.Ctx()
	all, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total == 0 {
		t.Fatal("expected operation logs in seed")
	}
	if len(all.Items) == 0 {
		t.Fatal("expected paginated items")
	}

	firstAction := all.Items[0].Action
	filtered, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{
		Page: 1, PageSize: 20, Action: firstAction,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range filtered.Items {
		if item.Action != firstAction {
			t.Fatalf("expected action %s, got %s", firstAction, item.Action)
		}
	}
}

func TestListCallsDateFilter(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStoreWithRuntimeSeed(t)
	querier := domainusage.NewCallLogQuerier(st.Ledger())
	ctx := testutil.Ctx()
	const from = "2026-06-10"
	const to = "2026-06-15"
	filtered, err := querier.ListCalls(ctx, types.AuditCallsQueryParams{
		Page: 1, PageSize: 100, From: from, To: to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filtered.Total == 0 {
		t.Fatal("expected call logs in date range")
	}
	for _, item := range filtered.Items {
		day := item.CreatedAt
		if len(day) > 10 {
			day = day[:10]
		}
		if day < from || day > to {
			t.Fatalf("call %s createdAt %s outside [%s, %s]", item.ID, item.CreatedAt, from, to)
		}
	}
}
