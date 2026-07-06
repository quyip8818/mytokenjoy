package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) Integration(ctx context.Context) (types.OrgIntegration, error) {
	companyID := store.CompanyID(ctx)
	var platform *string
	var connected bool
	var lastImport *time.Time
	var lastImportOK, lastImportFail *int
	var encrypted []byte
	var fieldMappingsJSON []byte
	var integration types.OrgIntegration
	err := r.db.QueryRow(ctx, `
		SELECT platform, connected, last_import, last_import_ok, last_import_fail,
			enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im, encrypted_credential, field_mappings
		FROM org_integration WHERE company_id = $1
	`, companyID).Scan(
		&platform, &connected, &lastImport, &lastImportOK, &lastImportFail,
		&integration.Enabled, &integration.StartTime, &integration.FrequencyHours,
		&integration.DeleteMemberThreshold, &integration.DeleteDepartmentThreshold,
		&integration.NotifyPhone, &integration.NotifyEmail, &integration.NotifyIm,
		&encrypted, &fieldMappingsJSON,
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
	mappings, err := decodeFieldMappings(fieldMappingsJSON)
	if err != nil {
		return types.OrgIntegration{}, err
	}
	integration.FieldMappings = mappings
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

func (r *pgOrgRepo) FieldMappings(ctx context.Context) ([]types.FieldMapping, error) {
	integration, err := r.Integration(ctx)
	if err != nil {
		return nil, err
	}
	return append([]types.FieldMapping{}, integration.FieldMappings...), nil
}

func (r *pgOrgRepo) SetFieldMappings(ctx context.Context, mappings []types.FieldMapping) error {
	companyID := store.CompanyID(ctx)
	payload, err := encodeFieldMappings(mappings)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		UPDATE org_integration SET field_mappings = $2, updated_at = NOW()
		WHERE company_id = $1
	`, companyID, payload)
	if err != nil {
		return fmt.Errorf("set field mappings: %w", err)
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
		INSERT INTO org_integration (company_id, platform, encrypted_credential, field_mappings, updated_at)
		VALUES ($1, $2, $3, '[]', NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			platform = EXCLUDED.platform,
			encrypted_credential = EXCLUDED.encrypted_credential,
			field_mappings = '[]',
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
		UPDATE org_integration
		SET encrypted_credential = NULL, field_mappings = '[]', updated_at = NOW()
		WHERE company_id = $1
	`, companyID)
	if err != nil {
		return fmt.Errorf("clear credential: %w", err)
	}
	return nil
}

func encodeFieldMappings(mappings []types.FieldMapping) ([]byte, error) {
	if mappings == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(mappings)
}

func decodeFieldMappings(data []byte) ([]types.FieldMapping, error) {
	if len(data) == 0 {
		return []types.FieldMapping{}, nil
	}
	var mappings []types.FieldMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, fmt.Errorf("decode field mappings: %w", err)
	}
	return mappings, nil
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
