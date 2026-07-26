package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"sms/backend/internal/domain/sync"
)

type pgSyncStore struct {
	db *pgxpool.Pool
}

func NewSyncStore(db *pgxpool.Pool) *pgSyncStore {
	return &pgSyncStore{db: db}
}

func (s *pgSyncStore) ListModelsForSync(ctx context.Context) ([]sync.CatalogModel, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			COALESCE(m.model_id, m.model_name) AS model_id,
			m.model_name AS display_name,
			COALESCE(sup.name, 'unknown') AS provider,
			COALESCE(m.model_type, 'chat') AS call_type,
			COALESCE(m.input_price, 0) AS input_price,
			COALESCE(m.output_price, 0) AS output_price
		FROM models m
		LEFT JOIN suppliers sup ON sup.id = m.supplier_id
		WHERE m.status = 'available'
		  AND m.model_id IS NOT NULL
		ORDER BY m.model_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []sync.CatalogModel
	for rows.Next() {
		var m sync.CatalogModel
		if err := rows.Scan(&m.ModelID, &m.DisplayName, &m.Provider, &m.CallType, &m.InputPrice, &m.OutputPrice); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

// ListChannelsForSync returns channels configured via the NewAPI sync log table.
// ponytail: For now returns empty — channels come from NewAPI directly.
// In future, SMS may track its own channel configs.
func (s *pgSyncStore) ListChannelsForSync(_ context.Context) ([]sync.CatalogChannel, error) {
	// ponytail: channels are managed in SMS-NewAPI, not in SMS DB.
	// The catalog exposes models (with pricing) from SMS; channels are
	// configured separately in the SMS-NewAPI instance and not part of this sync.
	return []sync.CatalogChannel{}, nil
}
