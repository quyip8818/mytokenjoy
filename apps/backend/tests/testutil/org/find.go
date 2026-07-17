package orgfix

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// FindMember returns a pointer to the member with the given ID, or nil if not found.
func FindMember(members []types.Member, id uuid.UUID) *types.Member {
	for i := range members {
		if members[i].ID == id {
			return &members[i]
		}
	}
	return nil
}

// FindPlatformKey returns a pointer to the platform key with the given ID, or nil if not found.
func FindPlatformKey(keys []types.PlatformKey, id uuid.UUID) *types.PlatformKey {
	for i := range keys {
		if keys[i].ID == id {
			return &keys[i]
		}
	}
	return nil
}
