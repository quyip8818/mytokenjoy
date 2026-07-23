package budget

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/approval"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// MemberBudgetApprovalHandler handles type="member_budget" approvals
// (department reserved pool → member personal budget).
type MemberBudgetApprovalHandler struct {
	svc *service
}

func NewMemberBudgetApprovalHandler(svc Service) *MemberBudgetApprovalHandler {
	return &MemberBudgetApprovalHandler{svc: svc.(*service)}
}

func (h *MemberBudgetApprovalHandler) Type() types.ApprovalType {
	return types.ApprovalTypeMemberBudget
}

func (h *MemberBudgetApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
	var meta types.MemberBudgetApprovalMeta
	if err := json.Unmarshal(input.Metadata, &meta); err != nil {
		return domain.Validation("invalid member_budget approval metadata")
	}
	if meta.Amount <= 0 {
		return domain.Validation("amount must be positive")
	}
	if meta.Reason == "" {
		return domain.Validation("reason required")
	}
	return nil
}

func (h *MemberBudgetApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
	// Quick check reserved pool (no lock, may be stale — real check in OnApprovedTx)
	var meta types.MemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	deptID := h.resolveDeptID(ctx, req)
	if deptID == uuid.Nil {
		return domain.Validation("applicant department not found")
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
		return domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %d quota", reserved))
	}
	return nil
}

func (h *MemberBudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
	var meta types.MemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	// Acquire budget lock to prevent concurrent deduction
	if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
		return nil, err
	}

	deptID := h.resolveDeptID(ctx, req)
	if deptID == uuid.Nil {
		member, err := tx.Org().MemberByID(ctx, req.ApplicantID)
		if err != nil {
			return nil, err
		}
		if member == nil {
			return nil, domain.NotFound("申请人不存在")
		}
		deptID = member.DepartmentID
	}

	row, found, err := tx.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.NotFound("部门不存在")
	}
	reserved := int64(0)
	if row.ReservedPool != nil {
		reserved = *row.ReservedPool
	}
	if reserved < meta.Amount {
		return nil, domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %d quota", reserved))
	}

	// Deduct reserved pool
	newReserved := reserved - meta.Amount
	row.ReservedPool = &newReserved
	if err := tx.Budget().OrgNodeBudget().Upsert(ctx, deptID, row); err != nil {
		return nil, fmt.Errorf("persist reserved pool: %w", err)
	}

	// Increase personal budget
	members, err := tx.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	members = pkgbudget.AddMemberPersonalBudget(members, req.ApplicantID, meta.Amount)
	if err := tx.Org().SetMembers(ctx, members); err != nil {
		return nil, fmt.Errorf("persist member personal budget: %w", err)
	}

	return nil, nil
}

func (h *MemberBudgetApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	// Enqueue idempotent rebalance job
	return h.svc.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, req.ApplicantID)
}

func (h *MemberBudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	// Rebalance is idempotent background task — failure doesn't need compensation
	return nil
}

func (h *MemberBudgetApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
	return nil
}

func (h *MemberBudgetApprovalHandler) PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error) {
	var meta types.MemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	deptID := h.resolveDeptID(ctx, req)
	if deptID == uuid.Nil {
		return json.Marshal(map[string]any{"sufficient": false, "reservedPool": 0, "requested": meta.Amount})
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

func (h *MemberBudgetApprovalHandler) resolveDeptID(ctx context.Context, req types.ApprovalRequest) uuid.UUID {
	var meta types.MemberBudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	if meta.DepartmentID != uuid.Nil {
		return meta.DepartmentID
	}
	member, err := h.svc.store.Org().MemberByID(ctx, req.ApplicantID)
	if err != nil || member == nil {
		return uuid.Nil
	}
	return member.DepartmentID
}
