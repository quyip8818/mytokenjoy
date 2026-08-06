package grants

import (
	"fmt"
	"sort"
)

// Normalizer is kept for backward compatibility with callers that inject it.
// New code should call NormalizeGrantIDs / RoleGrantIDs directly.
type Normalizer interface {
	NormalizeGrantIDs(refs []string) ([]string, error)
	RoleGrantIDs(roleType, roleName string, refs []string) ([]string, error)
}

// grantNormalizer implements Normalizer by delegating to package-level functions.
type grantNormalizer struct{}

func NewGrantNormalizer() Normalizer {
	return grantNormalizer{}
}

func (grantNormalizer) NormalizeGrantIDs(refs []string) ([]string, error) {
	return NormalizeGrantIDs(refs)
}

func (grantNormalizer) RoleGrantIDs(roleType, roleName string, refs []string) ([]string, error) {
	return RoleGrantIDs(roleType, roleName, refs)
}

// --- Package-level functions ---

var capabilityPermissionID map[string]string

func init() {
	capabilityPermissionID = make(map[string]string, len(PermissionIDMap))
	for id, cap := range PermissionIDMap {
		capabilityPermissionID[cap] = id
	}
}

func CapabilityToPermissionID(cap string) (string, bool) {
	if _, ok := PermissionIDMap[cap]; ok {
		return cap, true
	}
	id, ok := capabilityPermissionID[cap]
	return id, ok
}

func AllPermissionIDs() []string {
	ids := make([]string, 0, len(PermissionIDMap))
	for id := range PermissionIDMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func PresetRolePermissionIDs(roleName string) ([]string, error) {
	caps, ok := PresetRoleCapabilities()[roleName]
	if !ok {
		return nil, fmt.Errorf("unknown preset role %q", roleName)
	}
	// Filter out platform-domain capabilities — they have no permission ID mapping
	// and are not part of the company-level grant system.
	filtered := make([]string, 0, len(caps))
	for _, c := range caps {
		if IsPlatformPermission(c) {
			continue
		}
		filtered = append(filtered, c)
	}
	return NormalizeGrantIDs(filtered)
}

func NormalizeGrantIDs(refs []string) ([]string, error) {
	if len(refs) == 0 {
		return nil, nil
	}
	ids := make(map[string]struct{})
	for _, ref := range refs {
		switch ref {
		case "*":
			for _, id := range AllPermissionIDs() {
				ids[id] = struct{}{}
			}
			continue
		}
		if _, ok := PermissionIDMap[ref]; ok {
			ids[ref] = struct{}{}
			continue
		}
		if id, ok := CapabilityToPermissionID(ref); ok {
			ids[id] = struct{}{}
			continue
		}
		return nil, fmt.Errorf("unknown permission grant %q", ref)
	}
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

func RoleGrantIDs(roleType, roleName string, refs []string) ([]string, error) {
	if roleType == "preset" {
		return PresetRolePermissionIDs(roleName)
	}
	return NormalizeGrantIDs(refs)
}
