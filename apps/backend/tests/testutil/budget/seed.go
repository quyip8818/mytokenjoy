//go:build testhook

package budgetfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func SeedDeptOverrun(t *testing.T, st store.Store, deptID string, consumed float64) {
	t.Helper()
	seedDefaultMapping(t, st)
	SetDeptSnapshotConsumed(t, st, deptID, consumed)
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
