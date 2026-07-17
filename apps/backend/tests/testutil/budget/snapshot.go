package budgetfix

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/points"
)

func ctx() context.Context {
	return company.DefaultContext(contract.DefaultCompanyID)
}

func DisplayPoints(display float64) float64 {
	return points.FromDisplay(display)
}

func SnapshotConsumed(t *testing.T, st store.Store, axisKind string, axisID uuid.UUID) float64 {
	t.Helper()
	return SnapshotConsumedAtPeriod(t, st, axisKind, axisID, contract.DemoBudgetPeriod)
}

func SnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind string, axisID uuid.UUID, periodKey string) float64 {
	t.Helper()
	ctx := ctx()
	consumed, found, err := st.BudgetConsumed().GetConsumed(ctx, axisKind, axisID, periodKey)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		return 0
	}
	return consumed
}

func PlatformKeySnapshotConsumed(t *testing.T, st store.Store, keyID uuid.UUID) float64 {
	t.Helper()
	return SnapshotConsumed(t, st, store.AxisKindPlatformKey, keyID)
}

func SetPlatformKeySnapshotConsumed(t *testing.T, st store.Store, keyID uuid.UUID, consumed float64) {
	t.Helper()
	SetSnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, keyID, contract.DemoBudgetPeriod, consumed)
}

func SetSnapshotConsumedAtPeriod(t *testing.T, st store.Store, axisKind string, axisID uuid.UUID, periodKey string, consumed float64) {
	t.Helper()
	ctx := ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, axisKind, axisID, periodKey, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetProjectSnapshotConsumed(t *testing.T, st store.Store, projectID uuid.UUID, consumed float64) {
	t.Helper()
	ctx := ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindProject, projectID, contract.DemoBudgetPeriod, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetMemberSnapshotConsumed(t *testing.T, st store.Store, memberID uuid.UUID, consumed float64) {
	t.Helper()
	ctx := ctx()
	if err := st.BudgetConsumed().SetConsumed(ctx, store.AxisKindMember, memberID, contract.DemoBudgetPeriod, consumed); err != nil {
		t.Fatal(err)
	}
}

func SetCombinedKeyRemain(t *testing.T, st store.Store, keyID uuid.UUID, remain float64) {
	t.Helper()
	ctx := ctx()
	if _, err := st.CombinedKeySummaries().UpdateBatch(ctx, []store.CombinedKeySummaryUpdate{
		{PlatformKeyID: keyID, Remain: remain},
	}); err != nil {
		t.Fatal(err)
	}
}

func CombinedKeyRemain(t *testing.T, st store.Store, keyID uuid.UUID) (remain *float64, version int64) {
	t.Helper()
	ctx := ctx()
	items, err := st.CombinedKeySummaries().ListByPlatformKeyIDs(ctx, []uuid.UUID{keyID})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatalf("combined key summary not found for key %s", keyID)
	}
	item := items[0]
	if item.Remain != 0 || item.Version > 0 {
		r := item.Remain
		return &r, item.Version
	}
	return nil, item.Version
}

func CombinedKeyRemainVersion(t *testing.T, st store.Store, keyID uuid.UUID) int64 {
	t.Helper()
	_, version := CombinedKeyRemain(t, st, keyID)
	return version
}
