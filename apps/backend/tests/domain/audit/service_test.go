package audit_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newAuditService(t *testing.T) audit.Service {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	return audit.NewService(cfg, st)
}

func TestListOperationsPaginationAndActionFilter(t *testing.T) {
	svc := newAuditService(t)
	all := svc.ListOperations(types.AuditOperationsQueryParams{Page: 1, PageSize: 100})
	if all.Total == 0 {
		t.Fatal("expected operation logs in seed")
	}
	if len(all.Items) == 0 {
		t.Fatal("expected paginated items")
	}

	firstAction := all.Items[0].Action
	filtered := svc.ListOperations(types.AuditOperationsQueryParams{
		Page: 1, PageSize: 20, Action: firstAction,
	})
	for _, item := range filtered.Items {
		if item.Action != firstAction {
			t.Fatalf("expected action %s, got %s", firstAction, item.Action)
		}
	}
}

func TestListCallsDateFilter(t *testing.T) {
	svc := newAuditService(t)
	const from = "2026-06-10"
	const to = "2026-06-15"
	filtered := svc.ListCalls(types.AuditCallsQueryParams{
		Page: 1, PageSize: 100, From: from, To: to,
	})
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
