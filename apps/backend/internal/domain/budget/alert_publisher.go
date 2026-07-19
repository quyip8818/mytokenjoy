package budget

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// BudgetAlertEvent is a single alert to be published via the notification server.
type BudgetAlertEvent struct {
	CompanyID    uuid.UUID
	RecipientID  uuid.UUID
	DepartmentID uuid.UUID
	NodeName     string
	RuleID       uuid.UUID
	Threshold    int
	CurrentPct   int
	Consumed     float64
	Budget       float64
	PeriodKey    string
	DedupeKey    string // budget-alert:{companyID}:{ruleID}:{threshold}:{periodKey}:{memberID}
}

// AlertPublisher is the domain port for async budget alert delivery.
// The app layer adapts this to notification.Service.DispatchAsync.
type AlertPublisher interface {
	PublishBudgetAlerts(ctx context.Context, alerts []BudgetAlertEvent) error
}

// NoopAlertPublisher discards alerts (used when notification is not wired).
type noopAlertPublisher struct{}

func (noopAlertPublisher) PublishBudgetAlerts(context.Context, []BudgetAlertEvent) error { return nil }

var NoopAlertPublisher AlertPublisher = noopAlertPublisher{}

// CheckBudgetAlerts evaluates percentage alert rules for the given departments
// and publishes crossed-threshold alerts via the AlertPublisher.
// This is best-effort: errors are logged, not propagated.
func CheckBudgetAlerts(
	ctx context.Context,
	st store.Store,
	companyID uuid.UUID,
	touchedDepts map[uuid.UUID]struct{},
	publisher AlertPublisher,
	logger *slog.Logger,
) {
	if publisher == nil || len(touchedDepts) == 0 {
		return
	}
	checkBudgetAlertsImpl(ctx, st, companyID, touchedDepts, publisher, logger)
}

func checkBudgetAlertsImpl(
	ctx context.Context,
	st store.Store,
	companyID uuid.UUID,
	touchedDepts map[uuid.UUID]struct{},
	publisher AlertPublisher,
	logger *slog.Logger,
) {
	rules, err := st.Budget().AlertRules(ctx)
	if err != nil {
		if logger != nil {
			logger.Warn("checkBudgetAlerts: failed to load rules", "error", err)
		}
		return
	}
	if len(rules) == 0 {
		return
	}

	// Index rules by NodeID.
	rulesByNode := make(map[uuid.UUID][]types.AlertRule)
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		rulesByNode[rule.NodeID] = append(rulesByNode[rule.NodeID], rule)
	}

	// Resolve members for roles (needed for real recipient IDs).
	// AlertRule.NotifyRoleIDs stores role IDs; Member.Roles stores role names.
	// We need the role ID → name mapping to bridge.
	roles, err := st.Org().Roles(ctx)
	if err != nil {
		if logger != nil {
			logger.Warn("checkBudgetAlerts: failed to load roles", "error", err)
		}
		return
	}
	roleNameByID := make(map[uuid.UUID]string, len(roles))
	for _, r := range roles {
		roleNameByID[r.ID] = r.Name
	}

	members, err := st.Org().Members(ctx)
	if err != nil {
		if logger != nil {
			logger.Warn("checkBudgetAlerts: failed to load members", "error", err)
		}
		return
	}
	membersByRoleName := IndexMembersByRole(members)

	var alerts []BudgetAlertEvent

	for deptID := range touchedDepts {
		deptRules, ok := rulesByNode[deptID]
		if !ok {
			continue
		}

		budget, found, err := st.Org().Nodes().GetNodeBudget(ctx, deptID)
		if err != nil || !found || budget <= 0 {
			continue
		}

		open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), deptID, nil)
		if err != nil {
			continue
		}
		periodKey := open.String()

		consumed, err := st.Ledger().SumAmountByDepartment(ctx, deptID, periodKey)
		if err != nil {
			continue
		}

		pct := int(consumed * 100 / budget)

		for _, rule := range deptRules {
			// Check highest crossed threshold only.
			for i := len(rule.Thresholds) - 1; i >= 0; i-- {
				threshold := rule.Thresholds[i]
				if pct >= threshold {
					// Expand NotifyRoleIDs to real member IDs.
					recipients := ResolveRoleRecipients(rule.NotifyRoleIDs, roleNameByID, membersByRoleName)
					for _, memberID := range recipients {
						alerts = append(alerts, BudgetAlertEvent{
							CompanyID:    companyID,
							RecipientID:  memberID,
							DepartmentID: deptID,
							NodeName:     rule.NodeName,
							RuleID:       rule.ID,
							Threshold:    threshold,
							CurrentPct:   pct,
							Consumed:     float64(consumed),
							Budget:       float64(budget),
							PeriodKey:    periodKey,
							DedupeKey:    fmt.Sprintf("budget-alert:%s:%s:%d:%s:%s", companyID, rule.ID, threshold, periodKey, memberID),
						})
					}
					break
				}
			}
		}
	}

	if len(alerts) == 0 {
		return
	}
	if err := publisher.PublishBudgetAlerts(ctx, alerts); err != nil {
		if logger != nil {
			logger.Warn("checkBudgetAlerts: publish failed", "error", err, "count", len(alerts))
		}
	}
}

func IndexMembersByRole(members []types.Member) map[string][]uuid.UUID {
	out := make(map[string][]uuid.UUID)
	for _, m := range members {
		if m.Status != "active" {
			continue
		}
		for _, role := range m.Roles {
			out[role] = append(out[role], m.ID)
		}
	}
	return out
}

func ResolveRoleRecipients(roleIDs []uuid.UUID, roleNameByID map[uuid.UUID]string, membersByRoleName map[string][]uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	var out []uuid.UUID
	for _, roleID := range roleIDs {
		roleName, ok := roleNameByID[roleID]
		if !ok {
			continue
		}
		for _, memberID := range membersByRoleName[roleName] {
			if _, ok := seen[memberID]; ok {
				continue
			}
			seen[memberID] = struct{}{}
			out = append(out, memberID)
		}
	}
	return out
}
