package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetgroupquota"
	"github.com/tokenjoy/backend/internal/pkg/budgetlookup"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
	"github.com/tokenjoy/backend/internal/pkg/pkgconst"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListProviderKeys() []types.ProviderKey
	CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	ToggleProviderKey(ctx context.Context, id string) error
	RotateProviderKey(ctx context.Context, id string) (types.ProviderKey, error)
	DeleteProviderKey(id string) error
	ListPlatformKeys(memberID, budgetGroupID string) types.PageResult[types.PlatformKey]
	QuotaSummary(memberID string) types.MemberQuotaSummary
	CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error)
	UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error)
	TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error)
	RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error)
	RevokePlatformKey(ctx context.Context, id string) error
	DeletePlatformKey(id string) error
	ListApprovals(tab, memberID string) []types.KeyApproval
	CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error)
	ApprovalQuotaCheck(id string) types.ApprovalQuotaCheck
	ApproveApproval(ctx context.Context, id string, approverMemberID string) error
	RejectApproval(ctx context.Context, id string, approverMemberID string, reason *string) error
}

type service struct {
	cfg     config.Config
	store   store.Store
	delayer simulate.Delayer
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{
		cfg:     cfg,
		store:   st,
		delayer: simulate.NewDelayer(cfg.SimulateDelay),
	}
}

func (s *service) ListProviderKeys() []types.ProviderKey {
	return s.store.Keys().ProviderKeys()
}

func (s *service) CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.ProviderKey{}, err
	}
	created := types.ProviderKey{
		ID:       fmt.Sprintf("pk-%d", time.Now().UnixMilli()),
		Provider: input.Provider, Name: input.Name, KeyPrefix: "sk-new...",
		Status: "active", CreatedAt: s.cfg.DemoToday, RotateEnabled: false,
	}
	keys := s.store.Keys().ProviderKeys()
	keys = append(keys, created)
	s.store.Keys().SetProviderKeys(keys)
	return created, nil
}

func (s *service) ToggleProviderKey(ctx context.Context, id string) error {
	_ = id
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	return nil
}

func (s *service) RotateProviderKey(ctx context.Context, id string) (types.ProviderKey, error) {
	if err := s.delayer.Wait(ctx, time.Second); err != nil {
		return types.ProviderKey{}, err
	}
	keys := s.store.Keys().ProviderKeys()
	for i := range keys {
		if keys[i].ID == id {
			keys[i].KeyPrefix = fmt.Sprintf("sk-rot-%x...", time.Now().UnixMilli())
			lastUsed := time.Now().Format("2006-01-02 15:04")
			keys[i].LastUsed = &lastUsed
			s.store.Keys().SetProviderKeys(keys)
			return keys[i], nil
		}
	}
	return types.ProviderKey{}, domain.NewDomainError(404, "Not found")
}

func (s *service) DeleteProviderKey(id string) error {
	_ = id
	return nil
}

func (s *service) ListPlatformKeys(memberID, budgetGroupID string) types.PageResult[types.PlatformKey] {
	items := s.store.Keys().PlatformKeys()
	filtered := make([]types.PlatformKey, 0, len(items))
	for _, key := range items {
		if memberID != "" && (key.MemberID == nil || *key.MemberID != memberID) {
			continue
		}
		if budgetGroupID != "" && (key.BudgetGroupID == nil || *key.BudgetGroupID != budgetGroupID) {
			continue
		}
		filtered = append(filtered, key)
	}
	return types.PageResult[types.PlatformKey]{
		Items: filtered, Total: len(filtered), Page: 1, PageSize: 20,
	}
}

func (s *service) QuotaSummary(memberID string) types.MemberQuotaSummary {
	if memberID == "" {
		memberID = "m-1"
	}
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	reservedPool := budgetlookup.GetReservedPoolForMember(tree, members, memberID)
	return memberquota.BuildQuotaSummary(pools, platformKeys, memberID, reservedPool)
}

