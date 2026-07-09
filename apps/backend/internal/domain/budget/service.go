package budget

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func generateBudgetID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%d-%x", prefix, time.Now().UnixMilli(), b)
}

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
	GetGroupMemberConsumed(ctx context.Context, groupID string) (map[string]float64, error)
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
	var result types.BudgetNode
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		nodes, err := tx.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		tree := types.OrgNodesToBudgetTree(nodes)
		existing := pkgbudget.FindBudgetNode(tree, id)
		if existing == nil {
			return domain.NotFound("Node not found")
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
			return domain.Validation(*msg)
		}
		update := types.BudgetNode{Budget: budget, ReservedPool: reserved}
		if !pkgbudget.UpdateBudgetNodeInTree(tree, id, update) {
			return domain.NotFound("Node not found")
		}
		types.ApplyBudgetTreeToOrgNodes(nodes, tree)
		if err := tx.Org().Nodes().SetTree(ctx, nodes); err != nil {
			return fmt.Errorf("persist budget tree: %w", err)
		}
		updated := pkgbudget.FindBudgetNode(tree, id)
		result = *updated
		return nil
	})
	return result, err
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
	var result types.BudgetGroup
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		created := types.BudgetGroup{
			ID:   generateBudgetID("bg"),
			Name: group.Name, Budget: group.Budget, Consumed: 0,
			MemberIDs:     append([]string{}, group.MemberIDs...),
			DepartmentIDs: append([]string{}, group.DepartmentIDs...),
		}
		groups = append(groups, created)
		if err := tx.Budget().SetGroups(ctx, groups); err != nil {
			return fmt.Errorf("persist budget groups: %w", err)
		}
		result = created
		return nil
	})
	return result, err
}

func (s *service) UpdateGroup(ctx context.Context, id string, patch types.BudgetGroup) (types.BudgetGroup, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.BudgetGroup{}, err
	}
	var result types.BudgetGroup
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for i := range groups {
			if groups[i].ID == id {
				if patch.Name != "" {
					groups[i].Name = patch.Name
				}
				groups[i].Budget = patch.Budget
				if patch.MemberIDs != nil {
					groups[i].MemberIDs = append([]string{}, patch.MemberIDs...)
				}
				if patch.DepartmentIDs != nil {
					groups[i].DepartmentIDs = append([]string{}, patch.DepartmentIDs...)
				}
				if err := tx.Budget().SetGroups(ctx, groups); err != nil {
					return fmt.Errorf("persist budget groups: %w", err)
				}
				enriched, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, tx.BudgetSnapshots(), tx.Org(), tx.Budget())
				if err != nil {
					return fmt.Errorf("load budget group consumption: %w", err)
				}
				for _, group := range enriched {
					if group.ID == id {
						result = group
						return nil
					}
				}
				result = groups[i]
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	return result, err
}

func (s *service) DeleteGroup(ctx context.Context, id string) error {
	return s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		groups, err := tx.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for i := range groups {
			if groups[i].ID == id {
				groups = append(groups[:i], groups[i+1:]...)
				if err := tx.Budget().SetGroups(ctx, groups); err != nil {
					return fmt.Errorf("persist budget groups: %w", err)
				}
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
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
	var result types.AlertRule
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		created := rule
		created.ID = generateBudgetID("alert")
		rules, err := tx.Budget().AlertRules(ctx)
		if err != nil {
			return err
		}
		rules = append(rules, created)
		if err := tx.Budget().SetAlertRules(ctx, rules); err != nil {
			return fmt.Errorf("persist alert rules: %w", err)
		}
		result = created
		return nil
	})
	return result, err
}

