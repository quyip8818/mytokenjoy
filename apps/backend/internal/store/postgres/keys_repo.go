package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgKeysRepo struct {
	db        dbQuerier
	allowlist *pgModelAllowlistRepo
}

func (r *pgKeysRepo) ProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, provider, name, key_prefix, secret_key, relay_channel_id, status,
			balance, last_used, rotate_enabled, created_at
		FROM provider_keys ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.ProviderKey, 0)
	for rows.Next() {
		var item types.ProviderKey
		var lastUsed *time.Time
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.Provider, &item.Name, &item.KeyPrefix, &item.SecretKey,
			&item.RelayChannelID, &item.Status, &item.Balance, &lastUsed,
			&item.RotateEnabled, &createdAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatDateOnly(createdAt)
		if lastUsed != nil {
			s := formatSyncLogTime(*lastUsed)
			item.LastUsed = &s
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneProviderKeys(items), nil
}

func (r *pgKeysRepo) SetProviderKeys(ctx context.Context, keys []types.ProviderKey) error {
	cloned := store.CloneProviderKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
		createdAt, err := parseAPITime(key.CreatedAt)
		if err != nil {
			return err
		}
		var lastUsed *time.Time
		if key.LastUsed != nil {
			t, err := parseAPITime(*key.LastUsed)
			if err != nil {
				return err
			}
			lastUsed = &t
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, relay_channel_id, status,
				balance, last_used, rotate_enabled, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			ON CONFLICT (id) DO UPDATE SET
				provider = EXCLUDED.provider,
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				secret_key = EXCLUDED.secret_key,
				relay_channel_id = EXCLUDED.relay_channel_id,
				status = EXCLUDED.status,
				balance = EXCLUDED.balance,
				last_used = EXCLUDED.last_used,
				rotate_enabled = EXCLUDED.rotate_enabled,
				updated_at = NOW()
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, key.SecretKey, key.RelayChannelID,
			key.Status, key.Balance, lastUsed, key.RotateEnabled, createdAt); err != nil {
			return fmt.Errorf("upsert provider key %s: %w", key.ID, err)
		}
	}
	return pruneByID(ctx, r.db, "provider_keys", ids)
}

func (r *pgKeysRepo) PlatformKeys(ctx context.Context) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, key_prefix, full_key, member_id,
			budget_group_id, status, quota, used, created_at, expires_at
		FROM platform_keys WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		var item types.PlatformKey
		var createdAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Name, &item.KeyPrefix, &item.FullKey, &item.MemberID,
			&item.BudgetGroupID, &item.Status,
			&item.Quota, &item.Used, &createdAt, &expiresAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatDateOnly(createdAt)
		if expiresAt != nil {
			s := formatDateOnly(*expiresAt)
			item.ExpiresAt = &s
		}
		modelRows, err := r.db.Query(ctx, `
			SELECT model_name FROM model_allowlist
			WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
			ORDER BY model_name
		`, companyID, types.AllowlistOwnerPlatformKey, item.ID)
		if err != nil {
			return nil, err
		}
		for modelRows.Next() {
			var modelName string
			if err := modelRows.Scan(&modelName); err != nil {
				modelRows.Close()
				return nil, err
			}
			item.ModelWhitelist = append(item.ModelWhitelist, modelName)
		}
		modelRows.Close()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.ClonePlatformKeys(items), nil
}

