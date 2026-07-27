package newapisync

import "context"

// Channel 表示 sms-newapi 中的一个渠道
type Channel struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Status   int    `json:"status"`
	Models   string `json:"models"` // 逗号分隔的模型列表
	BaseURL  string `json:"base_url"`
	Priority int    `json:"priority"`
	Weight   int    `json:"weight"`
}

// PullResult 表示一次从 sms-newapi 拉取的结果
type PullResult struct {
	ChannelsSynced int `json:"channelsSynced"`
	ModelsCreated  int `json:"modelsCreated"`
	ModelsUpdated  int `json:"modelsUpdated"`
	ModelsRemoved  int `json:"modelsRemoved"`
}

// SyncedModelInput 表示从 sms-newapi 同步来的模型写入参数
type SyncedModelInput struct {
	ModelID    string
	ModelName  string // 默认用 ModelID
	ChannelID  int    // newapi channel ID
	CostInput  float64
	CostOutput float64
}

// PullStore 定义 PullFromNewAPI 需要的存储接口
type PullStore interface {
	UpsertChannel(ctx context.Context, ch *Channel) error
	UpsertSyncedModel(ctx context.Context, m *SyncedModelInput) error
	DeprecateStaleSyncModels(ctx context.Context, activeModelIDs []string) (int, error)
}
