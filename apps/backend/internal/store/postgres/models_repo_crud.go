package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
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
			max_context, max_tokens, deprecated, capabilities,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING model_id
	`, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
		model.ApiKey, model.EndpointModelName,
		model.MaxContext, model.MaxTokens, model.Deprecated,
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
			deprecated = $12,
			capabilities = $13,
			updated_at = NOW()
		WHERE model_id = $1 AND company_id = $2
	`, model.ID, companyID, model.Provider, model.Type, model.Name,
		model.Description, model.Endpoint,
		model.ApiKey, model.EndpointModelName,
		model.MaxContext, model.MaxTokens, model.Deprecated,
		capabilities)
	if err != nil {
		return fmt.Errorf("update model %d: %w", model.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("model %d not found in tenant scope", model.ID)
	}
	return nil
}

var _ store.ModelsRepository = (*pgModelsRepo)(nil)

func (r *pgModelsRepo) SyncFromPlatform(ctx context.Context, companyID uuid.UUID, models []types.ModelInfo) (store.SyncResult, error) {
	var result store.SyncResult

	// Step 0: Get a single batch timestamp from the DB.
	var batchTS time.Time
	if err := r.db.QueryRow(ctx, `SELECT NOW()`).Scan(&batchTS); err != nil {
		return result, fmt.Errorf("sync platform batch ts: %w", err)
	}

	// Step 1: Upsert all models — use xmax to distinguish insert vs update.
	for _, m := range models {
		caps := m.Capabilities
		if caps == nil {
			caps = []string{}
		}
		var xmax uint32
		err := r.db.QueryRow(ctx, `
			INSERT INTO models (company_id, provider, type, name, source, deprecated, capabilities, max_context, catalog_synced_at, updated_at)
			VALUES ($1, $2, $3, $4, 'platform', FALSE, $5, $6, $7, $7)
			ON CONFLICT (company_id, provider, type) DO UPDATE SET
				name = EXCLUDED.name,
				source = 'platform',
				deprecated = FALSE,
				capabilities = EXCLUDED.capabilities,
				max_context = EXCLUDED.max_context,
				catalog_synced_at = $7,
				updated_at = $7
			RETURNING xmax
		`, companyID, m.Provider, m.Type, m.Name, caps, m.MaxContext, batchTS).Scan(&xmax)
		if err != nil {
			return result, fmt.Errorf("upsert platform model %s: %w", m.Type, err)
		}
		if xmax == 0 {
			result.Added++
		} else {
			result.Updated++
		}
	}

	// Step 2: Hard-delete stale platform models not in this batch.
	// First clean FK references, then delete.
	_, _ = r.db.Exec(ctx, `
		DELETE FROM model_allowlist
		WHERE model_id IN (
			SELECT model_id FROM models
			WHERE company_id = $1 AND source = 'platform'
				AND (catalog_synced_at IS NULL OR catalog_synced_at < $2)
		)
	`, companyID, batchTS)

	_, _ = r.db.Exec(ctx, `
		UPDATE org_nodes SET default_model_id = NULL
		WHERE default_model_id IN (
			SELECT model_id FROM models
			WHERE company_id = $1 AND source = 'platform'
				AND (catalog_synced_at IS NULL OR catalog_synced_at < $2)
		)
	`, companyID, batchTS)

	_, _ = r.db.Exec(ctx, `
		UPDATE org_nodes SET fallback_model_id = NULL
		WHERE fallback_model_id IN (
			SELECT model_id FROM models
			WHERE company_id = $1 AND source = 'platform'
				AND (catalog_synced_at IS NULL OR catalog_synced_at < $2)
		)
	`, companyID, batchTS)

	tag, err := r.db.Exec(ctx, `
		DELETE FROM models
		WHERE company_id = $1 AND source = 'platform'
			AND (catalog_synced_at IS NULL OR catalog_synced_at < $2)
	`, companyID, batchTS)
	if err != nil {
		return result, fmt.Errorf("delete stale platform models: %w", err)
	}
	result.Removed = int(tag.RowsAffected())

	return result, nil
}
