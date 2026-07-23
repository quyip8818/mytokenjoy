package budget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/approval"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// ProjectMemberBudgetApprovalHandler handles type="project_member_budget" approvals.
// Flow: project member applies → project owner approves → project unallocated budget → member_budget.
type ProjectMemberBudgetApprovalHandler struct {
	svc *service
}

func NewProjectMemberBudgetApprovalHandler(svc Service) *ProjectMemberBudgetApprovalHandler {
	return &ProjectMemberBudgetApprovalHandler{svc: svc.(*service)}
}

func (h *ProjectMemberBudgetApprovalHandler) Type() types.ApprovalType {
	return types.ApprovalTypeProjectMemberBudget
}

func (h *ProjectMemberBudgetApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
	var meta types.ProjectMemberBudgetApprovalMeta
	if err := json.Unmarshal(input.Metadata, &meta); err != nil {
		return domain.Validation("invalid project_member_budget approval metadata")
	}
	if meta.Amount <= 0 {
		return domain.Validation("amount must be positive")
	}
	if meta.Reason == "" {
		return domain.Validation("reason required")
	}

	// Applicant must be a project member
	projects, err := h.svc.store.Budget().Projects(ctx)
	if err != nil {
		return err
	}
	for _, p := range projects {
		if p.ID == meta.ProjectID {
			for _, mid := range p.MemberIDs {
				if mid == input.ApplicantID {
					return nil
				}
			}
			return domain.Validation("applicant is not a member of this project")
		}
	}
	return domain.NotFound("project not found")
}

func (h *ProjectMemberBudgetApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
	var meta types.ProjectMemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	available, err := h.projectUnallocated(ctx, meta.ProjectID)
	if err != nil {
		return err
	}
	if available < meta.Amount {
		return domain.Validation(fmt.Sprintf("项目未分配余额不足，当前剩余 %d quota", available))
	}
	return nil
}

func (h *ProjectMemberBudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
	var meta types.ProjectMemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
		return nil, err
	}

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

	// Check unallocated budget
	allocated := int64(0)
	for _, b := range projects[idx].MemberBudgets {
		allocated += b
	}
	available := projects[idx].Budget - allocated
	if available < meta.Amount {
		return nil, domain.Validation(fmt.Sprintf("项目未分配余额不足，当前剩余 %d quota", available))
	}

	// Increase member budget
	if projects[idx].MemberBudgets == nil {
		projects[idx].MemberBudgets = make(map[uuid.UUID]int64)
	}
	projects[idx].MemberBudgets[req.ApplicantID] += meta.Amount

	if err := tx.Budget().SetProjects(ctx, projects); err != nil {
		return nil, fmt.Errorf("persist project member budget: %w", err)
	}

	return nil, nil
}

func (h *ProjectMemberBudgetApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	return nil
}

func (h *ProjectMemberBudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	return nil
}

func (h *ProjectMemberBudgetApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
	return nil
}

func (h *ProjectMemberBudgetApprovalHandler) PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error) {
	var meta types.ProjectMemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	available, err := h.projectUnallocated(ctx, meta.ProjectID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"sufficient":  available >= meta.Amount,
		"unallocated": available,
		"requested":   meta.Amount,
	})
}

func (h *ProjectMemberBudgetApprovalHandler) projectUnallocated(ctx context.Context, projectID uuid.UUID) (int64, error) {
	projects, err := h.svc.store.Budget().Projects(ctx)
	if err != nil {
		return 0, err
	}
	for _, p := range projects {
		if p.ID == projectID {
			allocated := int64(0)
			for _, b := range p.MemberBudgets {
				allocated += b
			}
			return p.Budget - allocated, nil
		}
	}
	return 0, domain.NotFound("project not found")
}
