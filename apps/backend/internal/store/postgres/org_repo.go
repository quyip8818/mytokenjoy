package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgRepo struct {
	db    dbQuerier
	nodes *pgOrgNodeRepo
}

func (r *pgOrgRepo) Nodes() store.OrgNodeRepository {
	return r.nodes
}

func (r *pgOrgRepo) Permissions(ctx context.Context) ([]types.Permission, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, grp FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Permission, 0)
	for rows.Next() {
		var item types.Permission
		if err := rows.Scan(&item.ID, &item.Name, &item.Group); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.ClonePermissions(items), nil
}

func (r *pgOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id, personal_quota
		FROM members WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Member, 0)
	for rows.Next() {
		var item types.Member
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Phone, &item.Email,
			&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
			&item.PersonalQuota,
		); err != nil {
			return nil, err
		}
		roleRows, err := r.db.Query(ctx, `
			SELECT ro.name FROM member_roles mr
			JOIN roles ro ON ro.company_id = mr.company_id AND ro.id = mr.role_id
			WHERE mr.company_id = $1 AND mr.member_id = $2
			ORDER BY ro.name
		`, companyID, item.ID)
		if err == nil {
			for roleRows.Next() {
				var name string
				if err := roleRows.Scan(&name); err == nil {
					item.Roles = append(item.Roles, name)
				}
			}
			roleRows.Close()
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneMembers(items), nil
}

