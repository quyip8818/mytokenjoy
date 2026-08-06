package budget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// UpdateSnapshotNode modifies a department's budget/reservedPool in a future period snapshot.
func (s *service) UpdateSnapshotNode(ctx context.Context, period string, nodeID uuid.UUID, budget float64, reservedPool *float64) error {
	if err := s.validateFuturePeriod(period); err != nil {
		return err
	}
	if budget < 0 {
		return domain.Validation("budget must be non-negative")
	}

	return s.store.WithTx(ctx, func(tx store.Store) error {
		payload, err := s.loadSnapshotForUpdate(ctx, tx, period)
		if err != nil {
			return err
		}

		found := updateNodeInTree(payload.Tree, nodeID, budget, reservedPool)
		if !found {
			return domain.NotFound("Node not found in snapshot")
		}

		if err := s.saveSnapshot(ctx, tx, period, payload); err != nil {
			return err
		}
		s.appendBudgetLog(ctx, "budget.snapshot.dept.update", nodeID.String(),
			fmt.Sprintf(`{"period":%q,"budget":%.2f}`, period, budget))
		return nil
	})
}

// UpdateSnapshotMember modifies a member's personalBudget in a future period snapshot.
func (s *service) UpdateSnapshotMember(ctx context.Context, period string, memberID uuid.UUID, personalBudget float64) error {
	if err := s.validateFuturePeriod(period); err != nil {
		return err
	}
	if personalBudget < 0 {
		return domain.Validation("personalBudget must be non-negative")
	}

	return s.store.WithTx(ctx, func(tx store.Store) error {
		payload, err := s.loadSnapshotForUpdate(ctx, tx, period)
		if err != nil {
			return err
		}

		found := false
		for i := range payload.Members {
			if payload.Members[i].MemberID == memberID.String() {
				payload.Members[i].PersonalBudget = personalBudget
				found = true
				break
			}
		}
		if !found {
			return domain.NotFound("Member not found in snapshot")
		}

		if err := s.saveSnapshot(ctx, tx, period, payload); err != nil {
			return err
		}
		s.appendBudgetLog(ctx, "budget.snapshot.member.update", memberID.String(),
			fmt.Sprintf(`{"period":%q,"personalBudget":%.2f}`, period, personalBudget))
		return nil
	})
}

// UpdateSnapshotProject modifies a project's budget in a future period snapshot.
func (s *service) UpdateSnapshotProject(ctx context.Context, period string, projectID uuid.UUID, budget float64) error {
	if err := s.validateFuturePeriod(period); err != nil {
		return err
	}
	if budget < 0 {
		return domain.Validation("budget must be non-negative")
	}

	return s.store.WithTx(ctx, func(tx store.Store) error {
		payload, err := s.loadSnapshotForUpdate(ctx, tx, period)
		if err != nil {
			return err
		}

		found := false
		for i := range payload.Projects {
			if payload.Projects[i].ID == projectID {
				payload.Projects[i].Budget = budget
				found = true
				break
			}
		}
		if !found {
			return domain.NotFound("Project not found in snapshot")
		}

		if err := s.saveSnapshot(ctx, tx, period, payload); err != nil {
			return err
		}
		s.appendBudgetLog(ctx, "budget.snapshot.project.update", projectID.String(),
			fmt.Sprintf(`{"period":%q,"budget":%.2f}`, period, budget))
		return nil
	})
}

// --- helpers ---

func (s *service) validateFuturePeriod(period string) error {
	if !periodKeyPattern.MatchString(period) {
		return domain.BadRequest("period must be YYYY-MM format")
	}
	currentPeriod := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, s.clk).String()
	if period <= currentPeriod {
		return domain.BadRequest("can only modify a future period snapshot")
	}
	return nil
}

func (s *service) loadSnapshotForUpdate(ctx context.Context, tx store.Store, period string) (*SnapshotPayload, error) {
	snap, err := tx.BudgetSnapshot().GetForUpdate(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("get snapshot for update: %w", err)
	}
	if snap == nil {
		return nil, domain.NotFound("No snapshot for this period. Use 'Copy' first to create one.")
	}
	var payload SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &payload, nil
}

func (s *service) saveSnapshot(ctx context.Context, tx store.Store, period string, payload *SnapshotPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	return tx.BudgetSnapshot().Upsert(ctx, period, data)
}

// updateNodeInTree recursively finds and updates a node's budget in the tree.
func updateNodeInTree(nodes []types.BudgetNode, nodeID uuid.UUID, budget float64, reservedPool *float64) bool {
	for i := range nodes {
		if nodes[i].ID == nodeID {
			nodes[i].Budget = budget
			if reservedPool != nil {
				nodes[i].ReservedPool = reservedPool
			}
			return true
		}
		if nodes[i].Children != nil {
			if updateNodeInTree(nodes[i].Children, nodeID, budget, reservedPool) {
				return true
			}
		}
	}
	return false
}
