package seed

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func insertModels(ctx context.Context, exec tableWriter, tid int64, models []types.ModelInfo) error {
	for _, model := range models {
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (
				id, company_id, provider, name, display_name, model_type, description, visibility, endpoint,
				input_price, output_price, max_context, enabled
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO NOTHING
		`, model.ID, tid, model.Provider, model.Name, model.DisplayName,
			defaultModelType(model.Type), model.Description, defaultVisibility(model.Visibility), model.Endpoint,
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

func defaultModelType(modelType string) string {
	if modelType == "" {
		return "builtin"
	}
	return modelType
}

func defaultVisibility(visibility string) string {
	if visibility == "" {
		return "all"
	}
	return visibility
}
