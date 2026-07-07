package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildModels() []types.ModelInfo {
	return []types.ModelInfo{
		{ID: "model-1", Provider: "openai", Name: "gpt-4o", DisplayName: "GPT-4o", Type: "builtin", Description: "OpenAI flagship multimodal model", Visibility: "all", InputPrice: 2.5, OutputPrice: 10, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-2", Provider: "openai", Name: "gpt-4o-mini", DisplayName: "GPT-4o Mini", Type: "builtin", Description: "Cost-efficient OpenAI model", Visibility: "all", InputPrice: 0.15, OutputPrice: 0.6, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-3", Provider: "anthropic", Name: "claude-opus-4-8", DisplayName: "Claude Opus 4.8", Type: "builtin", Description: "Anthropic flagship model", Visibility: "all", InputPrice: 15, OutputPrice: 75, MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-4", Provider: "anthropic", Name: "claude-sonnet-4-6", DisplayName: "Claude Sonnet 4.6", Type: "builtin", Description: "Balanced Anthropic model", Visibility: "all", InputPrice: 3, OutputPrice: 15, MaxContext: 200000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-5", Provider: "deepseek", Name: "deepseek-v3", DisplayName: "DeepSeek V3", Type: "builtin", Description: "DeepSeek general model", Visibility: "all", InputPrice: 0.27, OutputPrice: 1.1, MaxContext: 64000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-6", Provider: "deepseek", Name: "deepseek-r1", DisplayName: "DeepSeek R1", Type: "builtin", Description: "DeepSeek reasoning model", Visibility: "all", InputPrice: 0.55, OutputPrice: 2.19, MaxContext: 64000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-7", Provider: "qwen", Name: "qwen-max", DisplayName: "Qwen Max", Type: "builtin", Description: "Alibaba Qwen flagship", Visibility: "all", InputPrice: 2.0, OutputPrice: 6.0, MaxContext: 32000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-8", Provider: "qwen", Name: "qwen-plus", DisplayName: "Qwen Plus", Type: "builtin", Description: "Alibaba Qwen plus tier", Visibility: "all", InputPrice: 0.8, OutputPrice: 2.0, MaxContext: 131072, Enabled: true, Capabilities: []string{"chat", "vision"}},
	}
}