func (s *service) CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	groups := s.store.Budget().Groups()

	if input.BudgetGroupID != nil {
		var group *types.BudgetGroup
		for i := range groups {
			if groups[i].ID == *input.BudgetGroupID {
				group = &groups[i]
				break
			}
		}
		if group == nil {
			return types.PlatformKey{}, domain.NewDomainError(404, "Budget group not found")
		}
		if msg := budgetgroupquota.ValidateGroupKeyQuota(*group, platformKeys, input.Quota, ""); msg != nil {
			return types.PlatformKey{}, domain.NewDomainError(422, *msg)
		}
		if input.MemberID != nil {
			if msg := routingutil.ValidateModelsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, pkgconst.ModelNotInDeptMessage); msg != nil {
				return types.PlatformKey{}, domain.NewDomainError(422, *msg)
			}
		}
	} else {
		if input.MemberID == nil {
			return types.PlatformKey{}, domain.NewDomainError(400, "memberId required")
		}
		if msg := routingutil.ValidateModelsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, pkgconst.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.NewDomainError(422, *msg)
		}
		if input.Quota > memberquota.GetQuotaRemaining(pools, platformKeys, *input.MemberID) {
			return types.PlatformKey{}, domain.NewDomainError(422, "额度不足，请先申请追加")
		}
	}

	fullKey := fmt.Sprintf("tj-%d-demo-secret-key", time.Now().UnixMilli())
	memberName := (*string)(nil)
	if input.MemberID != nil {
		if member, ok := queryutil.FindMemberByID(members, *input.MemberID); ok {
			memberName = &member.Name
		}
	}
	var groupName *string
	if input.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *input.BudgetGroupID {
				groupName = &group.Name
				break
			}
		}
	}
	keyPrefix := fullKey
	if len(keyPrefix) > 12 {
		keyPrefix = keyPrefix[:12] + "..."
	}
	created := types.PlatformKey{
		ID:   fmt.Sprintf("plk-%d", time.Now().UnixMilli()),
		Name: input.Name, KeyPrefix: keyPrefix, FullKey: &fullKey,
		MemberID: input.MemberID, MemberName: memberName, AppName: input.AppName,
		BudgetGroupID: input.BudgetGroupID, BudgetGroupName: groupName,
		Status: "active", Quota: input.Quota, Used: 0,
		ModelWhitelist: append([]string{}, input.ModelWhitelist...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	}
	platformKeys = append(platformKeys, created)
	s.store.Keys().SetPlatformKeys(platformKeys)
	return created, nil
}

func (s *service) UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	idx := -1
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.PlatformKey{}, domain.NewDomainError(404, "Not found")
	}
	existing := platformKeys[idx]
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	pools := s.store.Budget().MemberQuotaPools()
	groups := s.store.Budget().Groups()

	if len(input.ModelWhitelist) > 0 && existing.MemberID != nil {
		if msg := routingutil.ValidateModelsForMember(*existing.MemberID, input.ModelWhitelist, members, departments, rules, models, pkgconst.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.NewDomainError(422, *msg)
		}
	}
	if input.Quota != nil {
		if existing.BudgetGroupID != nil {
			var group *types.BudgetGroup
			for i := range groups {
				if groups[i].ID == *existing.BudgetGroupID {
					group = &groups[i]
					break
				}
			}
			if group == nil {
				return types.PlatformKey{}, domain.NewDomainError(404, "Budget group not found")
			}
			if msg := budgetgroupquota.ValidateGroupKeyQuota(*group, platformKeys, *input.Quota, existing.ID); msg != nil {
				return types.PlatformKey{}, domain.NewDomainError(422, *msg)
			}
		} else if existing.MemberID != nil {
			otherAllocated := 0.0
			for _, key := range platformKeys {
				if key.MemberID != nil && *key.MemberID == *existing.MemberID && key.BudgetGroupID == nil && key.Status == "active" && key.ID != existing.ID {
					otherAllocated += key.Quota
				}
			}
			if otherAllocated+*input.Quota > memberquota.GetPersonalQuota(pools, *existing.MemberID) {
				return types.PlatformKey{}, domain.NewDomainError(422, "额度不足，请先申请追加")
			}
		}
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Quota != nil {
		existing.Quota = *input.Quota
	}
	if input.ModelWhitelist != nil {
		existing.ModelWhitelist = append([]string{}, input.ModelWhitelist...)
	}
	platformKeys[idx] = existing
	s.store.Keys().SetPlatformKeys(platformKeys)
	return existing, nil
}

