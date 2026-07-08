package testutil

import (
	"testing"
	"time"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func demoBudgetPeriodKey() string {
	return pkgbudget.SnapshotKey(contract.DemoBudgetPeriod, time.Now().UTC())
}

func SetDeptSnapshotConsumed(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	periodKey := demoBudgetPeriodKey()
	if err := st.BudgetSnapshots().SetConsumed(ctx, store.SnapshotAxisOrgNode, deptID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}

func SnapshotConsumed(t *testing.T, st store.Store, axisKind, axisID string) float64 {
	t.Helper()
	ctx := Ctx()
	periodKey := demoBudgetPeriodKey()
	consumed, found, err := st.BudgetSnapshots().GetConsumed(ctx, axisKind, axisID, periodKey)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		return 0
	}
	return consumed
}

func SnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind, axisID, periodKey string) float64 {
	t.Helper()
	ctx := Ctx()
	consumed, found, err := st.BudgetSnapshots().GetConsumed(ctx, axisKind, axisID, periodKey)
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
	return SnapshotConsumed(t, st, store.SnapshotAxisOrgNode, contract.IDDept3)
}

func PlatformKeySnapshotUsed(t *testing.T, st store.Store, keyID string) float64 {
	t.Helper()
	return SnapshotConsumed(t, st, store.SnapshotAxisPlatformKey, keyID)
}

func SetPlatformKeySnapshotUsed(t *testing.T, st store.Store, keyID string, used float64) {
	t.Helper()
	SetSnapshotConsumedAtPeriod(t, st, store.SnapshotAxisPlatformKey, keyID, demoBudgetPeriodKey(), used)
}

func SetSnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind, axisID, periodKey string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	if err := st.BudgetSnapshots().SetConsumed(ctx, axisKind, axisID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetGroupSnapshotConsumed(t *testing.T, st store.Store, groupID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	periodKey := demoBudgetPeriodKey()
	if err := st.BudgetSnapshots().SetConsumed(ctx, store.SnapshotAxisBudgetGroup, groupID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetMemberSnapshotConsumed(t *testing.T, st store.Store, memberID string, consumed float64) {
	t.Helper()
	ctx := Ctx()
	periodKey := demoBudgetPeriodKey()
	if err := st.BudgetSnapshots().SetConsumed(ctx, store.SnapshotAxisMember, memberID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}
