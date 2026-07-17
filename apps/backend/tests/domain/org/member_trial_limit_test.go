package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
)

func trialCtx() *ctxcompany.Info {
	return &ctxcompany.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeTrial,
		Status:    store.CompanyStatusActive,
	}
}

func TestTrialMemberLimitBlocks(t *testing.T) {
	t.Parallel()
	cfg, st, svc := orgfix.NewServiceFromStore(t, testutil.WithTrialMemberLimit(3))
	_ = cfg
	_ = st

	// Build a context with company type = trial
	ctx := company.WithContext(testutil.Ctx(), *trialCtx())

	// Seed: the default company already has members from bootstrap.
	// Count existing members to calibrate.
	page, err := svc.ListMembers(ctx, "", "", false, 1, 1000)
	if err != nil {
		t.Fatal(err)
	}
	existingCount := page.Total

	// The limit is 3 — if existing already >= 3, the next create should fail.
	if existingCount >= 3 {
		_, err := svc.CreateMember(ctx, types.Member{
			Name: "Over Limit", Phone: "13900009999", Email: "over@example.com",
			DepartmentID: contract.IDDept5,
		})
		if err == nil {
			t.Fatal("expected trial member limit error, got nil")
		}
		return
	}

	// Fill up to limit
	for i := existingCount; i < 3; i++ {
		_, err := svc.CreateMember(ctx, types.Member{
			Name: "Fill", Phone: "", Email: "",
			DepartmentID: contract.IDDept5,
		})
		if err != nil {
			t.Fatalf("expected create to succeed (filling up), got %v", err)
		}
	}

	// Now the next one should fail
	_, err = svc.CreateMember(ctx, types.Member{
		Name: "Over Limit", Phone: "13900008888", Email: "overlimit@example.com",
		DepartmentID: contract.IDDept5,
	})
	if err == nil {
		t.Fatal("expected trial member limit error, got nil")
	}
}

func TestTrialMemberLimitAllowsNonTrial(t *testing.T) {
	t.Parallel()
	cfg, st, svc := orgfix.NewServiceFromStore(t, testutil.WithTrialMemberLimit(1))
	_ = cfg
	_ = st

	// Standard company context (not trial) — should not be limited
	ctx := company.WithContext(testutil.Ctx(), ctxcompany.Info{
		CompanyID: contract.DefaultCompanyID,
		Type:      store.CompanyTypeSelfhosted,
		Status:    store.CompanyStatusActive,
	})

	// Should succeed even though limit is 1 and there are already members
	_, err := svc.CreateMember(ctx, types.Member{
		Name: "No Limit", Phone: "13900007777", Email: "nolimit@example.com",
		DepartmentID: contract.IDDept5,
	})
	if err != nil {
		t.Fatalf("non-trial should not be limited, got %v", err)
	}
}

func TestTrialBatchImportLimitBlocks(t *testing.T) {
	t.Parallel()
	cfg, st, svc := orgfix.NewServiceFromStore(t, testutil.WithTrialMemberLimit(2))
	_ = cfg
	_ = st

	ctx := company.WithContext(testutil.Ctx(), *trialCtx())

	// Try to batch import 100 rows — should be rejected (limit is 2 + existing > 2)
	rows := make([]types.BatchImportRow, 100)
	for i := range rows {
		rows[i] = types.BatchImportRow{Name: "Batch", DepartmentName: "技术部"}
	}
	_, err := svc.BatchImport(ctx, rows)
	if err == nil {
		t.Fatal("expected trial batch import limit error, got nil")
	}
}
