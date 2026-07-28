package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

type pgModelsRepo struct {
	db        dbQuerier
	allowlist *pgModelAllowlistRepo
	catalog   modelCatalog
}

func (r *pgModelsRepo) Allowlist() store.ModelAllowlistRepository {
	return r.allowlist
}

func (r *pgModelsRepo) Models(ctx context.Context) ([]types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	items, err := r.queryModels(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE company_id = $1 OR company_id = $2
		ORDER BY CASE WHEN company_id = $1 THEN 0 ELSE 1 END, model_id
	`, r.catalog.globalCompanyID(), companyID)
	if err != nil {
		return nil, err
	}
	return modelcatalog.DedupeEffective(items), nil
}

func (r *pgModelsRepo) ModelByType(ctx context.Context, modelType string) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	item, err := r.modelByCompanyAndType(ctx, companyID, modelType)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	return r.modelByCompanyAndType(ctx, r.catalog.globalCompanyID(), modelType)
}

func (r *pgModelsRepo) ModelByProviderType(ctx context.Context, provider, modelType string) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	item, err := r.modelByCompanyProviderAndType(ctx, companyID, provider, modelType)
	if err != nil {
		return nil, err
	}
	if item != nil {
		return item, nil
	}
	return r.modelByCompanyProviderAndType(ctx, r.catalog.globalCompanyID(), provider, modelType)
}

func (r *pgModelsRepo) GlobalModelByProviderType(ctx context.Context, provider, modelType string) (*types.ModelInfo, error) {
	return r.modelByCompanyProviderAndType(ctx, r.catalog.globalCompanyID(), provider, modelType)
}

func (r *pgModelsRepo) ModelByID(ctx context.Context, modelID uuid.UUID) (*types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE model_id = $1 AND (company_id = $2 OR company_id = $3)
	`, modelID, r.catalog.globalCompanyID(), companyID)
	return scanModelQueryRow(row)
}

func (r *pgModelsRepo) ModelByIDs(ctx context.Context, modelIDs []int64) ([]types.ModelInfo, error) {
	if len(modelIDs) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	return r.queryModels(ctx, `
		SELECT `+modelSelectColumns+`
		FROM models
		WHERE model_id = ANY($1) AND (company_id = $2 OR company_id = $3)
		ORDER BY model_id
	`, modelIDs, r.catalog.globalCompanyID(), companyID)
}

func (r *pgModelsRepo) InsertModel(ctx context.Context, model types.ModelInfo) (types.ModelInfo, error) {
	companyID := store.CompanyID(ctx)
	capabilities := model.Capabilities
	if capabilities == nil {
		capabilities = []string{}
	}
	var modelID uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO models (
			company_id, provider, type, name, description, endpoint,
			api_key, endpoint_model_name,
			max_context, max_tokens, active, capabilities, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING model_id
	`, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
		model.ApiKey, model.EndpointModelName,
		model.MaxContext, model.MaxTokens, model.Active,
		capabilities).Scan(&modelID)
	if err != nil {
		return types.ModelInfo{}, fmt.Errorf("insert model: %w", err)
	}
	model.ID = modelID
	model.CompanyID = companyID
	return model, nil
}

func (r *pgModelsRepo) UpdateModel(ctx context.Context, model types.ModelInfo) error {
	companyID := store.CompanyID(ctx)
	capabilities := model.Capabilities
	if capabilities == nil {
		capabilities = []string{}
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE models SET
			provider = $3,
			type = $4,
			name = $5,
			description = $6,
			endpoint = $7,
			api_key = $8,
			endpoint_model_name = $9,
			max_context = $10,
			max_tokens = $11,
			active = $12,
			capabilities = $13,
			updated_at = NOW()
		WHERE model_id = $1 AND company_id = $2
	`, model.ID, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
		model.ApiKey, model.EndpointModelName,
		model.MaxContext, model.MaxTokens, model.Active,
		capabilities)
	if err != nil {
		return fmt.Errorf("update model %d: %w", model.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("model %d not found in tenant scope", model.ID)
	}
	return nil
}

func (r *pgModelsRepo) DeleteModel(ctx context.Context, modelID uuid.UUID) error {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, `
		UPDATE org_nodes SET default_model_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND default_model_id = $2
	`, companyID, modelID); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE org_nodes SET fallback_model_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND fallback_model_id = $2
	`, companyID, modelID); err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `
		DELETE FROM models WHERE model_id = $1 AND company_id = $2
	`, modelID, companyID)
	if err != nil {
		return fmt.Errorf("delete model %d: %w", modelID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("model %d not found in tenant scope", modelID)
	}
	return nil
}

var _ store.ModelsRepository = (*pgModelsRepo)(nil)

func (r *pgModelsRepo) SyncFromSMS(ctx context.Context, companyID uuid.UUID, models []types.ModelInfo) error {
	// ponytail: use sms_synced_at as the marker — step 1 sets it to NOW(),
	// step 2 deactivates anything with an older timestamp. No type-list needed.
	// Upgrade path: if provider+type combos cause issues, switch to (provider,type) pair matching.

	// Step 1: Upsert all models from SMS — mark active, update name, bump sms_synced_at.
	for _, m := range models {
		_, err := r.db.Exec(ctx, `
			INSERT INTO models (company_id, provider, type, name, source, active, sms_synced_at, updated_at)
			VALUES ($1, $2, $3, $4, 'sms', TRUE, NOW(), NOW())
			ON CONFLICT (company_id, provider, type) DO UPDATE SET
				name = EXCLUDED.name,
				source = 'sms',
				active = TRUE,
				sms_synced_at = NOW(),
				updated_at = NOW()
		`, companyID, m.Provider, m.Type, m.Name)
		if err != nil {
			return fmt.Errorf("upsert sms model %s: %w", m.Type, err)
		}
	}

	// Step 2: Deactivate stale SMS models — those not touched by step 1.
	_, err := r.db.Exec(ctx, `
		UPDATE models SET active = FALSE, updated_at = NOW()
		WHERE company_id = $1 AND source = 'sms' AND active = TRUE
			AND sms_synced_at < NOW() - INTERVAL '10 seconds'
	`, companyID)
	if err != nil {
		return fmt.Errorf("deactivate stale sms models: %w", err)
	}
	return nil
}
