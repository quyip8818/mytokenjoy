package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgKeysRepo struct {
	db            dbQuerier
	allowlist     *pgModelAllowlistRepo
	credentialKey []byte
}

const platformKeySelect = `
	SELECT id, name, key_prefix, scope, member_id,
		project_id, status, budget, created_at, expires_at
	FROM platform_keys
`

const platformKeyListSelect = `
	SELECT pk.id, pk.name, pk.key_prefix, pk.scope, pk.member_id,
		pk.project_id, pk.status, pk.budget, pk.created_at, pk.expires_at,
		COALESCE(array_agg(ma.model_id ORDER BY ma.model_id) FILTER (WHERE ma.model_id IS NOT NULL), '{}') AS model_ids
	FROM platform_keys pk
	LEFT JOIN model_allowlist ma
		ON ma.company_id = pk.company_id
		AND ma.owner_type = 'platform_key'
		AND ma.owner_id = pk.id
`

func (r *pgKeysRepo) resolvePlatformKeyHash(ctx context.Context, companyID uuid.UUID, key types.PlatformKey) (string, error) {
	if key.FullKey != nil && *key.FullKey != "" {
		return store.HashPlatformKey(*key.FullKey), nil
	}
	var existing string
	err := r.db.QueryRow(ctx, `
		SELECT key_hash FROM platform_keys WHERE company_id = $1 AND id = $2
	`, companyID, key.ID).Scan(&existing)
	if err == pgx.ErrNoRows {
		return store.HashPlatformKey("pending:" + key.ID.String()), nil
	}
	if err != nil {
		return "", err
	}
	return existing, nil
}

func scanPlatformKeyWithModels(rows pgx.Rows) (types.PlatformKey, error) {
	var item types.PlatformKey
	var createdAt time.Time
	var expiresAt *time.Time
	var modelIDs []uuid.UUID
	if err := rows.Scan(
		&item.ID, &item.Name, &item.KeyPrefix, &item.Scope, &item.MemberID,
		&item.ProjectID, &item.Status,
		&item.Budget, &createdAt, &expiresAt,
		&modelIDs,
	); err != nil {
		return types.PlatformKey{}, err
	}
	item.Consumed = 0
	item.CreatedAt = formatDateOnly(createdAt)
	if expiresAt != nil {
		s := formatDateOnly(*expiresAt)
		item.ExpiresAt = &s
	}
	item.ModelWhitelist = modelIDs
	return item, nil
}

func scanPlatformKey(rows pgx.Rows) (types.PlatformKey, error) {
	var item types.PlatformKey
	var createdAt time.Time
	var expiresAt *time.Time
	if err := rows.Scan(
		&item.ID, &item.Name, &item.KeyPrefix, &item.Scope, &item.MemberID,
		&item.ProjectID, &item.Status,
		&item.Budget, &createdAt, &expiresAt,
	); err != nil {
		return types.PlatformKey{}, err
	}
	item.Consumed = 0
	item.CreatedAt = formatDateOnly(createdAt)
	if expiresAt != nil {
		s := formatDateOnly(*expiresAt)
		item.ExpiresAt = &s
	}
	return item, nil
}

func scanPlatformKeyRow(row pgx.Row) (types.PlatformKey, error) {
	var item types.PlatformKey
	var createdAt time.Time
	var expiresAt *time.Time
	if err := row.Scan(
		&item.ID, &item.Name, &item.KeyPrefix, &item.Scope, &item.MemberID,
		&item.ProjectID, &item.Status,
		&item.Budget, &createdAt, &expiresAt,
	); err != nil {
		return types.PlatformKey{}, err
	}
	item.Consumed = 0
	item.CreatedAt = formatDateOnly(createdAt)
	if expiresAt != nil {
		s := formatDateOnly(*expiresAt)
		item.ExpiresAt = &s
	}
	return item, nil
}
