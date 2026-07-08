package permission

import (
	"fmt"
	"sort"
)

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
	return NormalizeGrantIDs(caps)
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
