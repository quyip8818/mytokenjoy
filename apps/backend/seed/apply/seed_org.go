package apply

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/sms"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func insertSeedMembers(ctx context.Context, exec TableWriter, tid uuid.UUID, members []types.Member, roleIDByName map[string]uuid.UUID) error {
	demoHash := contract.DemoPasswordHash()
	for _, member := range members {
		// Create user for this member (use member ID as user ID when not explicitly set).
		userID := member.UserID
		if userID == uuid.Nil {
			userID = member.ID
		}
		var passwordHash *string
		if member.Status == "active" && member.Email != "" {
			hash := demoHash
			passwordHash = &hash
		}
		var phone *string
		if member.Phone != "" {
			formatted := sms.FormatPhone(member.Phone)
			phone = &formatted
		}
		var email *string
		if member.Email != "" {
			email = &member.Email
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO users (id, phone, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'active')
			ON CONFLICT (id) DO NOTHING
		`, userID, phone, email, passwordHash); err != nil {
			return fmt.Errorf("seed user for member %s: %w", member.ID, err)
		}

		if _, err := exec.Exec(ctx, `
			INSERT INTO members (
				id, company_id, user_id, name, department_id,
				status, source, external_id, personal_budget
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (company_id, id) DO NOTHING
		`, member.ID, member.CompanyID, userID, member.Name,
			member.DepartmentID, member.Status, member.Source, member.ExternalID, member.PersonalBudget); err != nil {
			return fmt.Errorf("seed member %s: %w", member.ID, err)
		}
		for _, roleName := range member.Roles {
			roleID, ok := roleIDByName[roleName]
			if !ok {
				continue
			}
			if _, err := exec.Exec(ctx, `
				INSERT INTO member_roles (company_id, member_id, role_id) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, member.CompanyID, member.ID, roleID); err != nil {
				return fmt.Errorf("seed member role: %w", err)
			}
		}
	}
	return nil
}

func insertSeedOrgIntegration(ctx context.Context, exec TableWriter, tid uuid.UUID, snap store.Snapshot) error {
	integration := snap.OrgIntegration
	var platform *string
	if integration.Platform != nil {
		s := string(*integration.Platform)
		platform = &s
	}
	fieldMappingsJSON, err := json.Marshal(integration.FieldMappings)
	if err != nil {
		return fmt.Errorf("marshal field mappings: %w", err)
	}
	if _, err := exec.Exec(ctx, `
		INSERT INTO org_integration (
			company_id, platform, connected,
			enabled, start_time, frequency_hours,
			delete_member_threshold, delete_department_threshold,
			notify_phone, notify_email, notify_im,
			encrypted_credential, field_mappings
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (company_id) DO NOTHING
	`, tid, platform, integration.Connected,
		integration.Enabled, integration.StartTime, integration.FrequencyHours,
		integration.DeleteMemberThreshold, integration.DeleteDepartmentThreshold,
		integration.NotifyPhone, integration.NotifyEmail, integration.NotifyIm,
		integration.EncryptedCredential, fieldMappingsJSON); err != nil {
		return err
	}
	for _, log := range snap.SyncLogs {
		t, err := pkgtime.Parse(log.Time)
		if err != nil {
			return err
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_sync_logs (id, company_id, time, type, result, detail)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (company_id, id) DO NOTHING
		`, log.ID, tid, t, log.Type, log.Result, log.Detail); err != nil {
			return err
		}
	}
	for _, failure := range snap.ImportFailures {
		if _, err := exec.Exec(ctx, `
			INSERT INTO org_import_failures (id, company_id, name, employee_id, reason)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (company_id, id) DO NOTHING
		`, failure.ID, tid, failure.Name, failure.EmployeeID, failure.Reason); err != nil {
			return err
		}
	}
	return nil
}
