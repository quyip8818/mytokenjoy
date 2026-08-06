package budget

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

// periodKeyPattern validates "YYYY-MM" format.
var periodKeyPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)

// SnapshotPayload is the JSON structure stored in budget_snapshot.snapshot.
// Mirrors the frontend's expected format.
type SnapshotPayload struct {
	Tree     []types.BudgetNode `json:"tree"`
	Members  []SnapshotMember   `json:"members"`
	Projects []types.Project    `json:"projects"`
}

// SnapshotMember is a minimal member record for the snapshot.
type SnapshotMember struct {
	MemberID       string  `json:"memberId"`
	DepartmentID   string  `json:"departmentId"`
	Name           string  `json:"name"`
	PersonalBudget float64 `json:"personalBudget"`
}

// CopyPeriod snapshots the current budget state into the target period.
func (s *service) CopyPeriod(ctx context.Context, toPeriod string) error {
	if !periodKeyPattern.MatchString(toPeriod) {
		return domain.BadRequest("period must be YYYY-MM format")
	}
	currentPeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, clock.NowUTC(s.clk))
	if toPeriod <= currentPeriod {
		return domain.BadRequest("can only pre-configure a future period")
	}

	// ponytail: idempotent — if snapshot already exists (user already edited), skip.
	exists, err := s.store.BudgetSnapshot().Exists(ctx, toPeriod)
	if err != nil {
		return fmt.Errorf("check snapshot exists: %w", err)
	}
	if exists {
		return nil
	}

	payload, err := s.buildCurrentSnapshot(ctx)
	if err != nil {
		return fmt.Errorf("build snapshot: %w", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := s.store.BudgetSnapshot().Upsert(ctx, toPeriod, data); err != nil {
		return fmt.Errorf("upsert snapshot: %w", err)
	}
	s.appendBudgetLog(ctx, "budget.period.copy", toPeriod,
		fmt.Sprintf(`{"to":%q}`, toPeriod))
	return nil
}

// GetTreeForPeriod returns the budget tree for a given period.
// For the current month it returns live data (same as GetTree); for other months it reads from snapshot.
// If no snapshot exists for a non-past period, falls back to live data (budget config carries over).
func (s *service) GetTreeForPeriod(ctx context.Context, period string) ([]types.BudgetNode, error) {
	if !periodKeyPattern.MatchString(period) {
		return nil, domain.BadRequest("period must be YYYY-MM format")
	}
	currentPeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, clock.NowUTC(s.clk))
	// Current or future period without snapshot → live data
	if period >= currentPeriod {
		snap, err := s.store.BudgetSnapshot().Get(ctx, period)
		if err != nil {
			return nil, fmt.Errorf("get snapshot: %w", err)
		}
		if snap != nil {
			var payload SnapshotPayload
			if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
				return nil, fmt.Errorf("unmarshal snapshot: %w", err)
			}
			return payload.Tree, nil
		}
		// No snapshot for current/future → return live tree
		return s.GetTree(ctx)
	}
	// Past period → must have snapshot
	snap, err := s.store.BudgetSnapshot().Get(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("get snapshot: %w", err)
	}
	if snap == nil {
		return nil, nil // no historical snapshot → frontend shows empty state
	}
	var payload SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return payload.Tree, nil
}

// buildCurrentSnapshot reads current org_nodes, members, and projects to produce a full snapshot.
func (s *service) buildCurrentSnapshot(ctx context.Context) (*SnapshotPayload, error) {
	tree, err := common.LoadBudgetTree(ctx, s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	projects, err := s.store.Budget().Projects(ctx)
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