func (r *pgOrgRepo) MemberByID(ctx context.Context, memberID string) (*types.Member, error) {
	companyID := store.CompanyID(ctx)
	row := r.db.QueryRow(ctx, `
		SELECT id, name, phone, email, department_id, department_name, status, source, external_id, personal_quota
		FROM members WHERE company_id = $1 AND id = $2
	`, companyID, memberID)
	var item types.Member
	if err := row.Scan(
		&item.ID, &item.Name, &item.Phone, &item.Email,
		&item.DepartmentID, &item.DepartmentName, &item.Status, &item.Source, &item.ExternalID,
		&item.PersonalQuota,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	roleRows, err := r.db.Query(ctx, `
		SELECT ro.name FROM member_roles mr
		JOIN roles ro ON ro.company_id = mr.company_id AND ro.id = mr.role_id
		WHERE mr.company_id = $1 AND mr.member_id = $2
		ORDER BY ro.name
	`, companyID, item.ID)
	if err == nil {
		for roleRows.Next() {
			var name string
			if err := roleRows.Scan(&name); err == nil {
				item.Roles = append(item.Roles, name)
			}
		}
		roleRows.Close()
	}
	cloned := store.CloneMember(item)
	return &cloned, nil
}

func (r *pgOrgRepo) MemberPersonalQuota(ctx context.Context, memberID string) (float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var quota float64
	err := r.db.QueryRow(ctx, `
		SELECT personal_quota FROM members WHERE company_id = $1 AND id = $2
	`, companyID, memberID).Scan(&quota)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return quota, true, nil
}

func (r *pgOrgRepo) SetMembers(ctx context.Context, members []types.Member) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneMembers(members)
	roleIDByName, err := loadRoleNameIndex(ctx, r.db, companyID)
	if err != nil {
		return err
	}
	ids := make([]string, len(cloned))
	for i, member := range cloned {
		ids[i] = member.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO members (
				id, company_id, name, phone, email, department_id, department_name,
				status, source, external_id, personal_quota, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				phone = EXCLUDED.phone,
				email = EXCLUDED.email,
				department_id = EXCLUDED.department_id,
				department_name = EXCLUDED.department_name,
				status = EXCLUDED.status,
				source = EXCLUDED.source,
				external_id = EXCLUDED.external_id,
				personal_quota = EXCLUDED.personal_quota,
				updated_at = NOW()
		`, member.ID, companyID, member.Name, member.Phone, member.Email,
			member.DepartmentID, member.DepartmentName, member.Status, member.Source, member.ExternalID, member.PersonalQuota); err != nil {
			return fmt.Errorf("upsert member %s: %w", member.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM member_roles WHERE company_id = $1 AND member_id = $2`, companyID, member.ID); err != nil {
			return fmt.Errorf("clear member roles: %w", err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := r.db.Exec(ctx, `
				INSERT INTO member_roles (company_id, member_id, role_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, companyID, member.ID, roleID); err != nil {
				return fmt.Errorf("insert member role: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM member_roles WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM members WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "member_roles", "member_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "members", companyID, ids)
}

func (r *pgOrgRepo) UpdateMemberPersonalQuota(ctx context.Context, memberID string, personalQuota float64) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE members SET personal_quota = $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, memberID, personalQuota)
	if err != nil {
		return fmt.Errorf("update member personal quota: %w", err)
	}
	return nil
}

func loadRoleNameIndex(ctx context.Context, db dbQuerier, companyID int64) (map[string]string, error) {
	rows, err := db.Query(ctx, `SELECT id, name FROM roles WHERE company_id = $1`, companyID)
	if err != nil {
		return nil, fmt.Errorf("load roles index: %w", err)
	}
	defer rows.Close()
	index := make(map[string]string)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		index[name] = id
	}
	return index, rows.Err()
}

func (r *pgOrgRepo) SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE members SET password_hash = $3, updated_at = NOW()
		WHERE company_id = $1 AND id = $2
	`, companyID, memberID, passwordHash)
	if err != nil {
		return fmt.Errorf("set member password: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) Roles(ctx context.Context) ([]types.Role, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `SELECT id, name, type, member_count FROM roles WHERE company_id = $1 ORDER BY id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.Role, 0)
	for rows.Next() {
		var role types.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Type, &role.MemberCount); err != nil {
			return nil, err
		}
		grantRows, err := r.db.Query(ctx, `
			SELECT permission_ref FROM role_permission_grants WHERE company_id = $1 AND role_id = $2 ORDER BY permission_ref
		`, companyID, role.ID)
		if err == nil {
			for grantRows.Next() {
				var ref string
				if err := grantRows.Scan(&ref); err == nil {
					role.Permissions = append(role.Permissions, ref)
				}
			}
			grantRows.Close()
		}
		items = append(items, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneRoles(items), nil
}

func (r *pgOrgRepo) SetRoles(ctx context.Context, roles []types.Role) error {
	companyID := store.CompanyID(ctx)
	cloned := store.CloneRoles(roles)
	ids := make([]string, len(cloned))
	for i, role := range cloned {
		ids[i] = role.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO roles (id, company_id, name, type, member_count) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				type = EXCLUDED.type,
				member_count = EXCLUDED.member_count
		`, role.ID, companyID, role.Name, role.Type, role.MemberCount); err != nil {
			return fmt.Errorf("upsert role %s: %w", role.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM role_permission_grants WHERE company_id = $1 AND role_id = $2`, companyID, role.ID); err != nil {
			return fmt.Errorf("clear role grants: %w", err)
		}
		for _, perm := range role.Permissions {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO role_permission_grants (company_id, role_id, permission_ref) VALUES ($1, $2, $3)
			`, companyID, role.ID, perm); err != nil {
				return fmt.Errorf("insert role grant: %w", err)
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM role_permission_grants WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM roles WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "role_permission_grants", "role_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "roles", companyID, ids)
}

func (r *pgOrgRepo) Integration(ctx context.Context) (types.OrgIntegration, error) {
	companyID := store.CompanyID(ctx)
	var platform *string
	var connected bool
	var lastImport *time.Time
	var lastImportOK, lastImportFail *int
	var encrypted []byte
	var integration types.OrgIntegration
	err := r.db.QueryRow(ctx, `
		SELECT platform, connected, last_import, last_import_ok, last_import_fail,
			enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im, encrypted_credential
		FROM org_integration WHERE company_id = $1
	`, companyID).Scan(
		&platform, &connected, &lastImport, &lastImportOK, &lastImportFail,
		&integration.Enabled, &integration.StartTime, &integration.FrequencyHours,
		&integration.DeleteMemberThreshold, &integration.DeleteDepartmentThreshold,
		&integration.NotifyPhone, &integration.NotifyEmail, &integration.NotifyIm,
		&encrypted,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.OrgIntegration{}, nil
		}
		return types.OrgIntegration{}, err
	}
	integration.Connected = connected
	if platform != nil && *platform != "" {
		p := types.Platform(*platform)
		integration.Platform = &p
	}
	if lastImport != nil {
		s := formatSyncLogTime(*lastImport)
		integration.LastImport = &s
	}
	integration.LastImportOK = lastImportOK
	integration.LastImportFail = lastImportFail
	if len(encrypted) > 0 {
		integration.EncryptedCredential = append([]byte(nil), encrypted...)
	}
	return integration, nil
}

func (r *pgOrgRepo) SetIntegration(ctx context.Context, integration types.OrgIntegration) error {
	companyID := store.CompanyID(ctx)
	var platform *string
	if integration.Platform != nil {
		s := string(*integration.Platform)
		platform = &s
	}
	var lastImport *time.Time
	if integration.LastImport != nil {
		t, err := parseAPITime(*integration.LastImport)
		if err != nil {
			return err
		}
		lastImport = &t
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO org_integration (
			company_id, platform, connected, last_import, last_import_ok, last_import_fail,
			enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			platform = EXCLUDED.platform,
			connected = EXCLUDED.connected,
			last_import = EXCLUDED.last_import,
			last_import_ok = EXCLUDED.last_import_ok,
			last_import_fail = EXCLUDED.last_import_fail,
			enabled = EXCLUDED.enabled,
			start_time = EXCLUDED.start_time,
			frequency_hours = EXCLUDED.frequency_hours,
			delete_member_threshold = EXCLUDED.delete_member_threshold,
			delete_department_threshold = EXCLUDED.delete_department_threshold,
			notify_phone = EXCLUDED.notify_phone,
			notify_email = EXCLUDED.notify_email,
			notify_im = EXCLUDED.notify_im,
			updated_at = NOW()
	`, companyID, platform, integration.Connected, lastImport,
		integration.LastImportOK, integration.LastImportFail,
		integration.Enabled, integration.StartTime, integration.FrequencyHours,
		integration.DeleteMemberThreshold, integration.DeleteDepartmentThreshold,
		integration.NotifyPhone, integration.NotifyEmail, integration.NotifyIm)
	if err != nil {
		return fmt.Errorf("upsert org integration: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) GetIntegrationCredential(ctx context.Context) (*types.StoredCredential, error) {
	integration, err := r.Integration(ctx)
	if err != nil {
		return nil, err
	}
	return integration.ToStoredCredential(), nil
}

func (r *pgOrgRepo) SaveIntegrationCredential(ctx context.Context, platform types.Platform, encrypted []byte) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO org_integration (company_id, platform, encrypted_credential, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			platform = EXCLUDED.platform,
			encrypted_credential = EXCLUDED.encrypted_credential,
			updated_at = NOW()
	`, companyID, string(platform), encrypted)
	if err != nil {
		return fmt.Errorf("save credential: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) ClearIntegrationCredential(ctx context.Context) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		UPDATE org_integration SET encrypted_credential = NULL, updated_at = NOW()
		WHERE company_id = $1
	`, companyID)
	if err != nil {
		return fmt.Errorf("clear credential: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) ImportFailures(ctx context.Context) ([]types.ImportFailure, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, employee_id, reason FROM org_import_failures WHERE company_id = $1
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.ImportFailure, 0)
	for rows.Next() {
		var item types.ImportFailure
		if err := rows.Scan(&item.ID, &item.Name, &item.EmployeeID, &item.Reason); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneImportFailures(items), nil
}

func (r *pgOrgRepo) SetImportFailures(ctx context.Context, failures []types.ImportFailure) error {
	companyID := store.CompanyID(ctx)
	if _, err := r.db.Exec(ctx, `DELETE FROM org_import_failures WHERE company_id = $1`, companyID); err != nil {
		return fmt.Errorf("clear import failures: %w", err)
	}
	for _, item := range failures {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO org_import_failures (id, company_id, name, employee_id, reason)
			VALUES ($1, $2, $3, $4, $5)
		`, item.ID, companyID, item.Name, item.EmployeeID, item.Reason); err != nil {
			return fmt.Errorf("insert import failure: %w", err)
		}
	}
	return nil
}

func (r *pgOrgRepo) SyncLogs(ctx context.Context) ([]types.SyncLog, error) {
	companyID := store.CompanyID(ctx)
	rows, err := r.db.Query(ctx, `
		SELECT id, time, type, result, detail
		FROM org_sync_logs
		WHERE company_id = $1
		ORDER BY time DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]types.SyncLog, 0)
	for rows.Next() {
		var item types.SyncLog
		var t time.Time
		if err := rows.Scan(&item.ID, &t, &item.Type, &item.Result, &item.Detail); err != nil {
			return nil, err
		}
		item.Time = formatSyncLogTime(t)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return store.CloneSyncLogs(items), nil
}

func (r *pgOrgRepo) AppendSyncLog(ctx context.Context, log types.SyncLog) error {
	companyID := store.CompanyID(ctx)
	t, err := parseAPITime(log.Time)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO org_sync_logs (id, company_id, time, type, result, detail)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (company_id, id) DO UPDATE SET
			time = EXCLUDED.time,
			type = EXCLUDED.type,
			result = EXCLUDED.result,
			detail = EXCLUDED.detail
	`, log.ID, companyID, t, log.Type, log.Result, log.Detail)
	if err != nil {
		return fmt.Errorf("append sync log: %w", err)
	}
	return nil
}
