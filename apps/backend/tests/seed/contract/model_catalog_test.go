package contract_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
)

func TestSeedModelCatalogIDs(t *testing.T) {
	t.Parallel()

	if contract.ModelTypeToID["test-model"] != contract.IDModelTest {
		t.Fatalf("expected test-model id %s, got %s", contract.IDModelTest, contract.ModelTypeToID["test-model"])
	}
	if contract.ModelTypeToID["deepseek-v4-pro"] != contract.IDModel1 {
		t.Fatalf("expected deepseek-v4-pro id %s, got %s", contract.IDModel1, contract.ModelTypeToID["deepseek-v4-pro"])
	}
	// Verify all expected model types are present in the map.
	expectedTypes := []string{"deepseek-v4-pro", "deepseek-v4-flash", "test-model"}
	for _, mt := range expectedTypes {
		if _, ok := contract.ModelTypeToID[mt]; !ok {
			t.Errorf("ModelTypeToID missing expected type %q", mt)
		}
	}
}
