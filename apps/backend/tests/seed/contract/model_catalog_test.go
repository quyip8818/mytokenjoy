package contract_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
)

func TestSeedModelCatalogIDs(t *testing.T) {
	t.Parallel()

	if contract.ModelTypeToID["local-test-model"] != contract.IDModelLocalTest {
		t.Fatalf("expected local-test-model id %d, got %d", contract.IDModelLocalTest, contract.ModelTypeToID["local-test-model"])
	}
	if contract.ModelTypeToID["gpt-4o"] != contract.IDModel10 {
		t.Fatalf("expected gpt-4o id %d, got %d", contract.IDModel10, contract.ModelTypeToID["gpt-4o"])
	}
	if contract.IDModel1 < contract.ProdCatalogModelIDStart {
		t.Fatalf("production catalog ids must start at %d", contract.ProdCatalogModelIDStart)
	}
}
