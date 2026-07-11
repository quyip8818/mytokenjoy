package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/points"
)

func DisplayPoints(display float64) float64 {
	return points.FromDisplay(display)
}

func SetDeptSnapshotConsumed(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindOrgNode, deptID, contract.DemoBudgetPeriod, consumed); err != nil {
		t.Fatal(err)
	}
}

func SnapshotConsumed(t *testing.T, st store.Store, axisKind, axisID string) float64 {
	t.Helper()
	return SnapshotConsumedAtPeriod(t, st, axisKind, axisID, contract.DemoBudgetPeriod)
}

func SnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind, axisID, periodKey string) float64 {
	t.Helper()
	ctx := Ctx()
	consumed, found, err := st.BudgetConsumed().GetConsumed(ctx, axisKind, axisID, periodKey)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		return 0
	}
	return consumed
}

func Dept3SnapshotConsumed(t *testing.T, st store.Store) float64 {
	t.Helper()
	return SnapshotConsumed(t, st, store.AxisKindOrgNode, contract.IDDept3)
}

func PlatformKeySnapshotUsed(t *testing.T, st store.Store, keyID string) float64 {
	t.Helper()
	return SnapshotConsumed(t, st, store.AxisKindPlatformKey, keyID)
}

func SetPlatformKeySnapshotUsed(t *testing.T, st store.Store, keyID string, used float64) {
	t.Helper()
	SetSnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, keyID, contract.DemoBudgetPeriod, used)
}

func SetSnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind, axisID, periodKey string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, axisKind, axisID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetGroupSnapshotConsumed(t *testing.T, st store.Store, groupID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindBudgetGroup, groupID, contract.DemoBudgetPeriod, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetMemberSnapshotConsumed(t *testing.T, st store.Store, memberID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindMember, memberID, contract.DemoBudgetPeriod, consumed); err != nil {
		t.Fatal(err)
	}
}
