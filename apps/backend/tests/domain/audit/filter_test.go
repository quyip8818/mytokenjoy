package audit_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/tests/testutil"
)

// PRD 7.1: 操作审计 — 按操作类型、时间范围、操作人、关键词筛选
// PRD 7.2: 调用日志 — 按状态、模型、调用人、时间范围、关键词筛选

func TestListOperationsKeywordFilter(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := audit.NewService(cfg, st)
	ctx := testutil.Ctx()

	// Get all operations first
	all, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total == 0 {
		t.Skip("no seed operation logs")
	}

	// Filter by keyword from first entry
	keyword := all.Items[0].Target
	if keyword == "" {
		t.Skip("first entry has no target for keyword test")
	}
	filtered, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{
		Page: 1, PageSize: 100, Keyword: keyword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filtered.Total == 0 {
		t.Errorf("expected at least 1 result for keyword %q", keyword)
	}
}

func TestListOperationsTimeRangeFilter(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := audit.NewService(cfg, st)
	ctx := testutil.Ctx()

	// Very narrow range that should exclude everything
	result, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{
		Page: 1, PageSize: 100, From: "2020-01-01", To: "2020-01-02",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 results for far-past range, got %d", result.Total)
	}
}

func TestListOperationsOperatorFilter(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := audit.NewService(cfg, st)
	ctx := testutil.Ctx()

	// Filter by a non-existent operator
	result, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{
		Page: 1, PageSize: 100, OperatorID: "nonexistent-operator",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 results for unknown operator, got %d", result.Total)
	}
}

func TestListOperationsPagination(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := audit.NewService(cfg, st)
	ctx := testutil.Ctx()

	// Get total count
	all, _ := svc.ListOperations(ctx, types.AuditOperationsQueryParams{Page: 1, PageSize: 100})
	if all.Total <= 1 {
		t.Skip("need multiple logs for pagination test")
	}

	// Page 1 with size 1
	page1, err := svc.ListOperations(ctx, types.AuditOperationsQueryParams{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Items) != 1 {
		t.Fatalf("expected 1 item on page 1, got %d", len(page1.Items))
	}
	if page1.Total != all.Total {
		t.Errorf("total should be %d regardless of page size, got %d", all.Total, page1.Total)
	}
}

func TestListCallsModelFilter(t *testing.T) {
	_, st := testutil.NewTestStore(t)
	querier := domainusage.NewCallLogQuerier(st.Ledger())
	ctx := testutil.Ctx()

	// Get all calls first
	all, err := querier.ListCalls(ctx, types.AuditCallsQueryParams{Page: 1, PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total == 0 {
		t.Skip("no seed call logs")
	}

	// Filter by model of first entry
	model := all.Items[0].Model
	filtered, err := querier.ListCalls(ctx, types.AuditCallsQueryParams{
		Page: 1, PageSize: 100, Model: model,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range filtered.Items {
		if item.Model != model {
			t.Errorf("expected model %q, got %q", model, item.Model)
		}
	}
}

func TestListCallsStatusFilter(t *testing.T) {
	_, st := testutil.NewTestStore(t)
	querier := domainusage.NewCallLogQuerier(st.Ledger())
	ctx := testutil.Ctx()

	// Filter by success status
	filtered, err := querier.ListCalls(ctx, types.AuditCallsQueryParams{
		Page: 1, PageSize: 100, Status: "success",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range filtered.Items {
		if item.Status != "success" {
			t.Errorf("expected status 'success', got %q", item.Status)
		}
	}
}

func TestListCallsPagination(t *testing.T) {
	_, st := testutil.NewTestStore(t)
	querier := domainusage.NewCallLogQuerier(st.Ledger())
	ctx := testutil.Ctx()

	all, _ := querier.ListCalls(ctx, types.AuditCallsQueryParams{Page: 1, PageSize: 100})
	if all.Total <= 1 {
		t.Skip("need multiple call logs for pagination test")
	}

	page1, err := querier.ListCalls(ctx, types.AuditCallsQueryParams{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Items) != 1 {
		t.Fatalf("expected 1 item on page 1, got %d", len(page1.Items))
	}
	if page1.Total != all.Total {
		t.Errorf("total mismatch: %d vs %d", page1.Total, all.Total)
	}
}
