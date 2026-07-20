//go:build testhook

package budgetfix

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/runtime"
)

func SeedDeptOverrun(t *testing.T, st store.Store, deptID uuid.UUID, ledgerSpent int64) {
	t.Helper()
	seedDefaultMapping(t, st)
	ctx := company.DefaultContext(contract.DefaultCompanyID)
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, contract.DefaultCompanyID, nil)
	if err != nil || len(lots) == 0 {
		t.Fatal("expected recharge lot for ledger seed")
	}
	memberID := contract.IDMember1
	now := time.Now().UTC()
	entry := types.UsageLedgerEntry{
		ID:               uuid.Must(uuid.NewV7()),
		CompanyID:        contract.DefaultCompanyID,
		EventType:        types.EventTypeCallSettled,
		IdempotencyKey:   fmt.Sprintf("test:dept-overrun:%s:%d", deptID, ledgerSpent),
		LotID:            lots[0].ID,
		Amount:           ledgerSpent,
		DisplayAmount:    float64(ledgerSpent) / float64(common.DefaultQuotaPerUnit),
		BillingCurrency:  common.DefaultBillingCurrency,
		DepartmentID:     deptID,
		MemberID:         &memberID,
		PlatformKeyID:    contract.IDPlatformKey1,
		PlatformKeyScope: types.PlatformKeyScopeMember,
		Source:           types.SourceWebhook,
		OccurredAt:       now,
		PeriodKey:        contract.DemoBudgetPeriod,
		Model:            "gpt-4o",
		CallDetail: types.UsageCallDetail{
			Caller:     "test",
			CallerID:   memberID.String(),
			CallerType: types.CallerTypeMember,
			Status:     types.CallStatusSuccess,
		},
		CreatedAt: now,
	}
	if _, err := st.Ledger().InsertOnConflict(ctx, entry); err != nil {
		t.Fatal(err)
	}
}

func seedDefaultMapping(t *testing.T, st store.Store) {
	t.Helper()
	ctx := company.DefaultContext(contract.DefaultCompanyID)
	deptID := contract.IDDept3
	memberID := contract.IDMember1
	keyID := int64(99)
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   &keyID,
		MemberID:      &memberID,
		DepartmentID:  deptID,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   newapiunits.NewAPIGroupForDepartment(deptID),
	}); err != nil {
		t.Fatal(err)
	}
}
