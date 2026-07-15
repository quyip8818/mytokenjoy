package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type platformKeyMappingRepo struct {
	db dbQuerier
}

func newPlatformKeyMappingRepo(db dbQuerier) *platformKeyMappingRepo {
	return &platformKeyMappingRepo{db: db}
}

var _ store.PlatformKeyMappingRepository = (*platformKeyMappingRepo)(nil)

const mappingSelect = `
	SELECT pkm.company_id, pkm.platform_key_id, pkm.newapi_key_id,
	       pk.member_id, m.department_id, pk.project_id,
	       pkm.newapi_group, pkm.sync_status, pkm.synced_at, pkm.newapi_key_remain_quota
	FROM platform_key_mappings pkm
	JOIN platform_keys pk ON pk.company_id = pkm.company_id AND pk.id = pkm.platform_key_id
	LEFT JOIN members m ON m.company_id = pk.company_id AND m.id = pk.member_id
`

func scanMapping(row pgx.Row) (store.PlatformKeyMapping, error) {
	var m store.PlatformKeyMapping
	var memberID, projectID, departmentID *string
	var keyID, remainQuota *int64
	var syncedAt *time.Time
	err := row.Scan(
		&m.CompanyID, &m.PlatformKeyID, &keyID, &memberID, &departmentID, &projectID,
		&m.NewAPIGroup, &m.SyncStatus, &syncedAt, &remainQuota,
	)
	if err != nil {
		return store.PlatformKeyMapping{}, err
	}
	m.NewAPIKeyID = keyID
	m.MemberID = memberID
	if departmentID != nil {
		m.DepartmentID = *departmentID
	}
	if m.DepartmentID == "" && strings.HasPrefix(m.NewAPIGroup, common.NewAPIGroupPrefix) {
		m.DepartmentID = strings.TrimPrefix(m.NewAPIGroup, common.NewAPIGroupPrefix)
	}
	m.ProjectID = projectID
	m.SyncedAt = syncedAt
	m.NewAPIKeyRemainQuota = remainQuota
	return m, nil
}

func (r *platformKeyMappingRepo) GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*store.PlatformKeyMapping, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE pkm.company_id = $1 AND pkm.platform_key_id = $2`, companyID, platformKeyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *platformKeyMappingRepo) GetMappingByKeyHash(ctx context.Context, keyHash string) (*store.PlatformKeyMapping, error) {
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

func (r *platformKeyMappingRepo) FindMappingByNewAPIKeyID(ctx context.Context, keyID int64) (*store.PlatformKeyMapping, error) {
	row := r.db.QueryRow(ctx, mappingSelect+` WHERE pkm.newapi_key_id = $1`, keyID)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *platformKeyMappingRepo) ListMappingsByNewAPIKeyIDs(ctx context.Context, keyIDs []int64) ([]store.PlatformKeyMapping, error) {
	if len(keyIDs) == 0 {
		return nil, nil
	}
	return r.listMappings(ctx, "pkm.newapi_key_id = ANY($1)", keyIDs)
}

func (r *platformKeyMappingRepo) listMappings(ctx context.Context, where string, args ...any) ([]store.PlatformKeyMapping, error) {
	query := mappingSelect
	if where != "" {
		query += " WHERE " + where
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]store.PlatformKeyMapping, 0)
	for rows.Next() {
		m, err := scanMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *platformKeyMappingRepo) ListMappingsByMemberID(ctx context.Context, memberID string) ([]store.PlatformKeyMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "pkm.company_id = $1 AND pk.member_id = $2", companyID, memberID)
}

func (r *platformKeyMappingRepo) ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]store.PlatformKeyMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "pkm.company_id = $1 AND m.department_id = $2", companyID, departmentID)
}

func (r *platformKeyMappingRepo) ListMappingsByProjectID(ctx context.Context, projectID string) ([]store.PlatformKeyMapping, error) {
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "pkm.company_id = $1 AND pk.project_id = $2", companyID, projectID)
}

func (r *platformKeyMappingRepo) ListMappingsByPlatformKeyIDs(ctx context.Context, platformKeyIDs []string) ([]store.PlatformKeyMapping, error) {
	if len(platformKeyIDs) == 0 {
		return nil, nil
	}
	companyID := store.CompanyID(ctx)
	return r.listMappings(ctx, "pkm.company_id = $1 AND pkm.platform_key_id = ANY($2)", companyID, platformKeyIDs)
}

func (r *platformKeyMappingRepo) ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]store.PlatformKeyMapping, error) {
	return r.listMappings(ctx, "pkm.company_id = $1 AND pkm.sync_status = $2", companyID, store.MappingSyncStatusSynced)
}

func (r *platformKeyMappingRepo) UpsertMapping(ctx context.Context, mapping store.PlatformKeyMapping) error {
	companyID := store.CompanyID(ctx)
	if mapping.CompanyID > 0 {
		companyID = mapping.CompanyID
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO platform_key_mappings (
			company_id, platform_key_id, newapi_key_id,
			newapi_group, sync_status, synced_at, newapi_key_remain_quota, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
		ON CONFLICT (company_id, platform_key_id) DO UPDATE SET
			newapi_key_id = EXCLUDED.newapi_key_id,
			newapi_group = EXCLUDED.newapi_group,
			sync_status = EXCLUDED.sync_status,
			synced_at = EXCLUDED.synced_at,
			newapi_key_remain_quota = EXCLUDED.newapi_key_remain_quota,
			updated_at = NOW()
	`, companyID, mapping.PlatformKeyID, mapping.NewAPIKeyID,
		mapping.NewAPIGroup, mapping.SyncStatus, mapping.SyncedAt, mapping.NewAPIKeyRemainQuota)
	return err
}

func (r *platformKeyMappingRepo) UpdateMappingSync(
	ctx context.Context,
	platformKeyID string,
	keyID int64,
	status string,
	remainQuota *int64,
	syncedAt time.Time,
) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE platform_key_mappings
		SET newapi_key_id = $3, sync_status = $4, newapi_key_remain_quota = $5,
		    synced_at = $6, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, keyID, status, remainQuota, syncedAt)
	return err
}

func (r *platformKeyMappingRepo) UpdateMappingNewAPIKeyRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE platform_key_mappings SET newapi_key_remain_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND platform_key_id = $2
	`, companyID, platformKeyID, remainQuota)
	return err
}
