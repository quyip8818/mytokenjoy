package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
)

type RelayMappingOpts struct {
	PlatformKeyID string
	NewAPITokenID int64
	MemberID      string
	NoMember      bool
	DepartmentID  string
	RelayGroup    string
}

func DefaultRelayMappingOpts() RelayMappingOpts {
	return RelayMappingOpts{
		PlatformKeyID: seed.IDPlatformKey1,
		NewAPITokenID: 99,
		MemberID:      seed.IDMember1,
		DepartmentID:  seed.IDDept3,
		RelayGroup:    "dept-" + seed.IDDept3,
	}
}

func UpsertRelayMapping(t *testing.T, st store.Store, opts RelayMappingOpts) {
	t.Helper()
	if opts.PlatformKeyID == "" {
		opts = DefaultRelayMappingOpts()
	}
	if opts.RelayGroup == "" {
		opts.RelayGroup = "dept-" + opts.DepartmentID
	}
	var memberID *string
	if !opts.NoMember {
		m := opts.MemberID
		if m == "" {
			m = seed.IDMember1
		}
		memberID = &m
	}
	tokenID := opts.NewAPITokenID
	if err := st.Relay().UpsertMapping(Ctx(), store.RelayMapping{
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
