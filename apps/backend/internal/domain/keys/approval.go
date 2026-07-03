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
	departments, err := common.LoadDepartments(ctx, s.store)
	if err != nil {
		return types.KeyApproval{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store)
	if err != nil {
		return types.KeyApproval{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.KeyApproval{}, err
	}
	if msg := common.ValidateModelsForMember(input.MemberID, input.RequestedModels, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
		return types.KeyApproval{}, domain.Validation(*msg)
	}
	applicant := "申请人"
	department := ""
	if member, ok := org.FindMemberByID(members, input.MemberID); ok {
		applicant = member.Name
		department = member.DepartmentName
	}
	created := types.KeyApproval{
		ID:   fmt.Sprintf("apv-%d", time.Now().UnixMilli()),
		Type: input.Type, Applicant: applicant, ApplicantID: input.MemberID, Department: department,
		Reason: input.Reason, RequestedQuota: input.RequestedQuota,
		RequestedModels: append([]string{}, input.RequestedModels...),
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
	tree, err := common.LoadBudgetTree(ctx, s.store)
	if err != nil {
		return err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	reservedPool := budget.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	if approval.Type == "quota" && approval.RequestedQuota > reservedPool {
		return domain.Validation("Reserved pool insufficient")
	}

	return s.store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		platformKeys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			return err
		}
		if approval.Type == "key" {
			keyQuota := approval.RequestedQuota
			remaining := budget.GetQuotaRemaining(members, platformKeys, approval.ApplicantID)
			if keyQuota > remaining {
				members = budget.AddMemberPersonalQuota(members, approval.ApplicantID, keyQuota-remaining)
			}
			memberName := approval.Applicant
			memberID := approval.ApplicantID
			fullKey := fmt.Sprintf("tj-apv-%d-demo-secret-key", time.Now().UnixMilli())
			prefix := fullKey
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			platformKeys = append(platformKeys, types.PlatformKey{
				ID:   fmt.Sprintf("plk-apv-%d", time.Now().UnixMilli()),
				Name: fmt.Sprintf("%s-审批 Key", approval.Applicant), KeyPrefix: prefix, FullKey: &fullKey,
				MemberID: &memberID, MemberName: &memberName, Status: "active", Quota: keyQuota, Used: 0,
				ModelWhitelist: append([]string{}, approval.RequestedModels...),
				CreatedAt:      time.Now().Format("2006-01-02"),
			})
			if err := st.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
				return err
			}
		} else if approval.Type == "quota" {
			members = budget.AddMemberPersonalQuota(members, approval.ApplicantID, approval.RequestedQuota)
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}

		approver := common.ResolveDemoMemberName(approverMemberID, members)
		now := time.Now().Format("2006-01-02 15:04")
		approvals[idx].Status = "approved"
		approvals[idx].Approver = &approver
		approvals[idx].ResolvedAt = &now
		return st.Keys().SetApprovals(ctx, approvals)
	})
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
			approver := common.ResolveDemoMemberName(approverMemberID, members)
			now := time.Now().Format("2006-01-02 15:04")
			approvals[i].Status = "rejected"
			approvals[i].Approver = &approver
			approvals[i].RejectReason = reason
			approvals[i].ResolvedAt = &now
			return s.store.Keys().SetApprovals(ctx, approvals)
		}
	}
	return nil
}
