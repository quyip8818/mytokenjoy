package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	devMockEndpoint := "http://127.0.0.1:8765"
	return []types.ModelInfo{
		{ID: contract.IDModelTest, CompanyID: contract.TokenJoyCompanyID, Provider: types.ProviderCustom, Type: modelcatalog.TestCallType, Name: "Test Model", Description: "Local upstream for full-path ingest testing; echoes requested usage", Endpoint: &devMockEndpoint, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat"}},
		// DeepSeek
		{ID: contract.IDModel1, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", Description: "DeepSeek 旗舰推理模型，性能对标 GPT-5", MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "function"}},
		{ID: contract.IDModel11, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", Description: "DeepSeek 高速经济模型，适合日常对话和轻量任务", MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "function"}},
	}
}
