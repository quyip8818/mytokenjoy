package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildModels() []types.ModelInfo {
	return []types.ModelInfo{
		{ID: "model-1", Provider: "openai", Name: "gpt-4o", DisplayName: "GPT-4o", InputPrice: 2.5, OutputPrice: 10, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-2", Provider: "openai", Name: "gpt-4o-mini", DisplayName: "GPT-4o Mini", InputPrice: 0.15, OutputPrice: 0.6, MaxContext: 128000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-3", Provider: "anthropic", Name: "claude-opus-4-8", DisplayName: "Claude Opus 4.8", InputPrice: 15, OutputPrice: 75, MaxContext: 1000000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-4", Provider: "anthropic", Name: "claude-sonnet-4-6", DisplayName: "Claude Sonnet 4.6", InputPrice: 3, OutputPrice: 15, MaxContext: 200000, Enabled: true, Capabilities: []string{"chat", "vision", "function_calling"}},
		{ID: "model-5", Provider: "deepseek", Name: "deepseek-v3", DisplayName: "DeepSeek V3", InputPrice: 0.27, OutputPrice: 1.1, MaxContext: 64000, Enabled: true, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-6", Provider: "deepseek", Name: "deepseek-r1", DisplayName: "DeepSeek R1", InputPrice: 0.55, OutputPrice: 2.19, MaxContext: 64000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-7", Provider: "qwen", Name: "qwen-max", DisplayName: "Qwen Max", InputPrice: 2.0, OutputPrice: 6.0, MaxContext: 32000, Enabled: false, Capabilities: []string{"chat", "function_calling"}},
		{ID: "model-8", Provider: "qwen", Name: "qwen-plus", DisplayName: "Qwen Plus", InputPrice: 0.8, OutputPrice: 2.0, MaxContext: 131072, Enabled: true, Capabilities: []string{"chat", "vision"}},
	}
}

func buildRoutingRules() []types.RoutingRule {
	gpt4oMini := "gpt-4o-mini"
	deepseek := "deepseek-v3"
	claudeSonnet := "claude-sonnet-4-6"
	gpt4o := "gpt-4o"
	qwenPlus := "qwen-plus"
	return []types.RoutingRule{
		{ID: "rr-1", NodeID: "dept-1", NodeName: "总公司", AllowedModels: []string{"gpt-4o", "gpt-4o-mini", "claude-sonnet-4-6", "deepseek-v3", "qwen-plus"}, DefaultModel: &gpt4oMini, FallbackModel: &deepseek, Inherited: false},
		{ID: "rr-2", NodeID: "dept-2", NodeName: "技术部", AllowedModels: []string{"gpt-4o", "gpt-4o-mini", "claude-sonnet-4-6", "claude-opus-4-8", "deepseek-v3"}, DefaultModel: &gpt4o, FallbackModel: &deepseek, Inherited: false},
		{ID: "rr-3", NodeID: "dept-3", NodeName: "后端组", AllowedModels: []string{"gpt-4o", "claude-sonnet-4-6", "deepseek-v3"}, Inherited: true},
		{ID: "rr-4", NodeID: "dept-6", NodeName: "产品部", AllowedModels: []string{"gpt-4o-mini", "deepseek-v3", "qwen-plus"}, DefaultModel: &gpt4oMini, FallbackModel: &qwenPlus, Inherited: false},
		{ID: "rr-5", NodeID: "dept-4", NodeName: "前端组", AllowedModels: []string{"gpt-4o-mini", "claude-sonnet-4-6", "deepseek-v3"}, DefaultModel: &claudeSonnet, FallbackModel: &gpt4oMini, Inherited: true},
		{ID: "rr-6", NodeID: "dept-5", NodeName: "测试组", AllowedModels: []string{"gpt-4o-mini", "deepseek-v3"}, DefaultModel: &deepseek, FallbackModel: &gpt4oMini, Inherited: true},
		{ID: "rr-7", NodeID: "dept-7", NodeName: "市场部", AllowedModels: []string{"gpt-4o-mini", "qwen-plus", "deepseek-v3"}, DefaultModel: &qwenPlus, FallbackModel: &gpt4oMini, Inherited: false},
		{ID: "rr-8", NodeID: "dept-8", NodeName: "行政部", AllowedModels: []string{"gpt-4o-mini"}, DefaultModel: &gpt4oMini, Inherited: true},
	}
}
