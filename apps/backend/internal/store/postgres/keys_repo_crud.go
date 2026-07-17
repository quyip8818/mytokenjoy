package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgKeysRepo) ProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, provider, name, key_prefix, secret_key, newapi_channel_id, status,
			rotate_enabled, created_at
		FROM provider_keys ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.ProviderKey, 0)
	for rows.Next() {
		var item types.ProviderKey
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Provider, &item.Name, &item.KeyPrefix, &item.SecretKey,
			&item.NewAPIChannelID, &item.Status,
			&item.RotateEnabled, &createdAt,
		); err != nil {
			return nil, err
		}
		item.SecretKey, err = common.DecryptField(r.credentialKey, item.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt provider key %s: %w", item.ID, err)
		}
		item.CreatedAt = formatDateOnly(createdAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgKeysRepo) SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error {
	cloned := cloneProviderKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
		createdAt, err := parseAPITime(key.CreatedAt)
		if err != nil {
			return err
		}
		storedSecret, err := common.EncryptField(r.credentialKey, key.SecretKey)
		if err != nil {
			return fmt.Errorf("encrypt provider key %s: %w", key.ID, err)
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, newapi_channel_id, status,
				rotate_enabled, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				secret_key = EXCLUDED.secret_key,
				newapi_channel_id = EXCLUDED.newapi_channel_id,
				status = EXCLUDED.status,
				rotate_enabled = EXCLUDED.rotate_enabled,
				updated_at = NOW()
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, storedSecret, key.NewAPIChannelID,
			key.Status, key.RotateEnabled, createdAt); err != nil {
			return fmt.Errorf("upsert provider key %s: %w", key.ID, err)
		}
	}
	return pruneByID(ctx, r.db, "provider_keys", ids)
}

func (r *pgKeysRepo) PlatformKeys(ctx context.Context) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, platformKeyListSelect+`
		WHERE pk.company_id = $1
		GROUP BY pk.id, pk.name, pk.key_prefix, pk.scope, pk.member_id, pk.project_id, pk.status, pk.budget, pk.created_at, pk.expires_at
		ORDER BY pk.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		item, err := scanPlatformKeyWithModels(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *pgKeysRepo) SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error {
	companyID := store.CompanyID(ctx)
	cloned := clonePlatformKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
		keyHash, err := r.resolvePlatformKeyHash(ctx, companyID, key)
		if err != nil {
			return err
		}
		createdAt, err := parseAPITime(key.CreatedAt)
		if err != nil {
			return err
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := parseAPITime(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO platform_keys (
				id, company_id, name, key_prefix, key_hash, scope, member_id,
				project_id, status, budget, created_at, expires_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				key_hash = EXCLUDED.key_hash,
				scope = EXCLUDED.scope,
				member_id = EXCLUDED.member_id,
				project_id = EXCLUDED.project_id,
				status = EXCLUDED.status,
				budget = EXCLUDED.budget,
				expires_at = EXCLUDED.expires_at,
				updated_at = NOW()
		`, key.ID, companyID, key.Name, key.KeyPrefix, keyHash, key.Scope, key.MemberID,
			key.ProjectID, key.Status,
			key.Budget, createdAt, expiresAt); err != nil {
			return fmt.Errorf("upsert platform key %s: %w", key.ID, err)
		}
		if err := r.allowlist.Replace(ctx, types.AllowlistOwnerPlatformKey, key.ID, key.ModelWhitelist); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		if err := pruneAllowlistByOwnerIDs(ctx, r.db, companyID, types.AllowlistOwnerPlatformKey, nil); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM platform_keys WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneAllowlistByOwnerIDs(ctx, r.db, companyID, types.AllowlistOwnerPlatformKey, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "platform_keys", companyID, ids)
}

func (r *pgKeysRepo) Approvals(ctx context.Context) ([]types.KeyApproval, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, type, applicant, applicant_id, department, reason, requested_budget,
			status, approver, reject_reason, created_at, resolved_at
		FROM key_approvals WHERE company_id = $1 ORDER BY created_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.KeyApproval, 0)
	for rows.Next() {
		var item types.KeyApproval
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Type, &item.Applicant, &item.ApplicantID, &item.Department,
			&item.Reason, &item.RequestedBudget, &item.Status, &item.Approver, &item.RejectReason,
			&createdAt, &resolvedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		if resolvedAt != nil {
			s := formatSyncLogTime(*resolvedAt)
			item.ResolvedAt = &s
		}
		models, err := r.allowlist.List(ctx, types.AllowlistOwnerKeyApproval, item.ID)
		if err != nil {
			return nil, err
		}
		item.RequestedModels = models
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgKeysRepo) SetApprovals(ctx context.Context, approvals []types.KeyApproval) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneApprovals(approvals)
	ids := make([]string, len(cloned))
	for i, approval := range cloned {
		ids[i] = approval.ID
		createdAt, err := parseAPITime(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := parseAPITime(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO key_approvals (
				id, company_id, type, applicant, applicant_id, department, reason, requested_budget,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO UPDATE SET
				type = EXCLUDED.type,
				applicant = EXCLUDED.applicant,
				applicant_id = EXCLUDED.applicant_id,
				department = EXCLUDED.department,
				reason = EXCLUDED.reason,
				requested_budget = EXCLUDED.requested_budget,
				status = EXCLUDED.status,
				approver = EXCLUDED.approver,
				reject_reason = EXCLUDED.reject_reason,
				created_at = EXCLUDED.created_at,
				resolved_at = EXCLUDED.resolved_at
		`, approval.ID, companyID, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedBudget, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return fmt.Errorf("upsert approval %s: %w", approval.ID, err)
		}
		if err := r.allowlist.Replace(ctx, types.AllowlistOwnerKeyApproval, approval.ID, approval.RequestedModels); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		if err := pruneAllowlistByOwnerIDs(ctx, r.db, companyID, types.AllowlistOwnerKeyApproval, nil); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM key_approvals WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneAllowlistByOwnerIDs(ctx, r.db, companyID, types.AllowlistOwnerKeyApproval, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "key_approvals", companyID, ids)
}

func (r *pgKeysRepo) PlatformKeyByID(ctx context.Context, keyID string) (*types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, platformKeySelect+` WHERE company_id = $1 AND id = $2`, companyID, keyID)
	item, err := scanPlatformKeyRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	models, err := r.allowlist.List(ctx, types.AllowlistOwnerPlatformKey, item.ID)
	if err != nil {
		return nil, err
	}
	item.ModelWhitelist = models
	return &item, nil
}

func (r *pgKeysRepo) PlatformKeyHashByID(ctx context.Context, keyID string) (string, bool, error) {
	companyID := store.CompanyID(ctx)
	var keyHash string
	err := r.db.QueryRow(ctx,
		`SELECT key_hash FROM platform_keys WHERE company_id = $1 AND id = $2`,
		companyID, keyID,
	).Scan(&keyHash)
	if err == pgx.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return keyHash, true, nil
}

func (r *pgKeysRepo) PlatformKeyByHash(ctx context.Context, keyHash string) (*types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, platformKeySelect+` WHERE company_id = $1 AND key_hash = $2`, companyID, keyHash)
	item, err := scanPlatformKeyRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	models, err := r.allowlist.List(ctx, types.AllowlistOwnerPlatformKey, item.ID)
	if err != nil {
		return nil, err
	}
	item.ModelWhitelist = models
	return &item, nil
}
