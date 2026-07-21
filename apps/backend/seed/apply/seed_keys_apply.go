package apply

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgtime "github.com/tokenjoy/backend/internal/pkg/timeutil"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func insertSeedModelAllowlist(ctx context.Context, exec TableWriter, tid uuid.UUID, rows []store.ModelAllowlistRow) error {
	for _, row := range rows {
		if _, err := exec.Exec(ctx, `
			INSERT INTO model_allowlist (company_id, owner_type, owner_id, model_id)
			VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING
		`, tid, row.OwnerType, row.OwnerID, row.ModelID); err != nil {
			return fmt.Errorf("seed allowlist %s/%s: %w", row.OwnerType, row.OwnerID, err)
		}
	}
	return nil
}

func insertSeedKeys(ctx context.Context, exec TableWriter, tid uuid.UUID, snap store.Snapshot) error {
	for _, key := range snap.ProviderKeys {
		createdAt, err := pkgtime.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, newapi_channel_id, status,
				rotate_enabled, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO NOTHING
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, key.SecretKey, key.NewAPIChannelID,
			key.Status, key.RotateEnabled, createdAt); err != nil {
			return err
		}
	}
	for _, key := range snap.PlatformKeys {
		createdAt, err := pkgtime.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := pkgtime.Parse(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		keyHash := store.HashPlatformKey("pending:" + key.ID.String())
		if key.FullKey != nil && *key.FullKey != "" {
			keyHash = store.HashPlatformKey(*key.FullKey)
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO platform_keys (
				id, company_id, name, key_prefix, key_hash, scope, member_id,
				project_id, status, budget, created_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (company_id, id) DO NOTHING
		`, key.ID, tid, key.Name, key.KeyPrefix, keyHash, key.Scope, key.MemberID,
			key.ProjectID, key.Status,
			key.Budget, createdAt, expiresAt); err != nil {
			return err
		}
	}
	for _, approval := range snap.Approvals {
		createdAt, err := pkgtime.Parse(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := pkgtime.Parse(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO key_approvals (
				id, company_id, type, applicant, applicant_id, department, reason, requested_budget,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO NOTHING
		`, approval.ID, tid, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedBudget, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedModels(ctx context.Context, exec TableWriter, tid uuid.UUID, models []types.ModelInfo) error {
	for _, model := range models {
		companyID := model.CompanyID
		if companyID == uuid.Nil {
			companyID = contract.TokenJoyCompanyID
		}
		if companyID != contract.TokenJoyCompanyID && companyID != tid {
			continue
		}
		capabilities := model.Capabilities
		if capabilities == nil {
			capabilities = []string{}
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO models (
				model_id, company_id, provider, type, name, description, endpoint,
				max_context, enabled, capabilities
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (model_id) DO NOTHING
		`, model.ID, companyID, model.Provider, model.Type, model.Name,
			model.Description, model.Endpoint,
			model.MaxContext, model.Enabled,
			capabilities); err != nil {
			return err
		}
	}
	return nil
}

func insertSeedAudit(ctx context.Context, exec TableWriter, tid uuid.UUID, snap store.Snapshot) error {
	if _, err := exec.Exec(ctx, `
		INSERT INTO audit_settings (company_id, content_retention_enabled)
		VALUES ($1, $2) ON CONFLICT (company_id) DO NOTHING
	`, tid, snap.AuditSettings.ContentRetentionEnabled); err != nil {
		return err
	}
	for _, log := range snap.OperationLogs {
		createdAt, err := pkgtime.Parse(log.CreatedAt)
		if err != nil {
			return err
		}
		actorType := log.ActorType
		if actorType == "" {
			actorType = store.ActorTypeMember
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO operation_logs (id, company_id, action, operator, operator_id, actor_type, target, detail, ip, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (company_id, id, created_at) DO NOTHING
		`, log.ID, tid, log.Action, log.Operator, log.OperatorID, actorType, log.Target, log.Detail, log.IP, createdAt); err != nil {
			return err
		}
	}
	return nil
}
