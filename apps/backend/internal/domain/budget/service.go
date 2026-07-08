package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetTree(ctx context.Context) ([]types.BudgetNode, error)
	UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error)
	ListMemberQuotas(ctx context.Context, deptID string) ([]types.MemberBudgetQuota, error)
	UpdateMemberQuota(ctx context.Context, memberID string, personalQuota float64) (types.MemberBudgetQuota, error)
	ListGroups(ctx context.Context) ([]types.BudgetGroup, error)
	CreateGroup(ctx context.Context, group types.BudgetGroup) (types.BudgetGroup, error)
	UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error)
	DeleteGroup(ctx context.Context, id string) error
	GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error)
	UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error)
	ListAlerts(ctx context.Context) ([]types.AlertRule, error)
	CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error)
	UpdateAlert(ctx context.Context, id string, patch types.AlertRule) (types.AlertRule, error)
	DeleteAlert(ctx context.Context, id string) error
	ListApprovals(ctx context.Context) ([]types.BudgetApproval, error)
	ResolveApproval(ctx context.Context, id string, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error)
}

type service struct {
	cfg     config.Config
	store   store.Store
	delayer common.Delayer
}

func NewService(cfg config.Config, st store.Store, delayer common.Delayer) Service {
	return &service{
		cfg:     cfg,
		store:   st,
		delayer: delayer,
	}
}

func (s *service) GetTree(ctx context.Context) ([]types.BudgetNode, error) {
	return pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes())
}

func (s *service) UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetNode{}, err
	}
	nodes, err := s.store.Org().Nodes().Tree(ctx)
	if err != nil {
		return types.BudgetNode{}, err
	}
	tree := types.OrgNodesToBudgetTree(nodes)
	existing := pkgbudget.FindBudgetNode(tree, id)
	if existing == nil {
		return types.BudgetNode{}, domain.NotFound("Node not found")
	}
	reserved := existing.ReservedPool
	if reservedPool != nil {
		reserved = reservedPool
	}
	reservedValue := 0.0
	if reserved != nil {
		reservedValue = *reserved
	}
	if msg := pkgbudget.ValidateBudgetNodeUpdate(tree, id, budget, reservedValue); msg != nil {
		return types.BudgetNode{}, domain.Validation(*msg)
	}
	update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
	if !pkgbudget.UpdateBudgetNodeInTree(tree, id, update) {
		return types.BudgetNode{}, domain.NotFound("Node not found")
	}
	types.ApplyBudgetTreeToOrgNodes(nodes, tree)
	if err := s.store.Org().Nodes().SetTree(ctx, nodes); err != nil {
		return types.BudgetNode{}, fmt.Errorf("persist budget tree: %w", err)
	}
	updated := pkgbudget.FindBudgetNode(tree, id)
	return *updated, nil
}

func (s *service) ListMemberQuotas(ctx context.Context, deptID string) ([]types.MemberBudgetQuota, error) {
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes())
	if err != nil {
		return nil, err
	}
	if pkgbudget.FindBudgetNode(tree, deptID) == nil {
		return nil, domain.NotFound("Department not found")
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys())
	if err != nil {
		return nil, err
	}
	quotas := make([]types.MemberBudgetQuota, 0)
	for _, member := range members {
		if member.DepartmentID == deptID {
			quotas = append(quotas, pkgbudget.BuildMemberBudgetQuota(member, platformKeys))
		}
	}
	return quotas, nil
}

func (s *service) UpdateMemberQuota(ctx context.Context, memberID string, personalQuota float64) (types.MemberBudgetQuota, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.MemberBudgetQuota{}, err
	}
	tree, err := pkgbudget.LoadBudgetTreeWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes())
	if err != nil {
		return types.MemberBudgetQuota{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.MemberBudgetQuota{}, err
	}
	platformKeys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys())
	if err != nil {
		return types.MemberBudgetQuota{}, err
	}
	if msg := pkgbudget.ValidateMemberQuotaUpdate(tree, members, platformKeys, memberID, personalQuota); msg != nil {
		return types.MemberBudgetQuota{}, domain.Validation(*msg)
	}
	result, updatedMembers := pkgbudget.ApplyMemberQuotaUpdate(members, platformKeys, memberID, personalQuota)
	if err := s.store.Org().SetMembers(ctx, updatedMembers); err != nil {
		return types.MemberBudgetQuota{}, fmt.Errorf("persist member personal quota: %w", err)
	}
	return result, nil
}

func (s *service) ListGroups(ctx context.Context) ([]types.BudgetGroup, error) {
	return pkgbudget.LoadBudgetGroupsWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget())
}

func (s *service) CreateGroup(ctx context.Context, group types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return types.BudgetGroup{}, err
	}
	created := types.BudgetGroup{
		ID:   fmt.Sprintf("bg-%d", time.Now().UnixMilli()),
		Name: group.Name, Budget: group.Budget, Consumed: 0,
		MemberIDs:     append([]string{}, group.MemberIDs...),
		DepartmentIDs: append([]string{}, group.DepartmentIDs...),
	}
	groups = append(groups, created)
	if err := s.store.Budget().SetGroups(ctx, groups); err != nil {
		return types.BudgetGroup{}, fmt.Errorf("persist budget groups: %w", err)
	}
	return created, nil
}

