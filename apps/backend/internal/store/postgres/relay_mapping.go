package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *relayRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE company_id = $1 AND platform_key_id = $2`, companyID, platformKeyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByFullKey(ctx context.Context, fullKey string) (*store.RelayMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+`
		WHERE (company_id, platform_key_id) IN (
			SELECT company_id, id FROM platform_keys WHERE full_key = $1
		)
	`, fullKey)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE company_id = $1 AND newapi_token_id = $2`, companyID, tokenID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*store.RelayMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE newapi_token_id = $1`, tokenID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) listMappings(ctx context.Context, where string, args ...any) ([]store.RelayMapping, error) {
	query := mappingSelect
	if where != "" {
		query += " WHERE " + where
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.RelayMapping, 0)
	for rows.Next() {
		m, err := scanMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *relayRepo) ListMappingsByMemberID(ctx context.Context, memberID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND member_id = $2", companyID, memberID)
}

func (r *relayRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND department_id = $2", companyID, departmentID)
}

func (r *relayRepo) ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND budget_group_id = $2", companyID, budgetGroupID)
}

func (r *relayRepo) ListActiveMappings(ctx context.Context) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "company_id = $1 AND sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.RelayMapping, error) {
	return r.listMappings(ctx, "company_id = $1 AND sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) UpsertMapping(ctx context.Context, mapping store.RelayMapping) error {
	companyID := store.CompanyID(ctx)
	if mapping.CompanyID > 0 {
		companyID = mapping.CompanyID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO relay_mappings (
			platform_key_id, company_id, newapi_token_id, member_id, department_id, budget_group_id,
			relay_group, sync_status, synced_at, relay_remain_quota, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
		ON CONFLICT (company_id, platform_key_id) DO UPDATE SET
			newapi_token_id = EXCLUDED.newapi_token_id,
			member_id = EXCLUDED.member_id,
			department_id = EXCLUDED.department_id,
			budget_group_id = EXCLUDED.budget_group_id,
			relay_group = EXCLUDED.relay_group,
			sync_status = EXCLUDED.sync_status,
			synced_at = EXCLUDED.synced_at,
			relay_remain_quota = EXCLUDED.relay_remain_quota,
			updated_at = NOW()
	`, mapping.PlatformKeyID, companyID, mapping.NewAPITokenID, mapping.MemberID, mapping.DepartmentID,
		mapping.BudgetGroupID, mapping.RelayGroup, mapping.SyncStatus, mapping.SyncedAt, mapping.RelayRemainQuota)
	return err
}

func (r *relayRepo) UpdateMappingSync(
	ctx context.Context,
	platformKeyID string,
	tokenID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE relay_mappings
		SET newapi_token_id = $3, sync_status = $4, relay_remain_quota = $5,
		    synced_at = $6, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, tokenID, status, remainQuota, syncedAt)
	return err
}

func (r *relayRepo) UpdateMappingRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE relay_mappings SET relay_remain_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, remainQuota)
	return err
}
