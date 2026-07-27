package newapisync

import (
	"context"
	"strings"
)

// PullFromNewAPI 从 sms-newapi 拉取渠道和模型数据，写入本地 DB
func (s *Service) PullFromNewAPI(ctx context.Context) (*PullResult, error) {
	if s.pullStore == nil {
		return nil, nil
	}

	// 1. 拉取渠道列表
	channels, err := s.admin.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 拉取模型价格（成本价）
	ratios, err := s.admin.ListCurrentRatios(ctx)
	if err != nil {
		return nil, err
	}

	result := &PullResult{}

	// 3. Upsert 渠道
	for i := range channels {
		if err := s.pullStore.UpsertChannel(ctx, &channels[i]); err != nil {
			s.logger.Warn("upsert channel failed", "channel", channels[i].Name, "error", err)
			continue
		}
		result.ChannelsSynced++
	}

	// 4. 从渠道 models 字段收集所有 model_id → channel 映射
	type modelEntry struct {
		modelID   string
		channelID int
	}
	var allModels []modelEntry
	activeIDs := make([]string, 0)

	for _, ch := range channels {
		if ch.Models == "" {
			continue
		}
		for _, modelID := range strings.Split(ch.Models, ",") {
			modelID = strings.TrimSpace(modelID)
			if modelID == "" {
				continue
			}
			allModels = append(allModels, modelEntry{modelID: modelID, channelID: ch.ID})
			activeIDs = append(activeIDs, modelID)
		}
	}

	// 5. Upsert 模型（带成本价）
	for _, entry := range allModels {
		pricing := ratios[entry.modelID]
		input := &SyncedModelInput{
			ModelID:    entry.modelID,
			ModelName:  entry.modelID,
			ChannelID:  entry.channelID,
			CostInput:  pricing.InputPrice,
			CostOutput: pricing.OutputPrice,
		}
		if err := s.pullStore.UpsertSyncedModel(ctx, input); err != nil {
			s.logger.Warn("upsert synced model failed", "modelId", entry.modelID, "error", err)
			continue
		}
		result.ModelsCreated++
	}

	// 6. 清理不再存在的 sync 模型
	removed, err := s.pullStore.DeprecateStaleSyncModels(ctx, activeIDs)
	if err != nil {
		s.logger.Warn("deprecate stale models failed", "error", err)
	} else {
		result.ModelsRemoved = removed
	}

	s.logger.Info("pull from newapi completed",
		"channels", result.ChannelsSynced,
		"models", result.ModelsCreated,
		"removed", result.ModelsRemoved)

	return result, nil
}
