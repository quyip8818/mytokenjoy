package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	return []types.ModelInfo{
		{ModelID: contract.IDModel1, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o", Name: "GPT-4o", Description: "OpenAI flagship multimodal model", InputPrice: 2.5, OutputPrice: 10, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel2, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o-mini", Name: "GPT-4o Mini", Description: "Cost-efficient OpenAI model", InputPrice: 0.15, OutputPrice: 0.6, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel3, CompanyID: contract.TokenJoyCompanyID, Provider: "anthropic", Type: "claude-opus-4-8", Name: "Claude Opus 4.8", Description: "Anthropic flagship model", InputPrice: 15, OutputPrice: 75, MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel4, CompanyID: contract.TokenJoyCompanyID, Provider: "anthropic", Type: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", Description: "Balanced Anthropic model", InputPrice: 3, OutputPrice: 15, MaxContext: 200000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel5, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v3", Name: "DeepSeek V3", Description: "DeepSeek general model", InputPrice: 0.27, OutputPrice: 1.1, MaxContext: 64000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel6, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-r1", Name: "DeepSeek R1", Description: "DeepSeek reasoning model", InputPrice: 0.55, OutputPrice: 2.19, MaxContext: 64000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel7, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-max", Name: "Qwen Max", Description: "Alibaba Qwen flagship", InputPrice: 2.0, OutputPrice: 6.0, MaxContext: 32000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel8, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-plus", Name: "Qwen Plus", Description: "Alibaba Qwen plus tier", InputPrice: 0.8, OutputPrice: 2.0, MaxContext: 131072, Enabled: true, Capabilities: []string{"chat", "vision"}},
	}
}
