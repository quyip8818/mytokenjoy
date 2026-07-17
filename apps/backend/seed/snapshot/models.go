package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	devMockEndpoint := "http://127.0.0.1:8765"
	return []types.ModelInfo{
		{ModelID: contract.IDModelLocalTest, CompanyID: contract.TokenJoyCompanyID, Provider: types.ProviderCustom, Type: modelcatalog.DevCallTypeLocalTest, Name: "Local Test Model", Description: "Local upstream for full-path ingest testing; echoes requested usage", Endpoint: &devMockEndpoint, InputPrice: seedPoints(0.15), OutputPrice: seedPoints(0.6), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat"}},
		// DeepSeek
		{ModelID: contract.IDModel1, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v4", Name: "DeepSeek V4", Description: "DeepSeek 旗舰通用模型，性能对标 GPT-4o", InputPrice: seedPoints(0.3), OutputPrice: seedPoints(0.5), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel2, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-r1", Name: "DeepSeek R1", Description: "DeepSeek 推理模型，擅长数学和代码", InputPrice: seedPoints(0.55), OutputPrice: seedPoints(2.19), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		// Qwen (通义千问)
		{ModelID: contract.IDModel3, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-3.5-plus", Name: "Qwen 3.5 Plus", Description: "通义千问 3.5 Plus，高性价比通用模型", InputPrice: seedPoints(0.8), OutputPrice: seedPoints(2.0), MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel4, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-max-2026", Name: "Qwen Max", Description: "通义千问旗舰模型，综合能力最强", InputPrice: seedPoints(2.4), OutputPrice: seedPoints(9.6), MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		// 智谱 GLM
		{ModelID: contract.IDModel5, CompanyID: contract.TokenJoyCompanyID, Provider: "zhipu", Type: "glm-5", Name: "GLM-5", Description: "智谱 GLM-5 旗舰模型，中文理解能力领先", InputPrice: seedPoints(2.0), OutputPrice: seedPoints(8.0), MaxContext: 512000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		// Kimi (月之暗面)
		{ModelID: contract.IDModel6, CompanyID: contract.TokenJoyCompanyID, Provider: "moonshot", Type: "kimi-k3", Name: "Kimi K3", Description: "月之暗面 Kimi K3，超长上下文旗舰模型", InputPrice: seedPoints(3.0), OutputPrice: seedPoints(15.0), MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		// 豆包 (字节跳动)
		{ModelID: contract.IDModel7, CompanyID: contract.TokenJoyCompanyID, Provider: "bytedance", Type: "doubao-pro-256k", Name: "豆包 Pro 256K", Description: "字节跳动豆包大模型，256K 长上下文", InputPrice: seedPoints(0.5), OutputPrice: seedPoints(0.9), MaxContext: 256000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		// MiniMax
		{ModelID: contract.IDModel8, CompanyID: contract.TokenJoyCompanyID, Provider: "minimax", Type: "minimax-m2", Name: "MiniMax M2", Description: "MiniMax M2 通用模型，多模态能力强", InputPrice: seedPoints(0.3), OutputPrice: seedPoints(1.2), MaxContext: 256000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		// 国际模型 (保留少量)
		{ModelID: contract.IDModel9, CompanyID: contract.TokenJoyCompanyID, Provider: "anthropic", Type: "claude-sonnet-5", Name: "Claude Sonnet 5", Description: "Anthropic 最新旗舰模型", InputPrice: seedPoints(3.0), OutputPrice: seedPoints(15.0), MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel10, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o", Name: "GPT-4o", Description: "OpenAI 多模态旗舰模型", InputPrice: seedPoints(2.5), OutputPrice: seedPoints(10.0), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
	}
}