func (s *service) UpdateAlert(ctx context.Context, id string, patch types.AlertRule) (types.AlertRule, error) {
	var result types.AlertRule
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		rules, err := tx.Budget().AlertRules(ctx)
		if err != nil {
			return err
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
				if err := tx.Budget().SetAlertRules(ctx, rules); err != nil {
					return fmt.Errorf("persist alert rules: %w", err)
				}
				result = rules[i]
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	return result, err
}

func (s *service) DeleteAlert(ctx context.Context, id string) error {
	return s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		rules, err := tx.Budget().AlertRules(ctx)
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
		if err := tx.Budget().SetAlertRules(ctx, filtered); err != nil {
			return fmt.Errorf("persist alert rules: %w", err)
		}
		return nil
	})
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

	approval := items[idx]
	now := time.Now().UTC()

	if input.Status == "approved" {
		// Look up applicant's department
		deptID := approval.DepartmentID
		if deptID == "" {
			member, err := s.store.Org().MemberByID(ctx, approval.ApplicantID)
			if err != nil {
				return types.BudgetApproval{}, err
			}
			if member == nil {
				return types.BudgetApproval{}, domain.NotFound("申请人不存在")
			}
			deptID = member.DepartmentID
		}

		// Load org tree and convert to budget tree
		nodes, err := s.store.Org().Nodes().Tree(ctx)
		if err != nil {
			return types.BudgetApproval{}, err
		}
		tree := types.OrgNodesToBudgetTree(nodes)
		deptNode := pkgbudget.FindBudgetNode(tree, deptID)
		if deptNode == nil {
			return types.BudgetApproval{}, domain.NotFound("部门不存在")
		}

		// Validate reserved pool balance
		reserved := float64(0)
		if deptNode.ReservedPool != nil {
			reserved = *deptNode.ReservedPool
		}
		if reserved < approval.Amount {
			return types.BudgetApproval{}, domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %.2f 元", reserved))
		}

		// Execute in transaction
		if err := s.store.WithTx(ctx, func(txStore store.Store) error {
			if err := txStore.Budget().AcquireBudgetLock(ctx); err != nil {
				return err
			}
			// Update approval status
			if err := txStore.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
				return err
			}
			// Deduct from department reserved pool
			newReserved := reserved - approval.Amount
			deptNode.ReservedPool = &newReserved
			types.ApplyBudgetTreeToOrgNodes(nodes, tree)
			if err := txStore.Org().Nodes().SetTree(ctx, nodes); err != nil {
				return fmt.Errorf("persist budget tree: %w", err)
			}
			// Add to member's personal quota
			members, err := txStore.Org().Members(ctx)
			if err != nil {
				return err
			}
			found := false
			for i := range members {
				if members[i].ID == approval.ApplicantID {
					members[i].PersonalQuota += approval.Amount
					found = true
					break
				}
			}
			if !found {
				return domain.NotFound("申请人不存在")
			}
			if err := txStore.Org().SetMembers(ctx, members); err != nil {
				return fmt.Errorf("persist member personal quota: %w", err)
			}
			return nil
		}); err != nil {
			return types.BudgetApproval{}, err
		}

		// Enqueue rebalance for the member axis
		_ = s.store.Relay().EnqueueRebalance(ctx, store.RebalanceAxisMember, approval.ApplicantID)
	} else {
		// Rejected - just update status
		if err := s.store.Budget().UpdateBudgetApproval(ctx, id, input.Status, input.RejectReason, now); err != nil {
			return types.BudgetApproval{}, err
		}
	}

	resolved := now.Format("2006-01-02 15:04")
	items[idx].Status = input.Status
	items[idx].ResolvedAt = &resolved
	if input.Status == "rejected" {
		items[idx].RejectReason = input.RejectReason
	}
	return items[idx], nil
}

func (s *service) GetGroupMemberConsumed(ctx context.Context, groupID string) (map[string]float64, error) {
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return nil, err
	}
	var target *types.BudgetGroup
	for i := range groups {
		if groups[i].ID == groupID {
			target = &groups[i]
			break
		}
	}
	if target == nil {
		return nil, domain.NotFound("Group not found")
	}

	nodes, err := s.store.Org().Nodes().Tree(ctx)
	if err != nil {
		return nil, err
	}
	tree := types.OrgNodesToBudgetTree(nodes)
	periodKey := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, time.Now().UTC())

	if len(target.DepartmentIDs) > 0 {
		if node := pkgbudget.FindBudgetNode(tree, target.DepartmentIDs[0]); node != nil {
			periodKey = pkgbudget.SnapshotKey(node.Period, time.Now().UTC())
		}
	}

	result := make(map[string]float64)
	for _, memberID := range target.MemberIDs {
		consumed, _, err := s.store.BudgetSnapshots().GetConsumed(ctx, store.SnapshotAxisMember, memberID, periodKey)
		if err != nil {
			return nil, err
		}
		result[memberID] = consumed
	}
	return result, nil
}
