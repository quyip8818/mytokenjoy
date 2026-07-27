package newapisync

import (
	"context"
	"log/slog"

	"sms/backend/internal/domain/types"
)

// ModelLister 是 newapisync 需要的 model domain 接口子集
type ModelLister interface {
	ListWithModelID(ctx context.Context) ([]types.AiModel, error)
}

// Service 负责 SMS 与 NewAPI 之间的定价同步
type Service struct {
	admin     AdminPort
	models    ModelLister
	pullStore PullStore
	logger    *slog.Logger
}

func NewService(admin AdminPort, models ModelLister, logger *slog.Logger) *Service {
	return &Service{admin: admin, models: models, logger: logger}
}

func NewServiceWithPullStore(admin AdminPort, models ModelLister, pullStore PullStore, logger *slog.Logger) *Service {
	return &Service{admin: admin, models: models, pullStore: pullStore, logger: logger}
}

// SyncStatus 表示本地 vs NewAPI 的对比结果
type SyncStatus struct {
	ModelID      string   `json:"modelId"`
	ModelName    string   `json:"modelName"`
	LocalInput   *float64 `json:"localInput"`
	LocalOutput  *float64 `json:"localOutput"`
	RemoteInput  *float64 `json:"remoteInput,omitempty"`
	RemoteOutput *float64 `json:"remoteOutput,omitempty"`
	Status       string   `json:"status"` // "synced" | "diverged" | "missing" | "skipped"
}

// GetStatus 对比本地模型价格与 NewAPI 当前 ratio，返回差异列表
func (s *Service) GetStatus(ctx context.Context) ([]SyncStatus, error) {
	models, err := s.models.ListWithModelID(ctx)
	if err != nil {
		return nil, err
	}

	remoteMap, err := s.admin.ListCurrentRatios(ctx)
	if err != nil {
		return nil, err
	}

	var results []SyncStatus
	for _, m := range models {
		modelID := ""
		if m.ModelID != nil {
			modelID = *m.ModelID
		}

		st := SyncStatus{
			ModelID:     modelID,
			ModelName:   m.ModelName,
			LocalInput:  m.InputPrice,
			LocalOutput: m.OutputPrice,
		}

		remote, exists := remoteMap[modelID]
		if !exists {
			st.Status = "missing"
		} else {
			ri := remote.InputPrice
			ro := remote.OutputPrice
			st.RemoteInput = &ri
			st.RemoteOutput = &ro
			if priceMatch(m.InputPrice, ri) && priceMatch(m.OutputPrice, ro) {
				st.Status = "synced"
			} else {
				st.Status = "diverged"
			}
		}
		results = append(results, st)
	}
	return results, nil
}

// SyncAll 全量同步：将所有有 model_id + 有价格的模型推送到 NewAPI
func (s *Service) SyncAll(ctx context.Context) (int, error) {
	models, err := s.models.ListWithModelID(ctx)
	if err != nil {
		return 0, err
	}

	var entries []PricingEntry
	for _, m := range models {
		modelID := ""
		if m.ModelID != nil {
			modelID = *m.ModelID
		}
		if modelID == "" {
			continue
		}
		inp := ptrFloat(m.InputPrice)
		out := ptrFloat(m.OutputPrice)
		if inp <= 0 {
			continue
		}
		entries = append(entries, PricingEntry{
			ModelID:     modelID,
			InputPrice:  inp,
			OutputPrice: out,
		})
	}

	if len(entries) == 0 {
		return 0, nil
	}

	if err := s.admin.SyncPricing(ctx, entries); err != nil {
		return 0, err
	}

	s.logger.Info("newapi sync completed", "count", len(entries))
	return len(entries), nil
}

// UpsertOne 同步单个模型（用于创建/更新时的 fire-and-forget）
func (s *Service) UpsertOne(ctx context.Context, modelID string, inputPrice, outputPrice float64) {
	if modelID == "" || inputPrice <= 0 {
		return
	}
	if err := s.admin.UpsertModelRatio(ctx, modelID, inputPrice, outputPrice); err != nil {
		s.logger.Warn("newapi upsert failed", "modelId", modelID, "error", err)
	}
}

// ListModels 返回 NewAPI 上所有模型
func (s *Service) ListModels(ctx context.Context) ([]NewAPIModel, error) {
	return s.admin.ListModels(ctx)
}

func ptrFloat(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func priceMatch(local *float64, remote float64) bool {
	if local == nil {
		return remote == 0
	}
	// 容忍浮点精度：差值 < 0.001
	diff := *local - remote
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.001
}
