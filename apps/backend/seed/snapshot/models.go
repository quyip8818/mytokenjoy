package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	devMockEndpoint := "http://127.0.0.1:8765"
	return []types.ModelInfo{
		{ID: contract.IDModelTest, CompanyID: contract.TokenJoyCompanyID, Provider: types.ProviderCustom, Type: modelcatalog.TestCallType, Name: "Test Model", Description: "Local upstream for full-path ingest testing; echoes requested usage", Endpoint: &devMockEndpoint, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat"}, Source: modelcatalog.SourceTest},
	}
}