func (s *service) UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return types.BudgetGroup{}, err
	}
	for i := range groups {
		if groups[i].ID == id {
			if patch.Name != "" {
				groups[i].Name = patch.Name
			}
			if patch.Budget != 0 {
				groups[i].Budget = patch.Budget
			}
			if patch.MemberIDs != nil {
				groups[i].MemberIDs = append([]string{}, patch.MemberIDs...)
			}
			if patch.DepartmentIDs != nil {
				groups[i].DepartmentIDs = append([]string{}, patch.DepartmentIDs...)
			}
			if err := s.store.Budget().SetGroups(ctx, groups); err != nil {
				return types.BudgetGroup{}, fmt.Errorf("persist budget groups: %w", err)
			}
			updated := groups[i]
			enriched, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget())
			if err != nil {
				return types.BudgetGroup{}, fmt.Errorf("load budget group consumption: %w", err)
			}
			for _, group := range enriched {
				if group.ID == id {
					return group, nil
				}
			}
			return updated, nil
		}
	}
	return types.BudgetGroup{}, domain.NotFound("Not found")
}

func (s *service) DeleteGroup(ctx context.Context, id string) error {
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return err
	}
	for i := range groups {
		if groups[i].ID == id {
			groups = append(groups[:i], groups[i+1:]...)
			if err := s.store.Budget().SetGroups(ctx, groups); err != nil {
				return fmt.Errorf("persist budget groups: %w", err)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) GetOverrunPolicy(ctx context.Context) (types.OverrunPolicyConfig, error) {
	return s.store.Budget().OverrunPolicy(ctx)
}

func (s *service) UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	if err := s.store.Budget().SetOverrunPolicy(ctx, policy); err != nil {
		return types.OverrunPolicyConfig{}, fmt.Errorf("persist overrun policy: %w", err)
	}
	return policy, nil
}

func (s *service) ListAlerts(ctx context.Context) ([]types.AlertRule, error) {
	return s.store.Budget().AlertRules(ctx)
}

func (s *service) CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.AlertRule{}, err
	}
	created := rule
	created.ID = fmt.Sprintf("alert-%d", time.Now().UnixMilli())
	rules, err := s.store.Budget().AlertRules(ctx)
	if err != nil {
		return types.AlertRule{}, err
	}
	rules = append(rules, created)
	if err := s.store.Budget().SetAlertRules(ctx, rules); err != nil {
		return types.AlertRule{}, fmt.Errorf("persist alert rules: %w", err)
	}
	return created, nil
}

func (s *service) UpdateAlert(ctx context.Context, id string, patch types.AlertRule) (types.AlertRule, error) {
	rules, err := s.store.Budget().AlertRules(ctx)
	if err != nil {
		return types.AlertRule{}, err
	}
	for i := range rules {
		if rules[i].ID == id {
			if patch.NodeID != "" {
				rules[i].NodeID = patch.NodeID
			}
			if patch.NodeName != "" {
				rules[i].NodeName = patch.NodeName
			}
			if patch.Thresholds != nil {
				rules[i].Thresholds = append([]int{}, patch.Thresholds...)
			}
			if patch.NotifyRoleIDs != nil {
				rules[i].NotifyRoleIDs = append([]string{}, patch.NotifyRoleIDs...)
			}
			rules[i].Enabled = patch.Enabled
			if err := s.store.Budget().SetAlertRules(ctx, rules); err != nil {
				return types.AlertRule{}, fmt.Errorf("persist alert rules: %w", err)
			}
			return rules[i], nil
		}
	}
	return types.AlertRule{}, domain.NotFound("Not found")
}

func (s *service) DeleteAlert(ctx context.Context, id string) error {
	rules, err := s.store.Budget().AlertRules(ctx)
	if err != nil {
		return err
	}
	filtered := make([]types.AlertRule, 0, len(rules))
	found := false
	for _, rule := range rules {
		if rule.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, rule)
	}
	if !found {
		return domain.NotFound("Not found")
	}
	if err := s.store.Budget().SetAlertRules(ctx, filtered); err != nil {
		return fmt.Errorf("persist alert rules: %w", err)
	}
	return nil
}

func (s *service) ListApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return nil, err
	}
	return s.store.Budget().BudgetApprovals(ctx)
}

func (s *service) ResolveApproval(ctx context.Context, id string, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return types.BudgetApproval{}, err
	}
	if input.Status != "approved" && input.Status != "rejected" {
		return types.BudgetApproval{}, domain.Validation("invalid status")
	}
	if input.Status == "rejected" && (input.RejectReason == nil || *input.RejectReason == "") {
		return types.BudgetApproval{}, domain.Validation("reject reason required")
	}
	items, err := s.store.Budget().BudgetApprovals(ctx)
	if err != nil {
		return types.BudgetApproval{}, err
	}
	idx := -1
	for i := range items {
		if items[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.BudgetApproval{}, domain.NotFound("Not found")
	}
	if items[idx].Status != "pending" {
		return types.BudgetApproval{}, domain.Validation("approval already resolved")
	}
	now := time.Now().UTC()
	if err := s.store.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
		return types.BudgetApproval{}, err
	}
	resolved := now.Format("2006-01-02 15:04")
	items[idx].Status = input.Status
	items[idx].ResolvedAt = &resolved
	if input.Status == "rejected" {
		items[idx].RejectReason = input.RejectReason
	}
	return items[idx], nil
}
