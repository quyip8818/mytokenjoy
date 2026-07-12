//go:build testhook

package newapisync

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

type MappingOpts struct {
	PlatformKeyID string
	NewAPIKeyID   int64
	MemberID      string
	NoMember      bool
	DepartmentID  string
	NewAPIGroup   string
}

func DefaultMappingOpts() MappingOpts {
	return MappingOpts{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPIKeyID:   99,
		MemberID:      contract.IDMember1,
		DepartmentID:  contract.IDDept3,
		NewAPIGroup:   newapiunits.NewAPIGroupForDepartment(contract.IDDept3),
	}
}

func UpsertMapping(t *testing.T, st store.Store, opts MappingOpts) {
	t.Helper()
	if opts.PlatformKeyID == "" {
		opts = DefaultMappingOpts()
	}
	if opts.DepartmentID == "" {
		opts.DepartmentID = contract.IDDept3
	}
	if opts.NewAPIGroup == "" {
		opts.NewAPIGroup = newapiunits.NewAPIGroupForDepartment(opts.DepartmentID)
	}
	var memberID *string
	if !opts.NoMember {
		m := opts.MemberID
		if m == "" {
			m = contract.IDMember1
		}
		memberID = &m
	}
	keyID := opts.NewAPIKeyID
	if err := st.PlatformKeyMappings().UpsertMapping(testutil.Ctx(), store.PlatformKeyMapping{
		PlatformKeyID: opts.PlatformKeyID,
		NewAPIKeyID:   &keyID,
		MemberID:      memberID,
		DepartmentID:  opts.DepartmentID,
		SyncStatus:    store.MappingSyncStatusSynced,
		NewAPIGroup:   opts.NewAPIGroup,
	}); err != nil {
		t.Fatal(err)
	}
}
