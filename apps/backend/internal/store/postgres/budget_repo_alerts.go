package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) OverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	companyID := store.CompanyID(ctx)
	var policy types.OverrunPolicyConfig
	err := r.db.QueryRow(ctx, `
		SELECT thresholds, notify_email, notify_phone, notify_im, block_message
		FROM overrun_policy WHERE company_id = $1
	`, companyID).Scan(&policy.Thresholds, &policy.NotifyEmail, &policy.NotifyPhone, &policy.NotifyIm, &policy.BlockMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return types.OverrunPolicyConfig{}, nil
		}
		return types.OverrunPolicyConfig{}, err
	}
	return policy, nil
}

func (r *pgBudgetRepo) SetOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) error {
	companyID := store.CompanyID(ctx)
	_, err := r.db.Exec(ctx, `
		INSERT INTO overrun_policy (company_id, thresholds, notify_email, notify_phone, notify_im, block_message, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			thresholds = EXCLUDED.thresholds,
			notify_email = EXCLUDED.notify_email,
			notify_phone = EXCLUDED.notify_phone,
			notify_im = EXCLUDED.notify_im,
			block_message = EXCLUDED.block_message,
			updated_at = NOW()
	`, companyID, policy.Thresholds, policy.NotifyEmail, policy.NotifyPhone, policy.NotifyIm, policy.BlockMessage)
	if err != nil {
		return fmt.Errorf("upsert overrun policy: %w", err)
	}
	return nil
}

func (r *pgBudgetRepo) AlertRules(ctx context.Context) ([]types.AlertRule, error) {
	companyID := store.CompanyID(ctx)

	rows, err := r.db.Query(ctx, `
		SELECT ar.id, ar.node_id, n.name, ar.thresholds, ar.enabled
		FROM alert_rules ar
		JOIN org_nodes n ON n.company_id = ar.company_id AND n.id = ar.node_id
		WHERE ar.company_id = $1
		ORDER BY ar.id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := make([]types.AlertRule, 0)
	ruleIndex := make(map[uuid.UUID]int)
	for rows.Next() {
		var rule types.AlertRule
		if err := rows.Scan(&rule.ID, &rule.NodeID, &rule.NodeName, &rule.Thresholds, &rule.Enabled); err != nil {
			return nil, err
		}
		rule.NotifyRoleIDs = []uuid.UUID{}
		ruleIndex[rule.ID] = len(rules)
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return rules, nil
	}

	roleRows, err := r.db.Query(ctx, `SELECT rule_id, role_id FROM alert_rule_notify_roles WHERE company_id = $1 ORDER BY rule_id, role_id`, companyID)
	if err != nil {
		return nil, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var ruleID, roleID uuid.UUID
		if err := roleRows.Scan(&ruleID, &roleID); err != nil {
			return nil, err
		}
		if idx, ok := ruleIndex[ruleID]; ok {
			rules[idx].NotifyRoleIDs = append(rules[idx].NotifyRoleIDs, roleID)
		}
	}
	if err := roleRows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *pgBudgetRepo) SetAlertRules(ctx context.Context, rules []types.AlertRule) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneAlertRules(rules)
	ids := make([]uuid.UUID, len(cloned))
	for i, rule := range cloned {
		ids[i] = rule.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO alert_rules (id, company_id, node_id, thresholds, enabled, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				node_id = EXCLUDED.node_id,
				thresholds = EXCLUDED.thresholds,
				enabled = EXCLUDED.enabled,
				updated_at = NOW()
		`, rule.ID, companyID, rule.NodeID, rule.Thresholds, rule.Enabled); err != nil {
			return fmt.Errorf("upsert alert rule %s: %w", rule.ID, err)
		}
		if _, err := r.db.Exec(ctx, `DELETE FROM alert_rule_notify_roles WHERE company_id = $1 AND rule_id = $2`, companyID, rule.ID); err != nil {
			return err
		}
		for _, roleID := range rule.NotifyRoleIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO alert_rule_notify_roles (company_id, rule_id, role_id) VALUES ($1, $2, $3)
			`, companyID, rule.ID, roleID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM alert_rule_notify_roles WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM alert_rules WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "alert_rule_notify_roles", "rule_id", companyID, ids); err != nil {
		return err
	}
	return pruneByIDForCompany(ctx, r.db, "alert_rules", companyID, ids)
}
