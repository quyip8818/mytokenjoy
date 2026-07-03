package seed

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func insertModels(ctx context.Context, exec tableWriter, tid int64, models []types.ModelInfo) error {
	for _, model := range models {
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (
				id, company_id, provider, name, display_name, input_price, output_price, max_context, enabled
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (company_id, id) DO NOTHING
		`, model.ID, tid, model.Provider, model.Name, model.DisplayName,
			model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled); err != nil {
			return err
		}
		for _, cap := range model.Capabilities {
			if _, err := exec.Exec(ctx, `
				INSERT INTO model_capabilities (company_id, model_id, capability) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, model.ID, cap); err != nil {
				return err
			}
		}
	}
	return nil
}
