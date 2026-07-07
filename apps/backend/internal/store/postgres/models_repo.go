package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelsRepo struct {
	db        dbQuerier
	allowlist *pgModelAllowlistRepo
}

func (r *pgModelsRepo) Allowlist() store.ModelAllowlistRepository {
	return r.allowlist
}

func (r *pgModelsRepo) Models(ctx context.Context) ([]types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, provider, name, display_name, model_type, description, visibility, endpoint,
			input_price, output_price, max_context, enabled
		FROM models WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	type modelRow struct {
		item types.ModelInfo
	}
	batch := make([]modelRow, 0)
	for rows.Next() {
		var row modelRow
		if err := rows.Scan(
			&row.item.ID, &row.item.Provider, &row.item.Name, &row.item.DisplayName,
			&row.item.Type, &row.item.Description, &row.item.Visibility, &row.item.Endpoint,
			&row.item.InputPrice, &row.item.OutputPrice, &row.item.MaxContext, &row.item.Enabled,
		); err != nil {
			rows.Close()
			return nil, err
		}
		batch = append(batch, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	items := make([]types.ModelInfo, 0, len(batch))
	for _, row := range batch {
		item := row.item
		capRows, err := r.db.Query(ctx, `
			SELECT capability FROM model_capabilities WHERE company_id = $1 AND model_id = $2 ORDER BY capability
		`, companyID, item.ID)
		if err != nil {
			return nil, err
		}
		for capRows.Next() {
			var cap string
			if err := capRows.Scan(&cap); err != nil {
				capRows.Close()
				return nil, err
			}
			item.Capabilities = append(item.Capabilities, cap)
		}
		capRows.Close()
		items = append(items, item)
	}
	return store.CloneModels(items), nil
}

func (r *pgModelsRepo) ModelByName(ctx context.Context, name string) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT id, provider, name, display_name, model_type, description, visibility, endpoint,
			input_price, output_price, max_context, enabled
		FROM models WHERE company_id = $1 AND name = $2
	`, companyID, name)
	var item types.ModelInfo
	if err := row.Scan(
		&item.ID, &item.Provider, &item.Name, &item.DisplayName,
		&item.Type, &item.Description, &item.Visibility, &item.Endpoint,
		&item.InputPrice, &item.OutputPrice, &item.MaxContext, &item.Enabled,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	capRows, err := r.db.Query(ctx, `
		SELECT capability FROM model_capabilities WHERE company_id = $1 AND model_id = $2 ORDER BY capability
	`, companyID, item.ID)
	if err != nil {
		return nil, err
	}
	for capRows.Next() {
		var cap string
		if err := capRows.Scan(&cap); err != nil {
			capRows.Close()
			return nil, err
		}
		item.Capabilities = append(item.Capabilities, cap)
	}
	capRows.Close()
	cloned := store.CloneModels([]types.ModelInfo{item})[0]
	return &cloned, nil
}

func (r *pgModelsRepo) SetModels(ctx context.Context, models []types.ModelInfo) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneModels(models)
	ids := make([]string, len(cloned))
	for i, model := range cloned {
		ids[i] = model.ID
		if model.Type == "" {
			return fmt.Errorf("model %s: type is required", model.ID)
		}
		if model.Visibility == "" {
			return fmt.Errorf("model %s: visibility is required", model.ID)
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO models (
				id, company_id, provider, name, display_name, model_type, description, visibility, endpoint,
				input_price, output_price, max_context, enabled, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				display_name = EXCLUDED.display_name,
				model_type = EXCLUDED.model_type,
				description = EXCLUDED.description,
				visibility = EXCLUDED.visibility,
				endpoint = EXCLUDED.endpoint,
				input_price = EXCLUDED.input_price,
				output_price = EXCLUDED.output_price,
				max_context = EXCLUDED.max_context,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, model.ID, companyID, model.Provider, model.Name, model.DisplayName,
			model.Type, model.Description, model.Visibility, model.Endpoint,
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
