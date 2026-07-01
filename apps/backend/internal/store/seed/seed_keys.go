package seed

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/timeparse"
)

func insertKeys(ctx context.Context, exec tableWriter, tid int64, snap store.Snapshot) error {
	for _, key := range snap.ProviderKeys {
		createdAt, err := timeparse.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var lastUsed *time.Time
		if key.LastUsed != nil {
			t, err := timeparse.Parse(*key.LastUsed)
			if err != nil {
				return err
			}
			lastUsed = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO provider_keys (
				id, provider, name, key_prefix, secret_key, relay_channel_id, status,
				balance, last_used, rotate_enabled, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING
		`, key.ID, key.Provider, key.Name, key.KeyPrefix, key.SecretKey, key.RelayChannelID,
			key.Status, key.Balance, lastUsed, key.RotateEnabled, createdAt); err != nil {
			return err
		}
	}
	for _, key := range snap.PlatformKeys {
		createdAt, err := timeparse.Parse(key.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}
		var expiresAt *time.Time
		if key.ExpiresAt != nil {
			t, err := timeparse.Parse(*key.ExpiresAt)
			if err != nil {
				return err
			}
			expiresAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO platform_keys (
				id, company_id, name, key_prefix, full_key, member_id, member_name, app_name,
				budget_group_id, budget_group_name, status, quota, used, created_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (company_id, id) DO NOTHING
		`, key.ID, tid, key.Name, key.KeyPrefix, key.FullKey, key.MemberID, key.MemberName,
			key.AppName, key.BudgetGroupID, key.BudgetGroupName, key.Status,
			key.Quota, key.Used, createdAt, expiresAt); err != nil {
			return err
		}
		for _, modelName := range key.ModelWhitelist {
			if _, err := exec.Exec(ctx, `
				INSERT INTO platform_key_models (company_id, platform_key_id, model_name) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, key.ID, modelName); err != nil {
				return err
			}
		}
	}
	for _, approval := range snap.Approvals {
		createdAt, err := timeparse.Parse(approval.CreatedAt)
		if err != nil {
			return err
		}
		var resolvedAt *time.Time
		if approval.ResolvedAt != nil {
			t, err := timeparse.Parse(*approval.ResolvedAt)
			if err != nil {
				return err
			}
			resolvedAt = &t
		}
		if _, err := exec.Exec(ctx, `
			INSERT INTO key_approvals (
				id, company_id, type, applicant, applicant_id, department, reason, requested_quota,
				status, approver, reject_reason, created_at, resolved_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (company_id, id) DO NOTHING
		`, approval.ID, tid, approval.Type, approval.Applicant, approval.ApplicantID, approval.Department,
			approval.Reason, approval.RequestedQuota, approval.Status, approval.Approver,
			approval.RejectReason, createdAt, resolvedAt); err != nil {
			return err
		}
		for _, modelName := range approval.RequestedModels {
			if _, err := exec.Exec(ctx, `
				INSERT INTO key_approval_models (company_id, approval_id, model_name) VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, tid, approval.ID, modelName); err != nil {
				return err
			}
		}
	}
	return nil
}
