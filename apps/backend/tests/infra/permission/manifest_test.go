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
	if m.Version != 2 {
		t.Fatalf("version: got %d want 2", m.Version)
	}
	if len(m.Capabilities) != 19 {
		t.Fatalf("capabilities: got %d want 19", len(m.Capabilities))
	}
	if len(m.PlatformCapabilities) != 3 {
		t.Fatalf("platformCapabilities: got %d want 3", len(m.PlatformCapabilities))
	}
	if len(m.PermissionIDMap) != 19 {
		t.Fatalf("permissionIdMap: got %d want 19", len(m.PermissionIDMap))
	}
	if len(m.PresetRoles) != 7 {
		t.Fatalf("presetRoles: got %d want 7", len(m.PresetRoles))
	}
	if len(m.WriteCapabilities) == 0 {
		t.Fatal("writeCapabilities empty")
	}
	if len(m.Hierarchy) == 0 {
		t.Fatal("hierarchy empty")
	}
}

func TestGeneratedKeysMatchManifest(t *testing.T) {
	m := permission.MustManifest()

	// CompanyPermissions must match capabilities exactly.
	if len(permission.CompanyPermissions) != len(m.Capabilities) {
		t.Fatalf("CompanyPermissions length: got %d want %d", len(permission.CompanyPermissions), len(m.Capabilities))
	}
	for _, cap := range m.Capabilities {
		if !contains(permission.CompanyPermissions, cap) {
			t.Fatalf("missing capability in CompanyPermissions: %s", cap)
		}
	}

	// AllPermissions must equal capabilities + platformCapabilities.
	expectedAll := len(m.Capabilities) + len(m.PlatformCapabilities)
	if len(permission.AllPermissions) != expectedAll {
		t.Fatalf("AllPermissions length: got %d want %d", len(permission.AllPermissions), expectedAll)
	}
	for _, cap := range m.PlatformCapabilities {
		if !contains(permission.AllPermissions, cap) {
			t.Fatalf("missing platform capability in AllPermissions: %s", cap)
		}
	}

	// PermissionIDMap entries should all resolve to company permissions.
	for id, cap := range m.PermissionIDMap {
		if permission.PermissionIDMap[id] != cap {
			t.Fatalf("permission id %s: got %q want %q", id, permission.PermissionIDMap[id], cap)
		}
	}
}

func TestFrontendPermissionKeysMatchManifest(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "..")
	keysPath := filepath.Join(root, "apps", "frontend", "src", "lib", "permission-keys.ts")

	if _, err := os.Stat(keysPath); err != nil {
		t.Fatal(err)
	}

	m := permission.MustManifest()
	keysContent, err := os.ReadFile(keysPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(keysContent)

	// All capabilities and platform capabilities should be in the frontend keys.
	for _, cap := range m.Capabilities {
		if !strings.Contains(content, "'"+cap+"'") {
			t.Fatalf("permission-keys.ts missing %s", cap)
		}
	}
	for _, cap := range m.PlatformCapabilities {
		if !strings.Contains(content, "'"+cap+"'") {
			t.Fatalf("permission-keys.ts missing platform capability %s", cap)
		}
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
