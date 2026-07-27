package postgres

import (
	"context"

	"sms/backend/internal/domain/newapisync"
)

func (s *Store) UpsertChannel(ctx context.Context, ch *newapisync.Channel) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO channels (newapi_id, name, channel_type, status, models, base_url, priority, weight)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (newapi_id) DO UPDATE SET
			name = EXCLUDED.name,
			channel_type = EXCLUDED.channel_type,
			status = EXCLUDED.status,
			models = EXCLUDED.models,
			base_url = EXCLUDED.base_url,
			priority = EXCLUDED.priority,
			weight = EXCLUDED.weight
	`, ch.ID, ch.Name, ch.Type, ch.Status, ch.Models, ch.BaseURL, ch.Priority, ch.Weight)
	return err
}

func (s *Store) UpsertSyncedModel(ctx context.Context, m *newapisync.SyncedModelInput) error {
	// 查找 channel 的 UUID（从 newapi_id）
	var channelUUID *string
	_ = s.pool.QueryRow(ctx, `SELECT id::text FROM channels WHERE newapi_id = $1`, m.ChannelID).Scan(&channelUUID)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO models (model_id, model_name, channel_id, cost_input, cost_output, status, source)
		VALUES ($1, $2, $3::uuid, $4, $5, 'available', 'sync')
		ON CONFLICT (model_id) WHERE source = 'sync' DO UPDATE SET
			model_name = EXCLUDED.model_name,
			channel_id = EXCLUDED.channel_id,
			cost_input = EXCLUDED.cost_input,
			cost_output = EXCLUDED.cost_output,
			status = 'available'
	`, m.ModelID, m.ModelName, channelUUID, m.CostInput, m.CostOutput)
	return err
}

func (s *Store) DeprecateStaleSyncModels(ctx context.Context, activeModelIDs []string) (int, error) {
	if len(activeModelIDs) == 0 {
		// 如果没有活跃模型，所有 sync 来源的模型都标记为 deprecated
		ct, err := s.pool.Exec(ctx, `
			UPDATE models SET status = 'deprecated' WHERE source = 'sync' AND status = 'available'
		`)
		if err != nil {
			return 0, err
		}
		return int(ct.RowsAffected()), nil
	}

	ct, err := s.pool.Exec(ctx, `
		UPDATE models SET status = 'deprecated'
		WHERE source = 'sync' AND status = 'available' AND model_id != ALL($1)
	`, activeModelIDs)
	if err != nil {
		return 0, err
	}
	return int(ct.RowsAffected()), nil
}
