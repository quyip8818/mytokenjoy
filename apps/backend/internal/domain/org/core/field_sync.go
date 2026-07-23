package core

import "slices"

// FieldSyncMode defines how a member field behaves during sync.
type FieldSyncMode int

const (
	// SyncModeImmutable: written once on first import, never overwritten.
	SyncModeImmutable FieldSyncMode = iota
	// SyncModeUserOwned: sync overwrites unless user/admin has manually edited (tracked via OverrideFields).
	SyncModeUserOwned
	// SyncModeAlways: remote source is always authoritative, unconditional overwrite.
	SyncModeAlways
	// SyncModeLocalOnly: never touched by sync.
	SyncModeLocalOnly
)

// FieldSyncPolicy maps member field names to their sync mode.
// Field names match the JSON/struct field names used in OverrideFields tracking.
var FieldSyncPolicy = map[string]FieldSyncMode{
	"employeeId":     SyncModeImmutable,
	"hireDate":       SyncModeImmutable,
	"alias":          SyncModeUserOwned,
	"avatar":         SyncModeUserOwned,
	"jobTitle":       SyncModeAlways,
	"departmentId":   SyncModeAlways,
	"departmentName": SyncModeAlways,
	"roles":          SyncModeLocalOnly,
	"status":         SyncModeLocalOnly,
	"personalBudget": SyncModeLocalOnly,
}

// ShouldSyncField determines whether a remote value should overwrite the local field.
//
//   - localValue: the current local value (as string; empty means unset)
//   - overrideFields: the member's current OverrideFields list
//   - fieldName: must be a key in FieldSyncPolicy
//
// Returns true if the remote value should be written to local.
func ShouldSyncField(fieldName string, localValue string, overrideFields []string) bool {
	mode, ok := FieldSyncPolicy[fieldName]
	if !ok {
		// Unknown field: don't touch.
		return false
	}
	switch mode {
	case SyncModeAlways:
		return true
	case SyncModeImmutable:
		return localValue == ""
	case SyncModeUserOwned:
		return !slices.Contains(overrideFields, fieldName)
	case SyncModeLocalOnly:
		return false
	default:
		return false
	}
}

// UserOwnedFields returns all field names with SyncModeUserOwned.
func UserOwnedFields() []string {
	var fields []string
	for name, mode := range FieldSyncPolicy {
		if mode == SyncModeUserOwned {
			fields = append(fields, name)
		}
	}
	return fields
}

// TrackOverride adds fieldName to overrideFields if it's a user-owned field
// and not already present. Returns the updated slice.
func TrackOverride(overrideFields []string, fieldName string) []string {
	mode, ok := FieldSyncPolicy[fieldName]
	if !ok || mode != SyncModeUserOwned {
		return overrideFields
	}
	if slices.Contains(overrideFields, fieldName) {
		return overrideFields
	}
	return append(overrideFields, fieldName)
}
