package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	var policy types.OverrunPolicyConfig
	err := r.db.QueryRow(r.ctx, `
		SELECT thresholds, notify_email, notify_phone, notify_im, block_message
		FROM overrun_policy WHERE id = 1
	`).Scan(&policy.Thresholds, &policy.NotifyEmail, &policy.NotifyPhone, &policy.NotifyIm, &policy.BlockMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.OverrunPolicyConfig{}
		}
		return types.OverrunPolicyConfig{}
	}
	return policy
}

func (r *pgBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	_, err := r.db.Exec(r.ctx, `
		INSERT INTO overrun_policy (id, thresholds, notify_email, notify_phone, notify_im, block_message, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			thresholds = EXCLUDED.thresholds,
			notify_email = EXCLUDED.notify_email,
			notify_phone = EXCLUDED.notify_phone,
			notify_im = EXCLUDED.notify_im,
			block_message = EXCLUDED.block_message,
			updated_at = NOW()
	`, policy.Thresholds, policy.NotifyEmail, policy.NotifyPhone, policy.NotifyIm, policy.BlockMessage)
	if err != nil {
		return fmt.Errorf("upsert overrun policy: %w", err)
	}
	return nil
}

func (r *pgBudgetRepo) AlertRules() []types.AlertRule {
	rows, err := r.db.Query(r.ctx, `
		SELECT id, node_id, node_name, thresholds, enabled FROM alert_rules ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	items := make([]types.AlertRule, 0)
	for rows.Next() {
		var rule types.AlertRule
		if err := rows.Scan(&rule.ID, &rule.NodeID, &rule.NodeName, &rule.Thresholds, &rule.Enabled); err != nil {
			return nil
		}
		roleRows, err := r.db.Query(r.ctx, `
			SELECT role_id FROM alert_rule_notify_roles WHERE rule_id = $1 ORDER BY role_id
		`, rule.ID)
		if err == nil {
			for roleRows.Next() {
				var roleID string
				if err := roleRows.Scan(&roleID); err == nil {
					rule.NotifyRoleIDs = append(rule.NotifyRoleIDs, roleID)
				}
			}
			roleRows.Close()
		}
		items = append(items, rule)
	}
	return store.CloneAlertRules(items)
}

func (r *pgBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	cloned := store.CloneAlertRules(rules)
	ids := make([]string, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO alert_rules (id, node_id, node_name, thresholds, enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				node_name = EXCLUDED.node_name,
				thresholds = EXCLUDED.thresholds,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, rule.ID, rule.NodeID, rule.NodeName, rule.Thresholds, rule.Enabled); err != nil {
			return fmt.Errorf("upsert alert rule %s: %w", rule.ID, err)
		}
		if _, err := r.db.Exec(r.ctx, `DELETE FROM alert_rule_notify_roles WHERE rule_id = $1`, rule.ID); err != nil {
			return err
		}
		for _, roleID := range rule.NotifyRoleIDs {
			if _, err := r.db.Exec(r.ctx, `
				INSERT INTO alert_rule_notify_roles (rule_id, role_id) VALUES ($1, $2)
			`, rule.ID, roleID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(r.ctx, `DELETE FROM alert_rule_notify_roles`); err != nil {
			return err
		}
		_, err := r.db.Exec(r.ctx, `DELETE FROM alert_rules`)
		return err
	}
	if err := pruneByColumn(r.ctx, r.db, "alert_rule_notify_roles", "rule_id", ids); err != nil {
		return err
	}
	return pruneByID(r.ctx, r.db, "alert_rules", ids)
}

func (r *pgBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	rows, err := r.db.Query(r.ctx, `
		SELECT member_id, personal_quota FROM member_quota_pools
	`)
	if err != nil {
		return map[string]types.MemberQuotaPool{}
	}
	defer rows.Close()
	pools := make(map[string]types.MemberQuotaPool)
	for rows.Next() {
		var memberID string
		var pool types.MemberQuotaPool
		if err := rows.Scan(&memberID, &pool.PersonalQuota); err != nil {
			return map[string]types.MemberQuotaPool{}
		}
		pools[memberID] = pool
	}
	return store.CloneMemberQuotaPools(pools)
}

func (r *pgBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	cloned := store.CloneMemberQuotaPools(pools)
	if _, err := r.db.Exec(r.ctx, `DELETE FROM member_quota_pools`); err != nil {
		return fmt.Errorf("clear member quota pools: %w", err)
	}
	for memberID, pool := range cloned {
		if _, err := r.db.Exec(r.ctx, `
			INSERT INTO member_quota_pools (member_id, personal_quota, updated_at)
			VALUES ($1, $2, NOW())
		`, memberID, pool.PersonalQuota); err != nil {
			return fmt.Errorf("insert member quota pool: %w", err)
		}
	}
	return nil
}
