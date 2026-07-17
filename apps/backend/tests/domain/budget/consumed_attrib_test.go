package budget_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func TestConsumptionDeltas_MemberScope(t *testing.T) {
	t.Parallel()
	memberID := uuid.MustParse("00000000-0000-7000-0000-00000000ee01")
	pkID := uuid.MustParse("00000000-0000-7000-0000-00000000f001")
	entry := types.UsageLedgerEntry{
		PlatformKeyID:    pkID,
		PlatformKeyScope: types.PlatformKeyScopeMember,
		MemberID:         &memberID,
		Amount:           42.5,
	}
	open := pkgbudget.TestOpenBudgetPeriod("2026-07")
	deltas, err := budget.ConsumptionDeltas(context.Background(), nil, entry, open)
	if err != nil {
		t.Fatal(err)
	}
	if len(deltas) != 2 {
		t.Fatalf("expected 2 deltas, got %d", len(deltas))
	}
	if deltas[0].Kind != store.AxisKindPlatformKey || deltas[0].AxisID != pkID {
		t.Errorf("delta[0] = %+v", deltas[0])
	}
	if deltas[0].Amount != 42.5 || deltas[0].PeriodKey != "2026-07" {
		t.Errorf("delta[0] amount/period = %v/%v", deltas[0].Amount, deltas[0].PeriodKey)
	}
	if deltas[1].Kind != store.AxisKindMember || deltas[1].AxisID != memberID {
		t.Errorf("delta[1] = %+v", deltas[1])
	}
}

func TestConsumptionDeltas_ProjectScope(t *testing.T) {
	t.Parallel()
	projectID := uuid.MustParse("00000000-0000-7000-0000-000000000a01")
	pkID2 := uuid.MustParse("00000000-0000-7000-0000-00000000f002")
	entry := types.UsageLedgerEntry{
		PlatformKeyID:    pkID2,
		PlatformKeyScope: types.PlatformKeyScopeProject,
		ProjectID:        &projectID,
		Amount:           10,
	}
	open := pkgbudget.TestOpenBudgetPeriod("2026-06")
	deltas, err := budget.ConsumptionDeltas(context.Background(), nil, entry, open)
	if err != nil {
		t.Fatal(err)
	}
	if len(deltas) != 2 {
		t.Fatalf("expected 2 deltas, got %d", len(deltas))
	}
	if deltas[1].Kind != store.AxisKindProject || deltas[1].AxisID != projectID {
		t.Errorf("delta[1] = %+v", deltas[1])
	}
}

func TestConsumptionDeltas_ZeroPeriodError(t *testing.T) {
	t.Parallel()
	entry := types.UsageLedgerEntry{
		PlatformKeyID:    uuid.MustParse("00000000-0000-7000-0000-00000000f001"),
		PlatformKeyScope: types.PlatformKeyScopeMember,
		Amount:           10,
	}
	_, err := budget.ConsumptionDeltas(context.Background(), nil, entry, pkgbudget.OpenBudgetPeriod{})
	if err == nil {
		t.Error("expected error for zero open period")
	}
}
