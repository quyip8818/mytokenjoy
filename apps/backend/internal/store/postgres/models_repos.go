package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelsRepo struct {
	ctx context.Context
	db  dbQuerier
}

func (r *pgModelsRepo) Models() []types.ModelInfo {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, provider, name, display_name, input_price, output_price, max_context, enabled
		FROM models ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.ModelInfo, 0)
	for rows.Next() {
		var item types.ModelInfo
		if err := rows.Scan(
			&item.ID, &item.Provider, &item.Name, &item.DisplayName,
			&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.Enabled,
		); err != nil {
			return nil
		}
		capRows, err := r.db.Query(r.ctx, `
			SELECT capability FROM model_capabilities WHERE model_id = $1 ORDER BY capability
		`, item.ID)
		if err == nil {
			for capRows.Next() {
				var cap string
				if err := capRows.Scan(&cap); err == nil {
					item.Capabilities = append(item.Capabilities, cap)
				}
			}
			capRows.Close()
		}
		items = append(items, item)
	}
	return store.CloneModels(items)
}

func (r *pgModelsRepo) SetModels(models []types.ModelInfo) error {
	cloned := store.CloneModels(models)
	ids := make([]string, len(cloned))
	for i, model := range cloned {
		ids[i] = model.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO models (
				id, provider, name, display_name, input_price, output_price, max_context, enabled, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			ON CONFLICT (id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				display_name = EXCLUDED.display_name,
				input_price = EXCLUDED.input_price,
				output_price = EXCLUDED.output_price,
				max_context = EXCLUDED.max_context,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, model.ID, model.Provider, model.Name, model.DisplayName,
			model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled); err != nil {
			return fmt.Errorf("upsert model %s: %w", model.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM model_capabilities WHERE model_id = $1`, model.ID); err != nil {
			return err
		}
		for _, capability := range model.Capabilities {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO model_capabilities (model_id, capability) VALUES ($1, $2)
			`, model.ID, capability); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM model_capabilities`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM models`)
		return err
	}
	if _, err := r.db.Exec(r.ctx, `DELETE FROM model_capabilities WHERE NOT (model_id = ANY($1))`, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(r.ctx, `DELETE FROM models WHERE NOT (id = ANY($1))`, ids); err != nil {
		return err
	}
	return nil
}

func (r *pgModelsRepo) RoutingRules() []types.RoutingRule {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, node_id, node_name, default_model, fallback_model, inherited
		FROM routing_rules ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.RoutingRule, 0)
	for rows.Next() {
		var rule types.RoutingRule
		if err := rows.Scan(
			&rule.ID, &rule.NodeID, &rule.NodeName,
			&rule.DefaultModel, &rule.FallbackModel, &rule.Inherited,
		); err != nil {
			return nil
		}
		modelRows, err := r.db.Query(r.ctx, `
			SELECT model_name FROM routing_rule_models WHERE rule_id = $1 ORDER BY model_name
		`, rule.ID)
		if err == nil {
			for modelRows.Next() {
				var modelName string
				if err := modelRows.Scan(&modelName); err == nil {
					rule.AllowedModels = append(rule.AllowedModels, modelName)
				}
			}
			modelRows.Close()
		}
		items = append(items, rule)
	}
	return store.CloneRoutingRules(items)
}

func (r *pgModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	cloned := store.CloneRoutingRules(rules)
	ids := make([]string, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO routing_rules (
				id, node_id, node_name, default_model, fallback_model, inherited, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				node_name = EXCLUDED.node_name,
				default_model = EXCLUDED.default_model,
				fallback_model = EXCLUDED.fallback_model,
				inherited = EXCLUDED.inherited,
				updated_at = NOW()
		`, rule.ID, rule.NodeID, rule.NodeName, rule.DefaultModel, rule.FallbackModel, rule.Inherited); err != nil {
			return fmt.Errorf("upsert routing rule %s: %w", rule.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM routing_rule_models WHERE rule_id = $1`, rule.ID); err != nil {
			return err
		}
		for _, modelName := range rule.AllowedModels {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO routing_rule_models (rule_id, model_name) VALUES ($1, $2)
			`, rule.ID, modelName); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM routing_rule_models`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM routing_rules`)
		return err
	}
	if _, err := r.db.Exec(r.ctx, `DELETE FROM routing_rule_models WHERE NOT (rule_id = ANY($1))`, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(r.ctx, `DELETE FROM routing_rules WHERE NOT (id = ANY($1))`, ids); err != nil {
		return err
	}
	return nil
}
