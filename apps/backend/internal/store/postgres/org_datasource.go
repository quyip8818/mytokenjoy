package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgOrgRepo) DataSourceStatus() types.DataSourceStatus {
	var platform *string
	var connected bool
	var lastImport *time.Time
	var lastImportOK, lastImportFail *int
	err := r.db.QueryRow(r.ctx, `
		SELECT platform, connected, last_import, last_import_ok, last_import_fail
		FROM org_data_source_status WHERE id = 1
	`).Scan(&platform, &connected, &lastImport, &lastImportOK, &lastImportFail)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.DataSourceStatus{}
		}
		return types.DataSourceStatus{}
	}
	status := types.DataSourceStatus{Connected: connected}
	if platform != nil && *platform != "" {
		p := types.Platform(*platform)
		status.Platform = &p
	}
	if lastImport != nil {
		s := formatSyncLogTime(*lastImport)
		status.LastImport = &s
	}
	if lastImportOK != nil || lastImportFail != nil {
		result := types.ImportResult{}
		if lastImportOK != nil {
			result.SuccessMembers = *lastImportOK
		}
		if lastImportFail != nil {
			result.Failures = make([]types.ImportFailure, *lastImportFail)
		}
		status.LastImportResult = &result
	}
	return status
}

func (r *pgOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	var platform *string
	if status.Platform != nil {
		s := string(*status.Platform)
		platform = &s
	}
	var lastImport *time.Time
	if status.LastImport != nil {
		t, err := parseAPITime(*status.LastImport)
		if err != nil {
			return err
		}
		lastImport = &t
	}
	var lastImportOK, lastImportFail *int
	if status.LastImportResult != nil {
		ok := status.LastImportResult.SuccessMembers
		lastImportOK = &ok
		fail := len(status.LastImportResult.Failures)
		lastImportFail = &fail
	}
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO org_data_source_status (id, platform, connected, last_import, last_import_ok, last_import_fail, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			platform = EXCLUDED.platform,
			connected = EXCLUDED.connected,
			last_import = EXCLUDED.last_import,
			last_import_ok = EXCLUDED.last_import_ok,
			last_import_fail = EXCLUDED.last_import_fail,
			updated_at = NOW()
	`, platform, status.Connected, lastImport, lastImportOK, lastImportFail)
	if err != nil {
		return fmt.Errorf("upsert data source status: %w", err)
	}
	return nil
}

func (r *pgOrgRepo) ImportFailures() []types.ImportFailure {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, name, employee_id, reason FROM org_import_failures
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.ImportFailure, 0)
	for rows.Next() {
		var item types.ImportFailure
		if err := rows.Scan(&item.ID, &item.Name, &item.EmployeeID, &item.Reason); err != nil {
			return nil
		}
		items = append(items, item)
	}
	return store.CloneImportFailures(items)
}

func (r *pgOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	if _, err := r.db.Exec(r.ctx, `DELETE FROM org_import_failures`); err != nil {
		return fmt.Errorf("clear import failures: %w", err)
	}
	for _, item := range failures {
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO org_import_failures (id, name, employee_id, reason)
			VALUES ($1, $2, $3, $4)
		`, item.ID, item.Name, item.EmployeeID, item.Reason); err != nil {
			return fmt.Errorf("insert import failure: %w", err)
		}
	}
	return nil
}
