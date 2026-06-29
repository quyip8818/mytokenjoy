package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/memberbudgetquota"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetTree() []types.BudgetNode
	UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error)
	ListMemberQuotas(deptID string) ([]types.MemberBudgetQuota, error)
	UpdateMemberQuota(ctx context.Context, memberID string, personalQuota float64) (types.MemberBudgetQuota, error)
	ListGroups() []types.BudgetGroup
	CreateGroup(ctx context.Context, group types.BudgetGroup) (types.BudgetGroup, error)
	UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error)
	DeleteGroup(id string) error
	GetOverrunPolicy() types.OverrunPolicyConfig
	UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error)
	ListAlerts() []types.AlertRule
	CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error)
	UpdateAlert(id string, patch types.AlertRule) (types.AlertRule, error)
	DeleteAlert(id string) error
}

type service struct {
	cfg     config.Config
	store   store.Store
	delayer simulate.Delayer
}

func NewService(cfg config.Config, st store.Store, delayer simulate.Delayer) Service {
	return &service{
		cfg:     cfg,
		store:   st,
		delayer: delayer,
	}
}

func (s *service) GetTree() []types.BudgetNode {
	return s.store.Budget().Tree()
}

func (s *service) UpdateNode(ctx context.Context, id string, budget float64, reservedPool *float64) (types.BudgetNode, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetNode{}, err
	}
	tree := s.store.Budget().Tree()
	existing := budgetutil.FindBudgetNode(tree, id)
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
	if msg := budgetutil.ValidateBudgetNodeUpdate(tree, id, budget, reservedValue); msg != nil {
		return types.BudgetNode{}, domain.Validation(*msg)
	}
	update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
	if !budgetutil.UpdateBudgetNodeInTree(tree, id, update) {
		return types.BudgetNode{}, domain.NotFound("Node not found")
	}
	if err := s.store.Budget().SetTree(tree); err != nil {
		return types.BudgetNode{}, fmt.Errorf("persist budget tree: %w", err)
	}
	updated := budgetutil.FindBudgetNode(tree, id)
	return *updated, nil
}

func (s *service) ListMemberQuotas(deptID string) ([]types.MemberBudgetQuota, error) {
	tree := s.store.Budget().Tree()
	if budgetutil.FindBudgetNode(tree, deptID) == nil {
		return nil, domain.NotFound("Department not found")
	}
	members := s.store.Org().Members()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	quotas := make([]types.MemberBudgetQuota, 0)
	for _, member := range members {
		if member.DepartmentID == deptID {
			quotas = append(quotas, memberbudgetquota.BuildMemberBudgetQuota(member, pools, platformKeys))
		}
	}
	return quotas, nil
}

func (s *service) UpdateMemberQuota(ctx context.Context, memberID string, personalQuota float64) (types.MemberBudgetQuota, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.MemberBudgetQuota{}, err
	}
	tree := s.store.Budget().Tree()
	members := s.store.Org().Members()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	if msg := memberbudgetquota.ValidateMemberQuotaUpdate(tree, members, pools, platformKeys, memberID, personalQuota); msg != nil {
		return types.MemberBudgetQuota{}, domain.Validation(*msg)
	}
	result := memberbudgetquota.ApplyMemberQuotaUpdate(members, pools, platformKeys, memberID, personalQuota)
	if err := s.store.Budget().SetMemberQuotaPools(pools); err != nil {
		return types.MemberBudgetQuota{}, fmt.Errorf("persist member quota pools: %w", err)
	}
	return result, nil
}

func (s *service) ListGroups() []types.BudgetGroup {
	return s.store.Budget().Groups()
}

func (s *service) CreateGroup(ctx context.Context, group types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	groups := s.store.Budget().Groups()
	created := types.BudgetGroup{
		ID:   fmt.Sprintf("bg-%d", time.Now().UnixMilli()),
		Name: group.Name, Budget: group.Budget, Consumed: 0,
		MemberIDs:     append([]string{}, group.MemberIDs...),
		DepartmentIDs: append([]string{}, group.DepartmentIDs...),
	}
	groups = append(groups, created)
	if err := s.store.Budget().SetGroups(groups); err != nil {
		return types.BudgetGroup{}, fmt.Errorf("persist budget groups: %w", err)
	}
	return created, nil
}

func (s *service) UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	groups := s.store.Budget().Groups()
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
			if err := s.store.Budget().SetGroups(groups); err != nil {
				return types.BudgetGroup{}, fmt.Errorf("persist budget groups: %w", err)
			}
			return groups[i], nil
		}
	}
	return types.BudgetGroup{}, domain.NotFound("Not found")
}

func (s *service) DeleteGroup(id string) error {
	groups := s.store.Budget().Groups()
	for i := range groups {
		if groups[i].ID == id {
			groups = append(groups[:i], groups[i+1:]...)
			if err := s.store.Budget().SetGroups(groups); err != nil {
				return fmt.Errorf("persist budget groups: %w", err)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) GetOverrunPolicy() types.OverrunPolicyConfig {
	return s.store.Budget().OverrunPolicy()
}

func (s *service) UpdateOverrunPolicy(ctx context.Context, policy types.OverrunPolicyConfig) (types.OverrunPolicyConfig, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.OverrunPolicyConfig{}, err
	}
	if err := s.store.Budget().SetOverrunPolicy(policy); err != nil {
		return types.OverrunPolicyConfig{}, fmt.Errorf("persist overrun policy: %w", err)
	}
	return policy, nil
}

func (s *service) ListAlerts() []types.AlertRule {
	return s.store.Budget().AlertRules()
}

func (s *service) CreateAlert(ctx context.Context, rule types.AlertRule) (types.AlertRule, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.AlertRule{}, err
	}
	created := rule
	created.ID = fmt.Sprintf("alert-%d", time.Now().UnixMilli())
	rules := s.store.Budget().AlertRules()
	rules = append(rules, created)
	if err := s.store.Budget().SetAlertRules(rules); err != nil {
		return types.AlertRule{}, fmt.Errorf("persist alert rules: %w", err)
	}
	return created, nil
}

func (s *service) UpdateAlert(id string, patch types.AlertRule) (types.AlertRule, error) {
	rules := s.store.Budget().AlertRules()
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
			if err := s.store.Budget().SetAlertRules(rules); err != nil {
				return types.AlertRule{}, fmt.Errorf("persist alert rules: %w", err)
			}
			return rules[i], nil
		}
	}
	return types.AlertRule{}, domain.NotFound("Not found")
}

func (s *service) DeleteAlert(id string) error {
	rules := s.store.Budget().AlertRules()
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
	if err := s.store.Budget().SetAlertRules(filtered); err != nil {
		return fmt.Errorf("persist alert rules: %w", err)
	}
	return nil
}
