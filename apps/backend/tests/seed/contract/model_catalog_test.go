package contract_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
)

func TestSeedModelCatalogIDs(t *testing.T) {
	t.Parallel()

	if contract.ModelTypeToID["dev-local-test"] != contract.IDModelLocalTest {
		t.Fatalf("expected dev-local-test id %s, got %s", contract.IDModelLocalTest, contract.ModelTypeToID["dev-local-test"])
	}
	if contract.ModelTypeToID["gpt-4o"] != contract.IDModel10 {
		t.Fatalf("expected gpt-4o id %s, got %s", contract.IDModel10, contract.ModelTypeToID["gpt-4o"])
	}
	// Verify all expected model types are present in the map.
	expectedTypes := []string{"deepseek-v4", "deepseek-r1", "qwen-3.5-plus", "qwen-max-2026", "glm-5", "kimi-k3", "doubao-pro-256k", "minimax-m2", "claude-sonnet-5", "gpt-4o", "dev-local-test"}
	for _, mt := range expectedTypes {
		if _, ok := contract.ModelTypeToID[mt]; !ok {
			t.Errorf("ModelTypeToID missing expected type %q", mt)
		}
	}
}