func (s *service) TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			if enabled {
				platformKeys[i].Status = "active"
			} else {
				platformKeys[i].Status = "disabled"
			}
			s.store.Keys().SetPlatformKeys(platformKeys)
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NewDomainError(404, "Not found")
}

func (s *service) RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			fullKey := fmt.Sprintf("tj-rot-%d-demo-secret", time.Now().UnixMilli())
			platformKeys[i].FullKey = &fullKey
			prefix := fullKey
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			platformKeys[i].KeyPrefix = prefix
			s.store.Keys().SetPlatformKeys(platformKeys)
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NewDomainError(404, "Not found")
}

func (s *service) RevokePlatformKey(ctx context.Context, id string) error {
	_ = id
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	return nil
}

func (s *service) DeletePlatformKey(id string) error {
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			platformKeys = append(platformKeys[:i], platformKeys[i+1:]...)
			s.store.Keys().SetPlatformKeys(platformKeys)
			return nil
		}
	}
	return nil
}

func (s *service) ListApprovals(tab, memberID string) []types.KeyApproval {
	items := s.store.Keys().Approvals()
	filtered := make([]types.KeyApproval, 0, len(items))
	for _, item := range items {
		if tab == "pending" && item.Status != "pending" {
			continue
		}
		if tab == "mine" && memberID != "" && item.ApplicantID != memberID {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func (s *service) CreateApproval(ctx context.Context, input types.CreateApprovalInput) (types.KeyApproval, error) {
	if err := s.delayer.Wait(ctx, 400*time.Millisecond); err != nil {
		return types.KeyApproval{}, err
	}
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	if msg := routingutil.ValidateModelsForMember(input.MemberID, input.RequestedModels, members, departments, rules, models, pkgconst.ModelNotInDeptMessage); msg != nil {
		return types.KeyApproval{}, domain.NewDomainError(422, *msg)
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
	s.store.Keys().SetApprovals(approvals)
	return created, nil
}

func (s *service) ApprovalQuotaCheck(id string) types.ApprovalQuotaCheck {
	approvals := s.store.Keys().Approvals()
	var approval *types.KeyApproval
	for i := range approvals {
		if approvals[i].ID == id {
			approval = &approvals[i]
			break
		}
	}
	requested := 0.0
	reservedPool := 0.0
	if approval != nil {
		requested = approval.RequestedQuota
		tree := s.store.Budget().Tree()
		members := s.store.Org().Members()
		reservedPool = budgetlookup.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	}
	return types.ApprovalQuotaCheck{
		Sufficient: requested <= reservedPool, ReservedPool: reservedPool, Requested: requested,
	}
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
		return domain.NewDomainError(404, "Not found")
	}
	approval := approvals[idx]
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	reservedPool := budgetlookup.GetReservedPoolForMember(tree, members, approval.ApplicantID)
	if approval.Type == "quota" && approval.RequestedQuota > reservedPool {
		return domain.NewDomainError(422, "Reserved pool insufficient")
	}
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
		s.store.Keys().SetPlatformKeys(platformKeys)
	} else if approval.Type == "quota" {
		memberquota.AddPersonalQuota(pools, approval.ApplicantID, approval.RequestedQuota)
	}
	s.store.Budget().SetMemberQuotaPools(pools)

	approver := sessionutil.ResolveDemoMemberName(approverMemberID, members)
	now := time.Now().Format("2006-01-02 15:04")
	approvals[idx].Status = "approved"
	approvals[idx].Approver = &approver
	approvals[idx].ResolvedAt = &now
	s.store.Keys().SetApprovals(approvals)
	return nil
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
			s.store.Keys().SetApprovals(approvals)
			return nil
		}
	}
	return nil
}
