package snapshot

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
)

func buildModels() []types.ModelInfo {
	return []types.ModelInfo{
		{ModelID: contract.IDModel1, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o", Name: "GPT-4o", Description: "OpenAI flagship multimodal model", InputPrice: seedPoints(2.5), OutputPrice: seedPoints(10), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel2, CompanyID: contract.TokenJoyCompanyID, Provider: "openai", Type: "gpt-4o-mini", Name: "GPT-4o Mini", Description: "Cost-efficient OpenAI model", InputPrice: seedPoints(0.15), OutputPrice: seedPoints(0.6), MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel3, CompanyID: contract.TokenJoyCompanyID, Provider: "anthropic", Type: "claude-opus-4-8", Name: "Claude Opus 4.8", Description: "Anthropic flagship model", InputPrice: seedPoints(15), OutputPrice: seedPoints(75), MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel4, CompanyID: contract.TokenJoyCompanyID, Provider: "anthropic", Type: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", Description: "Balanced Anthropic model", InputPrice: seedPoints(3), OutputPrice: seedPoints(15), MaxContext: 200000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ModelID: contract.IDModel5, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-v3", Name: "DeepSeek V3", Description: "DeepSeek general model", InputPrice: seedPoints(0.27), OutputPrice: seedPoints(1.1), MaxContext: 64000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel6, CompanyID: contract.TokenJoyCompanyID, Provider: "deepseek", Type: "deepseek-r1", Name: "DeepSeek R1", Description: "DeepSeek reasoning model", InputPrice: seedPoints(0.55), OutputPrice: seedPoints(2.19), MaxContext: 64000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel7, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-max", Name: "Qwen Max", Description: "Alibaba Qwen flagship", InputPrice: seedPoints(2.0), OutputPrice: seedPoints(6.0), MaxContext: 32000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ModelID: contract.IDModel8, CompanyID: contract.TokenJoyCompanyID, Provider: "qwen", Type: "qwen-plus", Name: "Qwen Plus", Description: "Alibaba Qwen plus tier", InputPrice: seedPoints(0.8), OutputPrice: seedPoints(2.0), MaxContext: 131072, Enabled: true, Capabilities: []string{"chat", "vision"}},
	}
}
