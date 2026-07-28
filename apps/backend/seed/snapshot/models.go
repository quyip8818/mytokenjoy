package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	devMockEndpoint := "http://127.0.0.1:8765"
	return []types.ModelInfo{
		{ID: contract.IDModelTest, CompanyID: contract.TokenJoyCompanyID, Provider: types.ProviderCustom, Type: modelcatalog.TestCallType, Name: "Test Model", Description: "Local upstream for full-path ingest testing; echoes requested usage", Endpoint: &devMockEndpoint, MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: modelcatalog.SourceTest},
		{ID: contract.IDModel1, CompanyID: contract.DefaultCompanyID, Provider: "custom", Type: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: "seed"},
		{ID: contract.IDModel11, CompanyID: contract.DefaultCompanyID, Provider: "custom", Type: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: "seed"},
	}
}
