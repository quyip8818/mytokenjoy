package budget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/internal/store"
)

// RotatePeriod performs the month rotation if needed: archive previous period,
// apply the current period's pre-configured snapshot, and mark rotation complete.
// Unlike MaybeRotatePeriod, it does NOT enqueue a follow-up rebalance (caller is
// expected to be in a rebalance context already). Returns nil if already rotated.
func (s *service) RotatePeriod(ctx context.Context) error {
	_, err := s.doRotate(ctx, false)
	return err
}

// MaybeRotatePeriod checks if a period rotation is needed and executes it lazily.
// Called at the beginning of GetTree to ensure the current period's snapshot is applied.
// Returns true if rotation was performed.
func (s *service) MaybeRotatePeriod(ctx context.Context) bool {
	rotated, _ := s.doRotate(ctx, true)
	return rotated
}

// doRotate is the shared rotation logic. If enqueueRebalance is true, a follow-up
// company rebalance is enqueued on success (for the lazy/GetTree path).
func (s *service) doRotate(ctx context.Context, enqueueRebalance bool) (bool, error) {
	currentPeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, s.clk).String()
	companyID := store.CompanyID(ctx)

	state, err := s.store.TenantBackgroundState().Get(ctx, companyID)
	if err != nil {
		s.logger.Warn("rotation: failed to get tenant state", "error", err)
		return false, fmt.Errorf("rotation: get tenant state: %w", err)
	}
	lastRotated := ""
	if state != nil {
		lastRotated = state.LastRebalancedPeriod
	}
	if lastRotated >= currentPeriod {
		return false, nil
	}

	if err := s.store.WithTx(ctx, func(tx store.Store) error {
		return s.rotatePeriod(ctx, tx, currentPeriod, lastRotated)
	}); err != nil {
		s.logger.Warn("rotation: failed", "period", currentPeriod, "error", err)
		return false, fmt.Errorf("rotation: %w", err)
	}

	if enqueueRebalance {
		s.enqueueCompanyRebalance(ctx, "budget.rotation")
	}
	s.logger.Info("budget period rotated", "period", currentPeriod, "company", companyID)
	return true, nil
}

func (s *service) rotatePeriod(ctx context.Context, tx store.Store, currentPeriod, lastRotated string) error {
	// Step 1: Snapshot the previous period (the one that just ended) if not already snapshotted.
	if lastRotated != "" {
		exists, err := tx.BudgetSnapshot().Exists(ctx, lastRotated)
		if err != nil {
			return fmt.Errorf("check previous snapshot: %w", err)
		}
		if !exists {
			// Save the current state as the previous period's terminal snapshot
			payload, err := s.buildSnapshotFromTx(ctx, tx, lastRotated)
			if err != nil {
				return fmt.Errorf("build previous snapshot: %w", err)
			}
			data, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal previous snapshot: %w", err)
			}
			if err := tx.BudgetSnapshot().Upsert(ctx, lastRotated, data); err != nil {
				return fmt.Errorf("save previous snapshot: %w", err)
			}
		}
	}

	// Step 2: Apply the current period's pre-configured snapshot (if any).
	snap, err := tx.BudgetSnapshot().Get(ctx, currentPeriod)
	if err != nil {
		return fmt.Errorf("get current period snapshot: %w", err)
	}
	if snap != nil {
		var payload SnapshotPayload
		if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
			return fmt.Errorf("unmarshal current snapshot: %w", err)
		}
		if err := s.applySnapshot(ctx, tx, &payload); err != nil {
			return fmt.Errorf("apply snapshot: %w", err)
		}
	}
	// If no snapshot for current period, do nothing (budget carries over).

	// Step 3: Mark rotation complete.
	if err := tx.TenantBackgroundState().SetLastRebalancedPeriod(ctx, store.CompanyID(ctx), currentPeriod); err != nil {
		return fmt.Errorf("set last rebalanced period: %w", err)
	}
	return nil
}