func (r *pgKeysRepo) SetPlatformKeys(ctx context.Context, keys []types.PlatformKey) error {
	companyID := store.CompanyID(ctx)
	cloned := store.ClonePlatformKeys(keys)
	ids := make([]string, len(cloned))
	for i, key := range cloned {
		ids[i] = key.ID
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
				id, company_id, name, key_prefix, full_key, member_id,
				budget_group_id, status, quota, used, created_at, expires_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				key_prefix = EXCLUDED.key_prefix,
				full_key = EXCLUDED.full_key,
				member_id = EXCLUDED.member_id,
				budget_group_id = EXCLUDED.budget_group_id,
				status = EXCLUDED.status,
				quota = EXCLUDED.quota,
				used = EXCLUDED.used,
				expires_at = EXCLUDED.expires_at,
				updated_at = NOW()
		`, key.ID, companyID, key.Name, key.KeyPrefix, key.FullKey, key.MemberID,
			key.BudgetGroupID, key.Status,
			key.Quota, key.Used, createdAt, expiresAt); err != nil {
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

func (r *pgKeysRepo) AddPlatformKeyUsed(ctx context.Context, keyID string, amountCNY float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE platform_keys SET used = used + $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, keyID, amountCNY)
	return err
}

func (r *pgKeysRepo) Approvals(ctx context.Context) ([]types.KeyApproval, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, type, applicant, applicant_id, department, reason, requested_quota,
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
			&item.Reason, &item.RequestedQuota, &item.Status, &item.Approver, &item.RejectReason,
			&createdAt, &resolvedAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatSyncLogTime(createdAt)
		if resolvedAt != nil {
			s := formatSyncLogTime(*resolvedAt)
			item.ResolvedAt = &s
		}
		modelRows, err := r.db.Query(ctx, `
			SELECT model_name FROM model_allowlist
			WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
			ORDER BY model_name
		`, companyID, types.AllowlistOwnerKeyApproval, item.ID)
		if err != nil {
			return nil, err
		}
		for modelRows.Next() {
			var modelName string
			if err := modelRows.Scan(&modelName); err != nil {
				modelRows.Close()
				return nil, err
			}
			item.RequestedModels = append(item.RequestedModels, modelName)
		}
		modelRows.Close()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneApprovals(items), nil
}

func (r *pgKeysRepo) SetApprovals(ctx context.Context, approvals []types.KeyApproval) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneApprovals(approvals)
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
				id, company_id, type, applicant, applicant_id, department, reason, requested_quota,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO UPDATE SET
				type = EXCLUDED.type,
				applicant = EXCLUDED.applicant,
				applicant_id = EXCLUDED.applicant_id,
				department = EXCLUDED.department,
				reason = EXCLUDED.reason,
				requested_quota = EXCLUDED.requested_quota,
				status = EXCLUDED.status,
				approver = EXCLUDED.approver,
				reject_reason = EXCLUDED.reject_reason,
				created_at = EXCLUDED.created_at,
				resolved_at = EXCLUDED.resolved_at
		`, approval.ID, companyID, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedQuota, approval.Status, approval.Approver,
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
	row := r.db.QueryRow(ctx, `
		SELECT id, name, key_prefix, full_key, member_id,
			budget_group_id, status, quota, used, created_at, expires_at
		FROM platform_keys WHERE company_id = $1 AND id = $2
	`, companyID, keyID)
	var item types.PlatformKey
	var createdAt time.Time
	var expiresAt *time.Time
	if err := row.Scan(
		&item.ID, &item.Name, &item.KeyPrefix, &item.FullKey, &item.MemberID,
		&item.BudgetGroupID, &item.Status,
		&item.Quota, &item.Used, &createdAt, &expiresAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CreatedAt = formatDateOnly(createdAt)
	if expiresAt != nil {
		s := formatDateOnly(*expiresAt)
		item.ExpiresAt = &s
	}
	modelRows, err := r.db.Query(ctx, `
		SELECT model_name FROM model_allowlist
		WHERE company_id = $1 AND owner_type = $2 AND owner_id = $3
		ORDER BY model_name
	`, companyID, types.AllowlistOwnerPlatformKey, item.ID)
	if err != nil {
		return nil, err
	}
	for modelRows.Next() {
		var modelName string
		if err := modelRows.Scan(&modelName); err != nil {
			modelRows.Close()
			return nil, err
		}
		item.ModelWhitelist = append(item.ModelWhitelist, modelName)
	}
	modelRows.Close()
	cloned := store.ClonePlatformKey(item)
	return &cloned, nil
}

func (r *pgKeysRepo) SumMemberKeyUsed(ctx context.Context, memberID string) (float64, error) {
	companyID := store.CompanyID(ctx)
	var total float64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(used), 0) FROM platform_keys
		WHERE company_id = $1 AND member_id = $2 AND budget_group_id IS NULL
	`, companyID, memberID).Scan(&total)
	return total, err
}

func (r *pgKeysRepo) ListActiveMemberKeys(ctx context.Context, memberID string) ([]types.PlatformKey, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, key_prefix, full_key, member_id,
			budget_group_id, status, quota, used, created_at, expires_at
		FROM platform_keys
		WHERE company_id = $1 AND member_id = $2 AND budget_group_id IS NULL AND status = 'active'
		ORDER BY id
	`, companyID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.PlatformKey, 0)
	for rows.Next() {
		var item types.PlatformKey
		var createdAt time.Time
		var expiresAt *time.Time
		if err := rows.Scan(
			&item.ID, &item.Name, &item.KeyPrefix, &item.FullKey, &item.MemberID,
			&item.BudgetGroupID, &item.Status,
			&item.Quota, &item.Used, &createdAt, &expiresAt,
		); err != nil {
			return nil, err
		}
		item.CreatedAt = formatDateOnly(createdAt)
		if expiresAt != nil {
			s := formatDateOnly(*expiresAt)
			item.ExpiresAt = &s
		}
		items = append(items, item)
	}
	return store.ClonePlatformKeys(items), rows.Err()
}
