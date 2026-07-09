package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/store"
)

const mappingSelect = `
	SELECT rm.company_id, rm.platform_key_id, rm.newapi_token_id,
	       pk.member_id, m.department_id, pk.budget_group_id,
	       rm.relay_group, rm.sync_status, rm.synced_at, rm.newapi_token_remain_quota
	FROM relay_mappings rm
	JOIN platform_keys pk ON pk.company_id = rm.company_id AND pk.id = rm.platform_key_id
	LEFT JOIN members m ON m.company_id = pk.company_id AND m.id = pk.member_id
`

func scanMapping(row pgx.Row) (store.RelayMapping, error) {
	var m store.RelayMapping
	var memberID, budgetGroupID, departmentID *string
	var tokenID, remainQuota *int64
	var syncedAt *time.Time
	err := row.Scan(
		&m.CompanyID, &m.PlatformKeyID, &tokenID, &memberID, &departmentID, &budgetGroupID,
		&m.RelayGroup, &m.SyncStatus, &syncedAt, &remainQuota,
	)
	if err != nil {
		return store.RelayMapping{}, err
	}
	m.NewAPITokenID = tokenID
	m.MemberID = memberID
	if departmentID != nil {
		m.DepartmentID = *departmentID
	}
	if m.DepartmentID == "" && strings.HasPrefix(m.RelayGroup, "dept-") {
		m.DepartmentID = strings.TrimPrefix(m.RelayGroup, "dept-")
	}
	m.BudgetGroupID = budgetGroupID
	m.SyncedAt = syncedAt
	m.NewAPITokenRemainQuota = remainQuota
	return m, nil
}

func (r *relayRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE rm.company_id = $1 AND rm.platform_key_id = $2`, companyID, platformKeyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *relayRepo) GetMappingByKeyHash(ctx context.Context, keyHash string) (*store.RelayMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE pk.key_hash = $1`, keyHash)
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
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE rm.company_id = $1 AND rm.newapi_token_id = $2`, companyID, tokenID)
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
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE rm.newapi_token_id = $1`, tokenID)
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
	return r.listMappings(ctx, "rm.company_id = $1 AND pk.member_id = $2", companyID, memberID)
}

func (r *relayRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "rm.company_id = $1 AND m.department_id = $2", companyID, departmentID)
}

func (r *relayRepo) ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "rm.company_id = $1 AND pk.budget_group_id = $2", companyID, budgetGroupID)
}

func (r *relayRepo) ListActiveMappings(ctx context.Context) ([]store.RelayMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "rm.company_id = $1 AND rm.sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.RelayMapping, error) {
	return r.listMappings(ctx, "rm.company_id = $1 AND rm.sync_status = $2", companyID, store.RelaySyncStatusSynced)
}

func (r *relayRepo) UpsertMapping(ctx context.Context, mapping store.RelayMapping) error {
	companyID := store.CompanyID(ctx)
	if mapping.CompanyID > 0 {
		companyID = mapping.CompanyID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO relay_mappings (
			company_id, platform_key_id, newapi_token_id,
			relay_group, sync_status, synced_at, newapi_token_remain_quota, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		ON CONFLICT (company_id, platform_key_id) DO UPDATE SET
			newapi_token_id = EXCLUDED.newapi_token_id,
			relay_group = EXCLUDED.relay_group,
			sync_status = EXCLUDED.sync_status,
			synced_at = EXCLUDED.synced_at,
			newapi_token_remain_quota = EXCLUDED.newapi_token_remain_quota,
			updated_at = NOW()
	`, companyID, mapping.PlatformKeyID, mapping.NewAPITokenID,
		mapping.RelayGroup, mapping.SyncStatus, mapping.SyncedAt, mapping.NewAPITokenRemainQuota)
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
		SET newapi_token_id = $3, sync_status = $4, newapi_token_remain_quota = $5,
		    synced_at = $6, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, tokenID, status, remainQuota, syncedAt)
	return err
}

func (r *relayRepo) UpdateMappingNewAPITokenRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE relay_mappings SET newapi_token_remain_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, remainQuota)
	return err
}