// applySnapshot writes the snapshot's budget values back into org_nodes/members/projects.
// Only applies to entities that still exist in the system.
func (s *service) applySnapshot(ctx context.Context, tx store.Store, payload *SnapshotPayload) error {
	// Apply department budgets
	nodes, err := tx.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	existingNodeIDs := collectNodeIDs(nodes)
	budgetRepo := tx.Budget().OrgNodeBudget()
	for _, node := range flattenBudgetTree(payload.Tree) {
		if !existingNodeIDs[node.ID] {
			continue // entity deleted since snapshot was taken
		}
		if err := pkgbudget.PersistNodeBudget(ctx, budgetRepo, node.ID, node); err != nil {
			return fmt.Errorf("apply node %s budget: %w", node.ID, err)
		}
	}

	// Apply member budgets
	members, err := tx.Org().Members(ctx)
	if err != nil {
		return err
	}
	memberMap := make(map[string]int, len(members))
	for i, m := range members {
		memberMap[m.ID.String()] = i
	}
	updated := false
	for _, sm := range payload.Members {
		idx, ok := memberMap[sm.MemberID]
		if !ok {
			continue // member deleted since snapshot
		}
		if members[idx].PersonalBudget != sm.PersonalBudget {
			members[idx].PersonalBudget = sm.PersonalBudget
			updated = true
		}
	}
	if updated {
		if err := tx.Org().SetMembers(ctx, members); err != nil {
			return fmt.Errorf("apply member budgets: %w", err)
		}
	}

	// Apply project budgets
	projects, err := tx.Budget().Projects(ctx)
	if err != nil {
		return err
	}
	projectMap := make(map[string]int, len(projects))
	for i, p := range projects {
		projectMap[p.ID.String()] = i
	}
	projectUpdated := false
	for _, sp := range payload.Projects {
		idx, ok := projectMap[sp.ID.String()]
		if !ok {
			continue // project deleted since snapshot
		}
		if projects[idx].Budget != sp.Budget {
			projects[idx].Budget = sp.Budget
			projectUpdated = true
		}
	}
	if projectUpdated {
		if err := tx.Budget().SetProjects(ctx, projects); err != nil {
			return fmt.Errorf("apply project budgets: %w", err)
		}
	}

	return nil
}

// buildSnapshotFromTx builds a snapshot using a transaction's store (for consistency).
// periodKey is used to enrich tree nodes with consumed data from ledger.
func (s *service) buildSnapshotFromTx(ctx context.Context, tx store.Store, periodKey string) (*SnapshotPayload, error) {
	tree, err := common.LoadBudgetTree(ctx, tx.Org().Nodes())
	if err != nil {
		return nil, err
	}
	if periodKey != "" {
		if err := enrichTreeConsumed(ctx, tx.Ledger(), tree, periodKey); err != nil {
			s.logger.Warn("archive enrich consumed failed", "period", periodKey, "err", err)
		}
	}
	members, err := tx.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := tx.Budget().Projects(ctx)
	if err != nil {
		return nil, err
	}
	snapshotMembers := make([]SnapshotMember, 0, len(members))
	for _, m := range members {
		snapshotMembers = append(snapshotMembers, SnapshotMember{
			MemberID:       m.ID.String(),
			DepartmentID:   m.DepartmentID.String(),
			Name:           m.Alias,
			PersonalBudget: m.PersonalBudget,
		})
	}
	return &SnapshotPayload{
		Tree:     tree,
		Members:  snapshotMembers,
		Projects: projects,
	}, nil
}

// --- helpers ---

func collectNodeIDs(nodes []types.OrgNode) map[uuid.UUID]bool {
	ids := make(map[uuid.UUID]bool)
	var walk func([]types.OrgNode)
	walk = func(list []types.OrgNode) {
		for _, n := range list {
			ids[n.ID] = true
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(nodes)
	return ids
}

func flattenBudgetTree(nodes []types.BudgetNode) []types.BudgetNode {
	var result []types.BudgetNode
	var walk func([]types.BudgetNode)
	walk = func(list []types.BudgetNode) {
		for _, n := range list {
			result = append(result, n)
			if n.Children != nil {
				walk(n.Children)
			}
		}
	}
	walk(nodes)
	return result
}
