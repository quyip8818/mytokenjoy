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

func (s *service) GetDepartmentTree(ctx context.Context) ([]types.Department, error) {
	return s.store.Org().Departments(ctx)
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
		departments, err := st.Org().Departments(ctx)
		if err != nil {
			return err
		}
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		state, err := loadProvisionState(ctx, st, departments)
		if err != nil {
			return err
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
		if err := st.Org().SetDepartments(ctx, state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(ctx, state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(ctx, state.Rules); err != nil {
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
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		departments, err := st.Org().Departments(ctx)
		if err != nil {
			return err
		}
		state, err := loadProvisionState(ctx, st, departments)
		if err != nil {
			return err
		}
		if pkgorg.FindDepartment(state.Departments, id) == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}
		if err := RenameDepartment(state, id, name); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(ctx, state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(ctx, state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(ctx, state.Rules); err != nil {
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
		departments, err := st.Org().Departments(ctx)
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
			return domain.NewDomainError(domain.StatusUnprocessable, DeptDeleteBlockedMsg)
		}

		state, err := loadProvisionState(ctx, st, departments)
		if err != nil {
			return err
		}
		if err := DeprovisionDepartment(state, id); err != nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(ctx, state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(ctx, state.BudgetTree); err != nil {
			return err
		}
		return st.Models().SetRoutingRules(ctx, state.Rules)
	})
}
