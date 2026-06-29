package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetlookup"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
	"github.com/tokenjoy/backend/internal/pkg/pkgconst"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error) {
	if err := s.delayer.Wait(ctx, 400*time.Millisecond); err != nil {
		return types.KeyApproval{}, err
	}
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	if msg := routingutil.ValidateModelsForMember(input.MemberID, input.RequestedModels, members, departments, rules, models, pkgconst.ModelNotInDeptMessage); msg != nil {
		return types.KeyApproval{}, domain.Validation(*msg)
	}
	applicant := "申请人"
	department := ""
	if member, ok := queryutil.FindMemberByID(members, input.MemberID); ok {
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
	approvals := s.store.Keys().Approvals()
	approvals = append(approvals, created)
	if err := s.store.Keys().SetApprovals(approvals); err != nil {
		return types.KeyApproval{}, err
	}
	return created, nil
}

func (s *service) ApproveApproval(ctx context.Context, id string, approverMemberID string) error {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return err
	}
	approvals := s.store.Keys().Approvals()
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
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	reservedPool := budgetlookup.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	if approval.Type == "quota" && approval.RequestedQuota > reservedPool {
		return domain.Validation("Reserved pool insufficient")
	}

	return s.store.WithTx(ctx, func(st store.Store) error {
		pools := st.Budget().MemberQuotaPools()
		platformKeys := st.Keys().PlatformKeys()
		if approval.Type == "key" {
			keyQuota := approval.RequestedQuota
			remaining := memberquota.GetQuotaRemaining(pools, platformKeys, approval.ApplicantID)
			if keyQuota > remaining {
				memberquota.AddPersonalQuota(pools, approval.ApplicantID, keyQuota-remaining)
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
			if err := st.Keys().SetPlatformKeys(platformKeys); err != nil {
				return err
			}
		} else if approval.Type == "quota" {
			memberquota.AddPersonalQuota(pools, approval.ApplicantID, approval.RequestedQuota)
		}
		if err := st.Budget().SetMemberQuotaPools(pools); err != nil {
			return err
		}

		approver := sessionutil.ResolveDemoMemberName(approverMemberID, members)
		now := time.Now().Format("2006-01-02 15:04")
		approvals[idx].Status = "approved"
		approvals[idx].Approver = &approver
		approvals[idx].ResolvedAt = &now
		return st.Keys().SetApprovals(approvals)
	})
}

func (s *service) RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return err
	}
	approvals := s.store.Keys().Approvals()
	members := s.store.Org().Members()
	for i := range approvals {
		if approvals[i].ID == id {
			approver := sessionutil.ResolveDemoMemberName(approverMemberID, members)
			now := time.Now().Format("2006-01-02 15:04")
			approvals[i].Status = "rejected"
			approvals[i].Approver = &approver
			approvals[i].RejectReason = reason
			approvals[i].ResolvedAt = &now
			return s.store.Keys().SetApprovals(approvals)
		}
	}
	return nil
}
