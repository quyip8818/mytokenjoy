//go:build testhook

package relayfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

type MappingOpts struct {
	PlatformKeyID string
	NewAPITokenID int64
	MemberID      string
	NoMember      bool
	DepartmentID  string
	RelayGroup    string
}

func DefaultMappingOpts() MappingOpts {
	return MappingOpts{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: 99,
		MemberID:      contract.IDMember1,
		DepartmentID:  contract.IDDept3,
		RelayGroup:    "dept-" + contract.IDDept3,
	}
}

func UpsertMapping(t *testing.T, st store.Store, opts MappingOpts) {
	t.Helper()
	if opts.PlatformKeyID == "" {
		opts = DefaultMappingOpts()
	}
	if opts.RelayGroup == "" {
		opts.RelayGroup = "dept-" + opts.DepartmentID
	}
	var memberID *string
	if !opts.NoMember {
		m := opts.MemberID
		if m == "" {
			m = contract.IDMember1
		}
		memberID = &m
	}
	tokenID := opts.NewAPITokenID
	if err := st.Relay().UpsertMapping(testutil.Ctx(), store.RelayMapping{
		PlatformKeyID: opts.PlatformKeyID,
		NewAPITokenID: &tokenID,
		MemberID:      memberID,
		DepartmentID:  opts.DepartmentID,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    opts.RelayGroup,
	}); err != nil {
		t.Fatal(err)
	}
}
