package budget

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ListProjects(ctx context.Context) ([]types.Project, error) {
	return pkgbudget.LoadProjectsWithConsumed(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.cfg.Clock())
}

func (s *service) CreateProject(ctx context.Context, project types.Project) (types.Project, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.Project{}, err
	}
	if strings.TrimSpace(project.Name) == "" {
		return types.Project{}, domain.Validation("project name is required")
	}
	if len(project.Name) > 100 {
		return types.Project{}, domain.Validation("project name must be 100 characters or less")
	}
	if project.Budget < 0 {
		return types.Project{}, domain.Validation("budget must be non-negative")
	}
	if strings.TrimSpace(project.OwnerDepartmentID) == "" {
		return types.Project{}, domain.Validation("ownerDepartmentId is required")
	}
	var result types.Project
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		trimmedName := strings.TrimSpace(project.Name)
		for _, existing := range projects {
			if existing.Name == trimmedName {
				return domain.Conflict("project name already exists")
			}
		}
		created := types.Project{
			ID:                generateBudgetID("proj"),
			Name:              trimmedName,
			Budget:            project.Budget,
			Consumed:          0,
			MemberIDs:         append([]string{}, project.MemberIDs...),
			OwnerDepartmentID: strings.TrimSpace(project.OwnerDepartmentID),
		}
		projects = append(projects, created)
		if err := tx.Budget().SetProjects(ctx, projects); err != nil {
			return fmt.Errorf("persist projects: %w", err)
		}
		result = created
		return nil
	})
	if err == nil {
		s.logger.Info("budget.project.created", "project_id", result.ID, "name", result.Name, "budget", result.Budget)
	}
	return result, err
}

func (s *service) UpdateProject(ctx context.Context, id string, patch types.UpdateProjectInput) (types.Project, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.Project{}, err
	}
	if patch.Name != nil && len(*patch.Name) > 100 {
		return types.Project{}, domain.Validation("project name must be 100 characters or less")
	}
	if patch.Budget != nil && *patch.Budget < 0 {
		return types.Project{}, domain.Validation("budget must be non-negative")
	}
	var result types.Project
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		for i := range projects {
			if projects[i].ID == id {
				if patch.Name != nil && *patch.Name != "" {
					projects[i].Name = *patch.Name
				}
				if patch.Budget != nil {
					projects[i].Budget = *patch.Budget
				}
				if patch.MemberIDs != nil {
					projects[i].MemberIDs = append([]string{}, (*patch.MemberIDs)...)
				}
				if patch.OwnerDepartmentID != nil && *patch.OwnerDepartmentID != "" {
					projects[i].OwnerDepartmentID = *patch.OwnerDepartmentID
				}
				if err := tx.Budget().SetProjects(ctx, projects); err != nil {
					return fmt.Errorf("persist projects: %w", err)
				}
				budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, tx.BudgetConsumed(), tx.Org(), tx.Budget(), tx.Keys(), s.cfg.Clock())
				if err != nil {
					return fmt.Errorf("load budget context: %w", err)
				}
				for _, loaded := range budgetCtx.Projects {
					if loaded.ID == id {
						result = loaded
						return nil
					}
				}
				result = projects[i]
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	if err == nil {
		s.logger.Info("budget.project.updated", "project_id", id, "name", result.Name, "budget", result.Budget)
	}
	return result, err
}

func (s *service) DeleteProject(ctx context.Context, id string) error {
	var deletedMemberIDs []string
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		for i := range projects {
			if projects[i].ID == id {
				deletedMemberIDs = append([]string{}, projects[i].MemberIDs...)
				projects = append(projects[:i], projects[i+1:]...)
				if err := tx.Budget().SetProjects(ctx, projects); err != nil {
					return fmt.Errorf("persist projects: %w", err)
				}
				return nil
			}
		}
		return domain.NotFound("Not found")
	})
	if err == nil {
		s.logger.Info("budget.project.deleted", "project_id", id)
		for _, memberID := range deletedMemberIDs {
			if rebalErr := s.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, memberID); rebalErr != nil {
				s.logger.Error("enqueue rebalance failed after project delete",
					"project_id", id, "member_id", memberID, "error", rebalErr)
			}
		}
	}
	return err
}
