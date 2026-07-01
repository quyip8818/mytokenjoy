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

func insertRoutingRules(ctx context.Context, exec tableWriter, tid int64, rules []types.RoutingRule) error {
	for _, rule := range rules {
		if _, err := exec.Exec(ctx, `
			INSERT INTO routing_rules (id, company_id, node_id, node_name, default_model, fallback_model, inherited)
			VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (company_id, id) DO NOTHING
		`, rule.ID, tid, rule.NodeID, rule.NodeName, rule.DefaultModel, rule.FallbackModel, rule.Inherited); err != nil {
			return err
		}
		for _, modelName := range rule.AllowedModels {
			if _, err := exec.Exec(ctx, `
				INSERT INTO routing_rule_models (company_id, rule_id, model_name) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, rule.ID, modelName); err != nil {
				return err
			}
		}
	}
	return nil
}
