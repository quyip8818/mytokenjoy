package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	devMockEndpoint := "http://127.0.0.1:8765"
	return []types.ModelInfo{
		// Global platform models (managed by platform_admin, visible to all tenants).
		{ID: contract.IDModel2, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: "platform"},
		{ID: contract.IDModel10, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o", Name: "GPT-4o", MaxContext: 128000, Active: true, Capabilities: []string{"chat", "vision"}, Source: "platform"},
		// Tenant custom models.
		{ID: contract.IDModel1, CompanyID: contract.DefaultCompanyID, Provider: "custom", Type: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: "seed"},
		{ID: contract.IDModel11, CompanyID: contract.DefaultCompanyID, Provider: "custom", Type: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: "seed"},
		// Dev/test model: belongs to demo tenant, not platform. Gateway type-guard limits access to demo/trial/testing companies.
		{ID: contract.IDModelTest, CompanyID: contract.DefaultCompanyID, Provider: types.ProviderCustom, Type: modelcatalog.TestCallType, Name: "Test Model", Description: "Local upstream for full-path ingest testing; echoes requested usage", Endpoint: &devMockEndpoint, MaxContext: 128000, Active: true, Capabilities: []string{"chat"}, Source: modelcatalog.SourceTest},
	}
}
