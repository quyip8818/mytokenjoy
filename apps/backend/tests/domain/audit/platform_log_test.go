package audit_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestAppendPlatformOperationLogWritesToTargetCompany(t *testing.T) {
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	svc := audit.NewService(cfg, st)

	const targetCompanyID int64 = 2
	const action = "platform.company.recharge"

	if err := company.AppendPlatformOperationLog(context.Background(), st, targetCompanyID, action, "op-1", "company:2", "amount=10"); err != nil {
		t.Fatal(err)
	}

	targetCtx := company.DefaultContext(targetCompanyID)
	result, err := svc.ListOperations(targetCtx, types.AuditOperationsQueryParams{Page: 1, PageSize: 100, Action: action})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total == 0 {
		t.Fatal("expected operation log under target company")
	}

	otherCtx := company.DefaultContext(1)
	other, err := svc.ListOperations(otherCtx, types.AuditOperationsQueryParams{Page: 1, PageSize: 100, Action: action})
	if err != nil {
		t.Fatal(err)
	}
	if other.Total != 0 {
		t.Fatalf("expected no platform recharge logs in company 1, got %d", other.Total)
	}
}
