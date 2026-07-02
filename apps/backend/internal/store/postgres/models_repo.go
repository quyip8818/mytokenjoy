package postgres

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelsRepo struct {
	db dbQuerier
}

func (r *pgModelsRepo) Models(ctx context.Context) ([]types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, provider, name, display_name, input_price, output_price, max_context, enabled
		FROM models WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.ModelInfo, 0)
	for rows.Next() {
		var item types.ModelInfo
		if err := rows.Scan(
			&item.ID, &item.Provider, &item.Name, &item.DisplayName,
			&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.Enabled,
		); err != nil {
			return nil, err
		}
		capRows, err := r.db.Query(ctx, `
			SELECT capability FROM model_capabilities WHERE company_id = $1 AND model_id = $2 ORDER BY capability
		`, companyID, item.ID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneModels(items), nil
}

func (r *pgModelsRepo) SetModels(ctx context.Context, models []types.ModelInfo) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneModels(models)
	ids := make([]string, len(cloned))
	for i, model := range cloned {
		ids[i] = model.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO models (
				id, company_id, provider, name, display_name, input_price, output_price, max_context, enabled, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				display_name = EXCLUDED.display_name,
				input_price = EXCLUDED.input_price,
				output_price = EXCLUDED.output_price,
				max_context = EXCLUDED.max_context,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, model.ID, companyID, model.Provider, model.Name, model.DisplayName,
			model.InputPrice, model.OutputPrice, model.MaxContext, model.Enabled); err != nil {
			return fmt.Errorf("upsert model %s: %w", model.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM model_capabilities WHERE company_id = $1 AND model_id = $2`, companyID, model.ID); err != nil {
			return err
		}
		for _, capability := range model.Capabilities {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO model_capabilities (company_id, model_id, capability) VALUES ($1, $2, $3)
			`, companyID, model.ID, capability); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM model_capabilities WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM models WHERE company_id = $1`, companyID)
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM model_capabilities WHERE company_id = $1 AND NOT (model_id = ANY($2))`, companyID, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM models WHERE company_id = $1 AND NOT (id = ANY($2))`, companyID, ids); err != nil {
		return err
	}
	return nil
}

func (r *pgModelsRepo) RoutingRules(ctx context.Context) ([]types.RoutingRule, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, node_id, node_name, default_model, fallback_model, inherited
		FROM routing_rules WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.RoutingRule, 0)
	for rows.Next() {
		var rule types.RoutingRule
		if err := rows.Scan(
			&rule.ID, &rule.NodeID, &rule.NodeName,
			&rule.DefaultModel, &rule.FallbackModel, &rule.Inherited,
		); err != nil {
			return nil, err
		}
		modelRows, err := r.db.Query(ctx, `
			SELECT model_name FROM routing_rule_models WHERE company_id = $1 AND rule_id = $2 ORDER BY model_name
		`, companyID, rule.ID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneRoutingRules(items), nil
}

func (r *pgModelsRepo) SetRoutingRules(ctx context.Context, rules []types.RoutingRule) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneRoutingRules(rules)
	ids := make([]string, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO routing_rules (
				id, company_id, node_id, node_name, default_model, fallback_model, inherited, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				node_name = EXCLUDED.node_name,
				default_model = EXCLUDED.default_model,
				fallback_model = EXCLUDED.fallback_model,
				inherited = EXCLUDED.inherited,
				updated_at = NOW()
		`, rule.ID, companyID, rule.NodeID, rule.NodeName, rule.DefaultModel, rule.FallbackModel, rule.Inherited); err != nil {
			return fmt.Errorf("upsert routing rule %s: %w", rule.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM routing_rule_models WHERE company_id = $1 AND rule_id = $2`, companyID, rule.ID); err != nil {
			return err
		}
		for _, modelName := range rule.AllowedModels {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO routing_rule_models (company_id, rule_id, model_name) VALUES ($1, $2, $3)
			`, companyID, rule.ID, modelName); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM routing_rule_models WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM routing_rules WHERE company_id = $1`, companyID)
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM routing_rule_models WHERE company_id = $1 AND NOT (rule_id = ANY($2))`, companyID, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM routing_rules WHERE company_id = $1 AND NOT (id = ANY($2))`, companyID, ids); err != nil {
		return err
	}
	return nil
}
