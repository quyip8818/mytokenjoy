package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error) {
	if err := s.delayer.Wait(ctx, 400*time.Millisecond); err != nil {
		return types.KeyApproval{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.KeyApproval{}, err
	}
	member, ok := org.FindMemberByID(members, input.MemberID)
	if !ok {
		return types.KeyApproval{}, domain.NotFound("member not found")
	}
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.KeyApproval{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return types.KeyApproval{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.KeyApproval{}, err
	}
	if msg := common.ValidateModelIDsForMember(input.MemberID, input.RequestedModels, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
		return types.KeyApproval{}, domain.Validation(*msg)
	}
	created := types.KeyApproval{
		ID:   fmt.Sprintf("apv-%d", time.Now().UnixMilli()),
		Type: input.Type, Applicant: member.Name, ApplicantID: input.MemberID, Department: member.DepartmentName,
		Reason: input.Reason, RequestedBudget: input.RequestedBudget,
		RequestedModels: append([]int64{}, input.RequestedModels...),
		Status:          "pending", CreatedAt: time.Now().Format("2006-01-02 15:04"),
	}
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return types.KeyApproval{}, err
	}
	approvals = append(approvals, created)
	if err := s.store.Keys().SetApprovals(ctx, approvals); err != nil {
		return types.KeyApproval{}, err
	}
	return created, nil
}

func (s *service) ApproveApproval(ctx context.Context, id string, approverMemberID string) error {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return err
	}
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return err
	}
	idx := -1
	for i := range approvals {
		if approvals[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.NotFound("Not found")
	}
	approval := approvals[idx]
	if approval.Type == "key" {
		if err := s.requireNewAPI(); err != nil {
			return err
		}
	}
	tree, err := budget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), s.cfg.Clock())
	if err != nil {
		return err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	reservedPool := budget.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	if approval.Type == "budget" && approval.RequestedBudget > reservedPool {
		return domain.Validation("Reserved pool insufficient")
	}

	var createdKeyID string
	var personalBudgetAdded float64
	if err := s.store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		platformKeys, err := budget.LoadPlatformKeysWithUsed(ctx, st.BudgetSnapshots(), st.Org(), st.Budget(), st.Keys(), s.cfg.Clock())
		if err != nil {
			return err
		}
		if approval.Type == "key" {
			keyQuota := approval.RequestedBudget
			remaining := budget.GetBudgetRemaining(members, platformKeys, approval.ApplicantID)
			if keyQuota > remaining {
				personalBudgetAdded = keyQuota - remaining
				members = budget.AddMemberPersonalBudget(members, approval.ApplicantID, personalBudgetAdded)
			}
			memberID := approval.ApplicantID
			createdKeyID = fmt.Sprintf("plk-apv-%d", time.Now().UnixMilli())
			platformKeys = append(platformKeys, types.PlatformKey{
				ID:   createdKeyID,
				Name: fmt.Sprintf("%s-审批 Key", approval.Applicant), KeyPrefix: "pending...",
				MemberID: &memberID, Status: "active", Budget: keyQuota, Used: 0,
				ModelWhitelist: append([]int64{}, approval.RequestedModels...),
				CreatedAt:      time.Now().Format("2006-01-02"),
			})
			if err := st.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return err
			}
		} else if approval.Type == "budget" {
			members = budget.AddMemberPersonalBudget(members, approval.ApplicantID, approval.RequestedBudget)
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}

		approver, err := resolveMemberName(approverMemberID, members)
		if err != nil {
			return err
		}
		now := time.Now().Format("2006-01-02 15:04")
		approvals[idx].Status = "approved"
		approvals[idx].Approver = &approver
		approvals[idx].ResolvedAt = &now
		return st.Keys().SetApprovals(ctx, approvals)
	}); err != nil {
		return err
	}

	if createdKeyID == "" {
		return nil
	}
	applicant, ok := org.FindMemberByID(members, approval.ApplicantID)
	if !ok || applicant.DepartmentID == "" {
		return domain.Validation("department required for newapi sync")
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	var created types.PlatformKey
	for _, key := range platformKeys {
		if key.ID == createdKeyID {
			created = key
			break
		}
	}
	if created.ID == "" {
		return domain.NotFound("Not found")
	}
	_, err = s.syncPlatformKeyCreate(ctx, created, applicant.DepartmentID)
	if err != nil {
		if compErr := s.compensateFailedKeyApproval(ctx, id, approval.ApplicantID, personalBudgetAdded); compErr != nil {
			return compErr
		}
	}
	return err
}

func (s *service) revertKeyApproval(ctx context.Context, approvalID string) error {
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return err
	}
	for i := range approvals {
		if approvals[i].ID != approvalID {
			continue
		}
		approvals[i].Status = "pending"
		approvals[i].Approver = nil
		approvals[i].ResolvedAt = nil
		return s.store.Keys().SetApprovals(ctx, approvals)
	}
	return domain.NotFound("Not found")
}

func (s *service) compensateFailedKeyApproval(ctx context.Context, approvalID, applicantID string, personalBudgetAdded float64) error {
	if err := s.revertKeyApproval(ctx, approvalID); err != nil {
		return err
	}
	if personalBudgetAdded <= 0 {
		return nil
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	members = budget.AddMemberPersonalBudget(members, applicantID, -personalBudgetAdded)
	return s.store.Org().SetMembers(ctx, members)
}

func (s *service) RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return err
	}
	approvals, err := s.store.Keys().Approvals(ctx)
	if err != nil {
		return err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	for i := range approvals {
		if approvals[i].ID == id {
			approver, err := resolveMemberName(approverMemberID, members)
			if err != nil {
				return err
			}
			now := time.Now().Format("2006-01-02 15:04")
			approvals[i].Status = "rejected"
			approvals[i].Approver = &approver
			approvals[i].RejectReason = reason
			approvals[i].ResolvedAt = &now
			return s.store.Keys().SetApprovals(ctx, approvals)
		}
	}
	return domain.NotFound("Not found")
}
