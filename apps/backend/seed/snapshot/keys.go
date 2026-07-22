package snapshot

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/seed/data"
)

func loadPlatformKeys() []types.PlatformKey {
	type platformKeySeed struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		KeyPrefix      string   `json:"keyPrefix"`
		Scope          string   `json:"scope"`
		MemberID       *string  `json:"memberId"`
		ProjectID      *string  `json:"projectId"`
		Status         string   `json:"status"`
		Budget         float64  `json:"budget"`
		ModelWhitelist []string `json:"modelWhitelist"`
		CreatedAt      string   `json:"createdAt"`
		ExpiresAt      *string  `json:"expiresAt"`
	}
	var raw []platformKeySeed
	if err := json.Unmarshal(data.PlatformKeysJSON, &raw); err != nil {
		panic("seed: load platform keys: " + err.Error())
	}
	var keys []types.PlatformKey
	keys = make([]types.PlatformKey, len(raw))
	for i, item := range raw {
		var memberID *uuid.UUID
		if item.MemberID != nil {
			parsed := uuid.MustParse(*item.MemberID)
			memberID = &parsed
		}
		var projectID *uuid.UUID
		if item.ProjectID != nil {
			parsed := uuid.MustParse(*item.ProjectID)
			projectID = &parsed
		}
		modelWhitelist := make([]uuid.UUID, 0, len(item.ModelWhitelist))
		for _, m := range item.ModelWhitelist {
			modelWhitelist = append(modelWhitelist, uuid.MustParse(m))
		}
		keys[i] = types.PlatformKey{
			ID: uuid.MustParse(item.ID), Name: item.Name, KeyPrefix: item.KeyPrefix, Scope: item.Scope,
			MemberID: memberID, ProjectID: projectID,
			Status: item.Status, Budget: seedQuota(item.Budget),
			ModelWhitelist: modelWhitelist,
			CreatedAt:      item.CreatedAt, ExpiresAt: item.ExpiresAt,
		}
	}
	for i := range keys {
		if consumed, ok := contract.DemoPlatformKeyConsumed[keys[i].ID]; ok {
			keys[i].Consumed = consumed
		}
	}
	return keys
}

func buildProviderKeys() []types.ProviderKey {
	return []types.ProviderKey{
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001001"), Provider: "openai", Name: "OpenAI 主力", KeyPrefix: "sk-proj-abc...", Status: "active", CreatedAt: "2026-01-15", RotateEnabled: true},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001002"), Provider: "anthropic", Name: "Anthropic 生产", KeyPrefix: "sk-ant-xyz...", Status: "active", CreatedAt: "2026-02-01", RotateEnabled: true},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001003"), Provider: "deepseek", Name: "DeepSeek V3", KeyPrefix: "sk-ds-mno...", Status: "active", CreatedAt: "2026-03-10", RotateEnabled: false},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001004"), Provider: "qwen", Name: "通义千问", KeyPrefix: "sk-qw-pqr...", Status: "disabled", CreatedAt: "2026-04-01", RotateEnabled: false},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001005"), Provider: "openai", Name: "OpenAI 备用", KeyPrefix: "sk-proj-def...", Status: "active", CreatedAt: "2026-05-01", RotateEnabled: true},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001006"), Provider: "openai", Name: "OpenAI 历史", KeyPrefix: "sk-proj-old...", Status: "expired", CreatedAt: "2025-12-01", RotateEnabled: false},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001007"), Provider: "anthropic", Name: "Anthropic 测试", KeyPrefix: "sk-ant-err...", Status: "error", CreatedAt: "2026-04-15", RotateEnabled: false},
		{ID: uuid.MustParse("00000000-0000-7000-8000-000000001008"), Provider: "custom", Name: "自建模型网关", KeyPrefix: "sk-cst-ghi...", Status: "active", CreatedAt: "2026-05-20", RotateEnabled: true},
	}
}
