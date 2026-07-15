package budget

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
)

// checkAlertThresholds evaluates percentage alert rules for departments
// touched in this batch. If consumed/budget crosses a configured threshold,
// a notification is sent. This is best-effort: failures are logged, not propagated.
// Deduplication: each (ruleID, threshold, periodKey) is only notified once per
// Projector lifetime (resets on process restart).
func (p *Projector) checkAlertThresholds(ctx context.Context, effects batchEffects) {
	if p.notifier == nil || len(effects.touchedDepts) == 0 {
		return
	}

	rules, err := p.store.Budget().AlertRules(ctx)
	if err != nil {
		p.logger.Warn("checkAlertThresholds: failed to load alert rules", "error", err)
		return
	}
	if len(rules) == 0 {
		return
	}

	// Index rules by NodeID for quick lookup.
	rulesByNode := make(map[string][]types.AlertRule)
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		rulesByNode[rule.NodeID] = append(rulesByNode[rule.NodeID], rule)
	}

	if p.alertsSent == nil {
		p.alertsSent = make(map[string]struct{})
	}

	for deptID := range effects.touchedDepts {
		deptRules, ok := rulesByNode[deptID]
		if !ok {
			continue
		}

		budget, found, err := p.store.Org().Nodes().GetNodeBudget(ctx, deptID)
		if err != nil || !found || budget <= 0 {
			continue
		}

		open, err := pkgbudget.OpenDepartmentPeriod(ctx, p.store.Org().Nodes(), deptID, p.cfg.Clock())
		if err != nil {
			continue
		}
		periodKey := open.String()

		consumed, err := p.store.Ledger().SumAmountByDepartment(ctx, deptID, periodKey)
		if err != nil {
			continue
		}

		pct := int(consumed * 100 / budget)

		for _, rule := range deptRules {
			// Thresholds are sorted ascending. Check from highest to lowest
			// so we only fire the highest crossed threshold per rule.
			for i := len(rule.Thresholds) - 1; i >= 0; i-- {
				threshold := rule.Thresholds[i]
				if pct >= threshold {
					dedupKey := rule.ID + ":" + fmt.Sprintf("%d", threshold) + ":" + periodKey
					if _, sent := p.alertsSent[dedupKey]; sent {
						break
					}
					p.alertsSent[dedupKey] = struct{}{}
					_ = p.notifier.Send(ctx, types.Notification{
						EventType: types.NotificationEventBudgetAlertReached,
						Recipient: fmt.Sprintf("department:%s", deptID),
						Payload: map[string]any{
							"departmentId": deptID,
							"nodeName":     rule.NodeName,
							"threshold":    threshold,
							"currentPct":   pct,
							"consumed":     consumed,
							"budget":       budget,
							"periodKey":    periodKey,
							"ruleId":       rule.ID,
						},
					})
					break
				}
			}
		}
	}
}
