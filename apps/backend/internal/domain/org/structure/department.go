package structure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *Local) GetDepartmentTree(ctx context.Context) ([]types.Department, error) {
	return common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
}

func (s *Local) CreateDepartment(ctx context.Context, name, parentID string) (types.Department, error) {
	if strings.TrimSpace(name) == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "types.Department name is required")
	}
	if parentID == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "Parent department is required")
	}

	deptID := fmt.Sprintf("dept-%d", time.Now().UnixMilli())
	var created types.Department

	err := s.d.Store.WithTx(ctx, func(st store.Store) error {
		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		state, err := core.LoadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		if err := core.ProvisionDepartment(state, core.ProvisionInput{
			ID: deptID, Name: name, ParentID: parentID, Period: s.d.BudgetPeriod(),
		}); err != nil {
			if strings.Contains(err.Error(), "parent department not found") {
				return domain.NewDomainError(domain.StatusNotFound, "Parent department not found")
			}
			return domain.NewDomainError(domain.StatusUnprocessable, err.Error())
		}

		state.Nodes = core.RecalcDepartmentMemberCounts(state.Nodes, members)
		if err := core.PersistProvisionState(ctx, st, state); err != nil {
			return err
		}

		found := pkgorg.FindOrgNode(state.Nodes, deptID)
		if found == nil {
			return fmt.Errorf("created department not found")
		}
		dept := types.OrgNodeToDepartment(*found)
		manualSource := types.DeptSourceManual
		dept.Source = &manualSource
		created = dept
		return nil
	})
	if err != nil {
		return types.Department{}, err
	}
	return created, nil
}

func (s *Local) UpdateDepartment(ctx context.Context, id, name string) (types.Department, error) {
	if strings.TrimSpace(name) == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "types.Department name is required")
	}

	var updated types.Department
	err := s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		state, err := core.LoadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		if pkgorg.FindOrgNode(state.Nodes, id) == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}
		if err := core.RenameDepartment(state, id, name); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Nodes = core.RecalcDepartmentMemberCounts(state.Nodes, members)
		if err := core.PersistProvisionState(ctx, st, state); err != nil {
			return err
		}
		if err := syncDenormalizedNodeNames(ctx, st, id, name, members); err != nil {
			return err
		}

		found := pkgorg.FindOrgNode(state.Nodes, id)
		if found == nil {
			return fmt.Errorf("updated department not found")
		}
		updated = types.OrgNodeToDepartment(*found)
		return nil
	})
	if err != nil {
		return types.Department{}, err
	}
	return updated, nil
}

func (s *Local) DeleteDepartment(ctx context.Context, id string) error {
	if id == core.RootDepartmentID {
		return domain.NewDomainError(domain.StatusUnprocessable, core.DeptDeleteBlockedMsg)
	}

	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		departments, err := common.LoadDepartments(ctx, st.Org().Nodes())
		if err != nil {
			return err
		}
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}

		if pkgorg.FindDepartment(departments, id) == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}
		if pkgorg.HasDirectChildDepartments(departments, id) || pkgorg.HasDirectActiveMembers(members, id) {
			return domain.NewDomainError(domain.StatusUnprocessable, core.DeptDeleteBlockedMsg)
		}

		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		state, err := core.LoadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		if err := core.DeprovisionDepartment(state, id); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Nodes = core.RecalcDepartmentMemberCounts(state.Nodes, members)
		return core.PersistProvisionState(ctx, st, state)
	})
}

func syncDenormalizedNodeNames(ctx context.Context, st store.Store, nodeID, name string, members []types.Member) error {
	changedMembers := false
	for i := range members {
		if members[i].DepartmentID == nodeID && members[i].DepartmentName != name {
			members[i].DepartmentName = name
			changedMembers = true
		}
	}
	if changedMembers {
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
	}
	rules, err := st.Budget().AlertRules(ctx)
	if err != nil {
		return err
	}
	changedAlerts := false
	for i := range rules {
		if rules[i].NodeID == nodeID && rules[i].NodeName != name {
			rules[i].NodeName = name
			changedAlerts = true
		}
	}
	if changedAlerts {
		return st.Budget().SetAlertRules(ctx, rules)
	}
	return nil
}
