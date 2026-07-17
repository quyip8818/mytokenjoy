package budget

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
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
	if project.OwnerDepartmentID == uuid.Nil {
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
			ID:                uuid.Must(uuid.NewV7()),
			Name:              trimmedName,
			Budget:            project.Budget,
			Consumed:          0,
			MemberIDs:         append([]uuid.UUID{}, project.MemberIDs...),
			MemberBudgets:     cloneMemberBudgetsUUID(project.MemberBudgets),
			OwnerDepartmentID: project.OwnerDepartmentID,
		}
		if err := validateProjectMemberBudgetsUUID(created.Budget, created.MemberIDs, created.MemberBudgets); err != nil {
			return err
		}
		projects = append(projects, created)
		if err := tx.Budget().SetProjects(ctx, projects); err != nil {
			return fmt.Errorf("persist projects: %w", err)
		}
		result = created
		return nil
	})
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
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return types.Project{}, err
	}
	var result types.Project
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		for i := range projects {
			if projects[i].ID == parsedID {
				if patch.Name != nil && *patch.Name != "" {
					projects[i].Name = *patch.Name
				}
				if patch.Budget != nil {
					projects[i].Budget = *patch.Budget
				}
				if patch.MemberIDs != nil {
					parsed := make([]uuid.UUID, 0, len(*patch.MemberIDs))
					for _, s := range *patch.MemberIDs {
						id, err := uuid.Parse(s)
						if err != nil {
							return err
						}
						parsed = append(parsed, id)
					}
					projects[i].MemberIDs = parsed
					projects[i].MemberBudgets = pruneMemberBudgetsUUID(projects[i].MemberBudgets, projects[i].MemberIDs)
				}
				if patch.MemberBudgets != nil {
					merged, err := mergeMemberBudgetPatchUUID(projects[i].MemberBudgets, *patch.MemberBudgets, projects[i].MemberIDs)
					if err != nil {
						return err
					}
					projects[i].MemberBudgets = merged
				}
				if patch.OwnerDepartmentID != nil && *patch.OwnerDepartmentID != "" {
					ownerID, err := uuid.Parse(*patch.OwnerDepartmentID)
					if err != nil {
						return err
					}
					projects[i].OwnerDepartmentID = ownerID
				}
				if err := validateProjectMemberBudgetsUUID(projects[i].Budget, projects[i].MemberIDs, projects[i].MemberBudgets); err != nil {
					return err
				}
				if err := tx.Budget().SetProjects(ctx, projects); err != nil {
					return fmt.Errorf("persist projects: %w", err)
				}
				budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, tx.BudgetConsumed(), tx.Org(), tx.Budget(), tx.Keys(), s.cfg.Clock())
				if err != nil {
					return fmt.Errorf("load budget context: %w", err)
				}
				for _, loaded := range budgetCtx.Projects {
					if loaded.ID == parsedID {
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
	return result, err
}

func (s *service) DeleteProject(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	var deletedMemberIDs []uuid.UUID
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		projects, err := tx.Budget().Projects(ctx)
		if err != nil {
			return err
		}
		for i := range projects {
			if projects[i].ID == parsedID {
				deletedMemberIDs = append([]uuid.UUID{}, projects[i].MemberIDs...)
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
		for _, memberID := range deletedMemberIDs {
			if rebalErr := s.enqueuer.InsertRebalance(ctx, store.CompanyID(ctx), store.RebalanceAxisMember, memberID.String()); rebalErr != nil {
				return rebalErr
			}
		}
	}
	return err
}
