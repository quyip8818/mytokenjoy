package org

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// remoteIDNamespace is a fixed namespace UUID for deterministic ID generation from external IDs.
// NOTE: This happens to be RFC 4122 NamespaceDNS. It was chosen arbitrarily as a stable seed
// and MUST NOT be changed — doing so would produce different UUIDs for existing synced entities.
var remoteIDNamespace = uuid.MustParse("6ba7b814-9dad-11d1-80b4-00c04fd430c8")

func LocalDeptID(platform types.Platform, externalID string) uuid.UUID {
	return uuid.NewSHA1(remoteIDNamespace, []byte("dept-"+string(platform)+"-"+externalID))
}

func LocalMemberID(platform types.Platform, externalID string) uuid.UUID {
	return uuid.NewSHA1(remoteIDNamespace, []byte("m-"+string(platform)+"-"+externalID))
}

func IsManualDeptSource(source *string) bool {
	return source != nil && *source == types.DeptSourceManual
}

func IsManualMemberSource(source string) bool {
	return source == types.MemberSourceManual
}
