package org

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) budgetPeriod() string {
	if len(s.cfg.DemoToday) >= 7 {
		return s.cfg.DemoToday[:7]
	}
	return time.Now().Format("2006-01")
}

func (s *service) GetDepartmentTree() []types.Department {
	return s.store.Org().Departments()
}

func (s *service) CreateDepartment(ctx context.Context, name, parentID string) (types.Department, error) {
	if strings.TrimSpace(name) == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "types.Department name is required")
	}
	if parentID == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "Parent department is required")
	}

	deptID := fmt.Sprintf("dept-%d", time.Now().UnixMilli())
	var created types.Department

	err := s.store.WithTx(ctx, func(st store.Store) error {
		departments := st.Org().Departments()
		members := st.Org().Members()
		state := &ProvisionState{
			Departments: departments,
			BudgetTree:  st.Budget().Tree(),
			Rules:       st.Models().RoutingRules(),
			Models:      st.Models().Models(),
		}
		if err := ProvisionDepartment(state, ProvisionInput{
			ID: deptID, Name: name, ParentID: parentID, Period: s.budgetPeriod(),
		}); err != nil {
			if strings.Contains(err.Error(), "parent department not found") {
				return domain.NewDomainError(domain.StatusNotFound, "Parent department not found")
			}
			return domain.NewDomainError(domain.StatusUnprocessable, err.Error())
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(state.Rules); err != nil {
			return err
		}

		found := pkgorg.FindDepartment(state.Departments, deptID)
		if found == nil {
			return fmt.Errorf("created department not found")
		}
		manualSource := types.DeptSourceManual
		found.Source = &manualSource
		created = *found
		return nil
	})
	if err != nil {
		return types.Department{}, err
	}
	return created, nil
}

func (s *service) UpdateDepartment(ctx context.Context, id, name string) (types.Department, error) {
	if strings.TrimSpace(name) == "" {
		return types.Department{}, domain.NewDomainError(domain.StatusUnprocessable, "types.Department name is required")
	}

	var updated types.Department
	err := s.store.WithTx(ctx, func(st store.Store) error {
		members := st.Org().Members()
		state := &ProvisionState{
			Departments: st.Org().Departments(),
			BudgetTree:  st.Budget().Tree(),
			Rules:       st.Models().RoutingRules(),
			Models:      st.Models().Models(),
		}
		if pkgorg.FindDepartment(state.Departments, id) == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}
		if err := RenameDepartment(state, id, name); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(state.Rules); err != nil {
			return err
		}

		found := pkgorg.FindDepartment(state.Departments, id)
		if found == nil {
			return fmt.Errorf("updated department not found")
		}
		updated = *found
		return nil
	})
	if err != nil {
		return types.Department{}, err
	}
	return updated, nil
}

func (s *service) DeleteDepartment(ctx context.Context, id string) error {
	if id == RootDepartmentID {
		return domain.NewDomainError(domain.StatusUnprocessable, DeptDeleteBlockedMsg)
	}

	return s.store.WithTx(ctx, func(st store.Store) error {
		departments := st.Org().Departments()
		members := st.Org().Members()

		if pkgorg.FindDepartment(departments, id) == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}
		if pkgorg.HasDirectChildDepartments(departments, id) || pkgorg.HasDirectActiveMembers(members, id) {
			return domain.NewDomainError(domain.StatusUnprocessable, DeptDeleteBlockedMsg)
		}

		state := &ProvisionState{
			Departments: departments,
			BudgetTree:  st.Budget().Tree(),
			Rules:       st.Models().RoutingRules(),
			Models:      st.Models().Models(),
		}
		if err := DeprovisionDepartment(state, id); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(state.BudgetTree); err != nil {
			return err
		}
		return st.Models().SetRoutingRules(state.Rules)
	})
}
