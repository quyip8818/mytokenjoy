package seed

import "github.com/tokenjoy/backend/internal/domain/types"

func buildModelUsage() []types.ModelUsage {
	return []types.ModelUsage{
		{ModelID: "model-1", ModelName: "GPT-4o", Provider: "openai", Requests: 12000, Tokens: 18000000, Cost: 32000, Percentage: 47.4},
		{ModelID: "model-5", ModelName: "DeepSeek V3", Provider: "deepseek", Requests: 8500, Tokens: 15000000, Cost: 12500, Percentage: 18.5},
		{ModelID: "model-4", ModelName: "Claude Sonnet 4.6", Provider: "anthropic", Requests: 4500, Tokens: 7000000, Cost: 14000, Percentage: 20.7},
		{ModelID: "model-2", ModelName: "GPT-4o Mini", Provider: "openai", Requests: 3000, Tokens: 4000000, Cost: 5500, Percentage: 8.1},
		{ModelID: "model-8", ModelName: "Qwen Plus", Provider: "qwen", Requests: 500, Tokens: 1000000, Cost: 3500, Percentage: 5.2},
	}
}

func buildTeamUsage() []types.TeamUsage {
	return []types.TeamUsage{
		{DepartmentID: "dept-3", DepartmentName: "后端组", Quota: 25000, Consumed: 21000, MemberCount: 20, TopModel: "GPT-4o"},
		{DepartmentID: "dept-4", DepartmentName: "前端组", Quota: 15000, Consumed: 11200, MemberCount: 15, TopModel: "Claude Sonnet 4.6"},
		{DepartmentID: "dept-5", DepartmentName: "测试组", Quota: 10000, Consumed: 6000, MemberCount: 10, TopModel: "DeepSeek V3"},
		{DepartmentID: "dept-6", DepartmentName: "产品部", Quota: 20000, Consumed: 14300, MemberCount: 25, TopModel: "GPT-4o Mini"},
		{DepartmentID: "dept-7", DepartmentName: "市场部", Quota: 15000, Consumed: 8500, MemberCount: 30, TopModel: "Qwen Plus"},
		{DepartmentID: "dept-8", DepartmentName: "行政部", Quota: 15000, Consumed: 6500, MemberCount: 28, TopModel: "GPT-4o Mini"},
	}
}
