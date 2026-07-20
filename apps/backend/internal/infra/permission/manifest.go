package permission

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
	Version                 int                       `json:"version"`
	Capabilities            []string                  `json:"capabilities"`
	PermissionIDMap         map[string]string         `json:"permissionIdMap"`
	PermissionNames         map[string]PermissionMeta `json:"permissionNames"`
	PresetRoles             map[string][]string       `json:"presetRoles"`
	WriteCapabilities       []string                  `json:"writeCapabilities"`
	BudgetWriteCapabilities []string                  `json:"budgetWriteCapabilities"`
	Rules                   struct {
		BudgetWriteImpliesRead bool `json:"budgetWriteImpliesRead"`
	} `json:"rules"`
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
			return
		}
		if manifestData.Rules.BudgetWriteImpliesRead && len(manifestData.BudgetWriteCapabilities) == 0 {
			manifestErr = fmt.Errorf("manifest: budgetWriteCapabilities required when budgetWriteImpliesRead")
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
			expanded = append([]string{}, m.Capabilities...)
		}
		out[name] = expanded
	}
	return out
}

func WriteCapabilitiesFromManifest() []string {
	return append([]string{}, MustManifest().WriteCapabilities...)
}

func BudgetWriteCapabilitiesFromManifest() []string {
	return append([]string{}, MustManifest().BudgetWriteCapabilities...)
}
