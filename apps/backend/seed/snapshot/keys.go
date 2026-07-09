package snapshot

import (
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/data"
)

func loadPlatformKeys() []types.PlatformKey {
	type platformKeySeed struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		KeyPrefix      string  `json:"keyPrefix"`
		MemberID       *string `json:"memberId"`
		BudgetGroupID  *string `json:"budgetGroupId"`
		Status         string  `json:"status"`
		Quota          float64 `json:"quota"`
		ModelWhitelist []int64 `json:"modelWhitelist"`
		CreatedAt      string  `json:"createdAt"`
		ExpiresAt      *string `json:"expiresAt"`
	}
	var raw []platformKeySeed
	if err := json.Unmarshal(data.PlatformKeysJSON, &raw); err != nil {
		panic("seed: load platform keys: " + err.Error())
	}
	var keys []types.PlatformKey
	keys = make([]types.PlatformKey, len(raw))
	for i, item := range raw {
		keys[i] = types.PlatformKey{
			ID: item.ID, Name: item.Name, KeyPrefix: item.KeyPrefix,
			MemberID: item.MemberID, BudgetGroupID: item.BudgetGroupID,
			Status: item.Status, Quota: seedPoints(item.Quota),
			ModelWhitelist: append([]int64{}, item.ModelWhitelist...),
			CreatedAt:      item.CreatedAt, ExpiresAt: item.ExpiresAt,
		}
	}
	for i := range keys {
		if used, ok := contract.DemoPlatformKeyUsed[keys[i].ID]; ok {
			keys[i].Used = used
		}
	}
	return keys
}

func buildProviderKeys(refDate string) []types.ProviderKey {
	b4250 := 4250.0
	b2100 := 2100.0
	b800 := 800.0
	b0 := 0.0
	b1500 := 1500.0
	b3200 := 3200.0
	lastUsed1 := refDate + " 10:32"
	lastUsed2 := refDate + " 09:45"
	lastUsed3 := "2026-06-18 16:20"
	lastUsed4 := "2026-05-20 12:00"
	lastUsed5 := "2026-06-17 08:00"
	lastUsed6 := "2026-03-01 10:00"
	lastUsed7 := "2026-06-10 14:00"
	lastUsed8 := "2026-06-18 11:30"
	return []types.ProviderKey{
		{ID: "pk-1", Provider: "openai", Name: "OpenAI 主力", KeyPrefix: "sk-proj-abc...", Status: "active", Balance: &b4250, LastUsed: &lastUsed1, CreatedAt: "2026-01-15", RotateEnabled: true},
		{ID: "pk-2", Provider: "anthropic", Name: "Anthropic 生产", KeyPrefix: "sk-ant-xyz...", Status: "active", Balance: &b2100, LastUsed: &lastUsed2, CreatedAt: "2026-02-01", RotateEnabled: true},
		{ID: "pk-3", Provider: "deepseek", Name: "DeepSeek V3", KeyPrefix: "sk-ds-mno...", Status: "active", Balance: &b800, LastUsed: &lastUsed3, CreatedAt: "2026-03-10", RotateEnabled: false},
		{ID: "pk-4", Provider: "qwen", Name: "通义千问", KeyPrefix: "sk-qw-pqr...", Status: "disabled", Balance: nil, LastUsed: &lastUsed4, CreatedAt: "2026-04-01", RotateEnabled: false},
		{ID: "pk-5", Provider: "openai", Name: "OpenAI 备用", KeyPrefix: "sk-proj-def...", Status: "active", Balance: &b1500, LastUsed: &lastUsed5, CreatedAt: "2026-05-01", RotateEnabled: true},
		{ID: "pk-6", Provider: "openai", Name: "OpenAI 历史", KeyPrefix: "sk-proj-old...", Status: "expired", Balance: &b0, LastUsed: &lastUsed6, CreatedAt: "2025-12-01", RotateEnabled: false},
		{ID: "pk-7", Provider: "anthropic", Name: "Anthropic 测试", KeyPrefix: "sk-ant-err...", Status: "error", Balance: nil, LastUsed: &lastUsed7, CreatedAt: "2026-04-15", RotateEnabled: false},
		{ID: "pk-8", Provider: "custom", Name: "自建模型网关", KeyPrefix: "sk-cst-ghi...", Status: "active", Balance: &b3200, LastUsed: &lastUsed8, CreatedAt: "2026-05-20", RotateEnabled: true},
	}
}

func buildApprovals() []types.KeyApproval {
	return []types.KeyApproval{
		{ID: contract.IDApproval1, Type: "key", Applicant: "钱七", ApplicantID: "m-5", Department: "前端组", Reason: "需要接入 GPT-4o 进行代码辅助开发", RequestedQuota: seedPoints(5000), RequestedModels: []int64{contract.IDModel1, contract.IDModel4}, Status: "pending", CreatedAt: "2026-06-18 14:30"},
		{ID: "apv-2", Type: "key", Applicant: "王五", ApplicantID: "m-3", Department: "后端组", Reason: "新项目需要多模型测试", RequestedQuota: seedPoints(8000), RequestedModels: []int64{contract.IDModel1, contract.IDModel5, contract.IDModel4}, Status: "pending", CreatedAt: "2026-06-17 09:15"},
		{ID: "apv-3", Type: "quota", Applicant: "张三", ApplicantID: "m-1", Department: "后端组", Reason: "额度即将用完，申请追加", RequestedQuota: seedPoints(3000), RequestedModels: []int64{contract.IDModel1}, Status: "approved", Approver: strPtr("李四"), CreatedAt: "2026-06-15 11:00", ResolvedAt: strPtr("2026-06-15 14:20")},
	}
}

func strPtr(v string) *string { return &v }
