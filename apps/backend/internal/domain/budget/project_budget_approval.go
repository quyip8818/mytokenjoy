package budget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/approval"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// ProjectBudgetApprovalHandler handles type="project_budget" approvals.
// Flow: project owner applies → admin (budget:approve) approves → department reserved pool → project budget.
type ProjectBudgetApprovalHandler struct {
	svc *service
}

func NewProjectBudgetApprovalHandler(svc Service) *ProjectBudgetApprovalHandler {
	return &ProjectBudgetApprovalHandler{svc: svc.(*service)}
}

func (h *ProjectBudgetApprovalHandler) Type() types.ApprovalType {
	return types.ApprovalTypeProjectBudget
}

func (h *ProjectBudgetApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
	var meta types.ProjectBudgetApprovalMeta
	if err := json.Unmarshal(input.Metadata, &meta); err != nil {
		return domain.Validation("invalid project_budget approval metadata")
	}
	if meta.Amount <= 0 {
		return domain.Validation("amount must be positive")
	}
	if meta.Reason == "" {
		return domain.Validation("reason required")
	}

	// Applicant must be project owner
	projects, err := h.svc.store.Budget().Projects(ctx)
	if err != nil {
		return err
	}
	for _, p := range projects {
		if p.ID == meta.ProjectID {
			if p.OwnerID == nil || *p.OwnerID != input.ApplicantID {
				return domain.Validation("only project owner can apply for project budget")
			}
			return nil
		}
	}
	return domain.NotFound("project not found")
}

func (h *ProjectBudgetApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
	var meta types.ProjectBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	// Find project to get ownerDepartmentID
	projects, err := h.svc.store.Budget().Projects(ctx)
	if err != nil {
		return err
	}
	var deptID = req.ScopeID // fallback
	for _, p := range projects {
		if p.ID == meta.ProjectID {
			deptID = p.OwnerDepartmentID
			break
		}
	}

	row, found, err := h.svc.store.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		return err
	}
	if !found {
		return domain.NotFound("department budget not found")
	}
	reserved := int64(0)
	if row.ReservedPool != nil {
		reserved = *row.ReservedPool
	}
	if reserved < meta.Amount {
		return domain.Validation(fmt.Sprintf("部门预留池余额不足，当前剩余 %d quota", reserved))
	}
	return nil
}

func (h *ProjectBudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
	var meta types.ProjectBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
		return nil, err
	}

	// Find project and its department
	projects, err := tx.Budget().Projects(ctx)
	if err != nil {
		return nil, err
	}
	idx := -1
	for i, p := range projects {
		if p.ID == meta.ProjectID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, domain.NotFound("project not found")
	}

	deptID := projects[idx].OwnerDepartmentID

	// Check department reserved pool
	row, found, err := tx.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.NotFound("department budget not found")
	}
	reserved := int64(0)
	if row.ReservedPool != nil {
		reserved = *row.ReservedPool
	}
	if reserved < meta.Amount {
		return nil, domain.Validation(fmt.Sprintf("部门预留池余额不足，当前剩余 %d quota", reserved))
	}

	// Deduct department reserved pool
	newReserved := reserved - meta.Amount
	row.ReservedPool = &newReserved
	if err := tx.Budget().OrgNodeBudget().Upsert(ctx, deptID, row); err != nil {
		return nil, fmt.Errorf("persist reserved pool: %w", err)
	}

	// Increase project budget
	projects[idx].Budget += meta.Amount
	if err := tx.Budget().SetProjects(ctx, projects); err != nil {
		return nil, fmt.Errorf("persist project budget: %w", err)
	}

	return nil, nil
}

func (h *ProjectBudgetApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	return nil
}

func (h *ProjectBudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	return nil
}

func (h *ProjectBudgetApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
	return nil
}

func (h *ProjectBudgetApprovalHandler) PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error) {
	var meta types.ProjectBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	projects, err := h.svc.store.Budget().Projects(ctx)
	if err != nil {
		return nil, err
	}
	var deptID = req.ScopeID
	for _, p := range projects {
		if p.ID == meta.ProjectID {
			deptID = p.OwnerDepartmentID
			break
		}
	}

	row, found, err := h.svc.store.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		return nil, err
	}
	reserved := int64(0)
	if found && row.ReservedPool != nil {
		reserved = *row.ReservedPool
	}
	return json.Marshal(map[string]any{
		"sufficient":   reserved >= meta.Amount,
		"reservedPool": reserved,
		"requested":    meta.Amount,
	})
}
