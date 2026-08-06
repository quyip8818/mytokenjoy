package grants

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed manifest.json
var manifestJSON []byte

type PermissionMeta struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type Manifest struct {
	Version              int                       `json:"version"`
	Capabilities         []string                  `json:"capabilities"`
	PlatformCapabilities []string                  `json:"platformCapabilities"`
	PermissionIDMap      map[string]string         `json:"permissionIdMap"`
	PermissionNames      map[string]PermissionMeta `json:"permissionNames"`
	PresetRoles          map[string][]string       `json:"presetRoles"`
	WriteCapabilities    []string                  `json:"writeCapabilities"`
	Hierarchy            map[string][]string       `json:"hierarchy"`
}

var (
	manifestOnce sync.Once
	manifestData Manifest
	manifestErr  error
)

func ManifestData() (Manifest, error) {
	manifestOnce.Do(func() {
		if err := json.Unmarshal(manifestJSON, &manifestData); err != nil {
			manifestErr = fmt.Errorf("parse manifest: %w", err)
		}
	})
	return manifestData, manifestErr
}

func MustManifest() Manifest {
	m, err := ManifestData()
	if err != nil {
		panic(err)
	}
	return m
}

func PresetRoleCapabilities() map[string][]string {
	m := MustManifest()
	out := make(map[string][]string, len(m.PresetRoles))
	for name, caps := range m.PresetRoles {
		expanded := append([]string{}, caps...)
		if len(expanded) == 1 && expanded[0] == "*" {
			// "*" expands to Capabilities only (company-scoped permissions).
			// It intentionally excludes PlatformCapabilities (platform:*),
			// which must be granted explicitly via the 平台管理员 role.
			expanded = append([]string{}, m.Capabilities...)
		}
		out[name] = expanded
	}
	return out
}

func WriteCapabilitiesFromManifest() []string {
	return append([]string{}, MustManifest().WriteCapabilities...)
}

// ExpandHierarchy expands a set of permissions by applying the hierarchy rules
// from manifest. e.g. platform:admin implies platform:manage and platform:read.
// ponytail: simple iterative expansion, no recursion needed since hierarchy is max 2 levels deep.
// Upgrade path: if hierarchy grows deeper, switch to BFS.
func ExpandHierarchy(perms []string) []string {
	hierarchy := MustManifest().Hierarchy
	if len(hierarchy) == 0 {
		return perms
	}
	result := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		result[p] = struct{}{}
	}
	// Iterate until no new permissions are added (handles transitive chains like admin→manage→read).
	changed := true
	for changed {
		changed = false
		for p := range result {
			implied, ok := hierarchy[p]
			if !ok {
				continue
			}
			for _, ip := range implied {
				if _, exists := result[ip]; !exists {
					result[ip] = struct{}{}
					changed = true
				}
			}
		}
	}
	out := make([]string, 0, len(result))
	for p := range result {
		out = append(out, p)
	}
	return out
}
