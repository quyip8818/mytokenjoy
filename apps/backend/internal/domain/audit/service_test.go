package audit_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
)

func newAuditService(t *testing.T) audit.Service {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return audit.NewService(cfg, store.NewMemory(seed.Load(cfg)))
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
