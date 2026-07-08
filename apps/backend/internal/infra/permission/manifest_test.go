package permission_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/infra/permission"
)

func TestManifestLoads(t *testing.T) {
	m, err := permission.ManifestData()
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != 1 {
		t.Fatalf("version: got %d want 1", m.Version)
	}
	if len(m.Capabilities) != 23 {
		t.Fatalf("capabilities: got %d want 23", len(m.Capabilities))
	}
	if len(m.PermissionIDMap) != 23 {
		t.Fatalf("permissionIdMap: got %d want 23", len(m.PermissionIDMap))
	}
	if len(m.PresetRoles) != 5 {
		t.Fatalf("presetRoles: got %d want 5", len(m.PresetRoles))
	}
	if len(m.WriteCapabilities) == 0 {
		t.Fatal("writeCapabilities empty")
	}
	if len(m.BudgetWriteCapabilities) == 0 {
		t.Fatal("budgetWriteCapabilities empty")
	}
}

func TestGeneratedKeysMatchManifest(t *testing.T) {
	m := permission.MustManifest()
	if len(m.Capabilities) != len(permission.AllPermissions) {
		t.Fatalf("AllPermissions length: got %d want %d", len(permission.AllPermissions), len(m.Capabilities))
	}
	for _, cap := range m.Capabilities {
		found := false
		for _, p := range permission.AllPermissions {
			if p == cap {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing capability in AllPermissions: %s", cap)
		}
	}
	for id, cap := range m.PermissionIDMap {
		if permission.PermissionIDMap[id] != cap {
			t.Fatalf("permission id %s: got %q want %q", id, permission.PermissionIDMap[id], cap)
		}
	}
}

func TestFrontendPermissionKeysMatchManifest(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..")
	manifestPath := filepath.Join(root, "packages", "contracts", "permission", "manifest.json")
	keysPath := filepath.Join(root, "apps", "frontend", "src", "lib", "permission-keys.ts")

	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(keysPath); err != nil {
		t.Fatal(err)
	}

	m := permission.MustManifest()
	keysContent, err := os.ReadFile(keysPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(keysContent)
	for _, cap := range m.Capabilities {
		if !strings.Contains(content, "'"+cap+"'") {
			t.Fatalf("permission-keys.ts missing %s", cap)
		}
	}
}
