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

// ListChannelsForSync returns channels for the catalog.
// ponytail: channels are managed in SMS-NewAPI, not in SMS DB for now.
func (s *pgSyncStore) ListChannelsForSync(_ context.Context) ([]sync.CatalogChannel, error) {
	return []sync.CatalogChannel{}, nil
}

func (s *pgSyncStore) GetPartitionVersions(ctx context.Context) (sync.PartitionVersions, error) {
	var v sync.PartitionVersions
	rows, err := s.db.Query(ctx, `SELECT partition, version FROM sync_versions`)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	for rows.Next() {
		var partition string
		var version int
		if err := rows.Scan(&partition, &version); err != nil {
			return v, err
		}
		switch partition {
		case "channels":
			v.Channels = version
		case "models":
			v.Models = version
		case "currencies":
			v.Currencies = version
		}
	}
	return v, rows.Err()
}

func (s *pgSyncStore) GetPartitionVersion(ctx context.Context, partition string) (int, error) {
	var version int
	err := s.db.QueryRow(ctx, `SELECT version FROM sync_versions WHERE partition = $1`, partition).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
