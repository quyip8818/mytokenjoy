package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *pgBudgetRepo) GetProjectBudget(ctx context.Context, projectID string) (float64, float64, bool, error) {
	companyID := store.CompanyID(ctx)
	var budget float64
	err := r.db.QueryRow(ctx, `
		SELECT budget FROM projects WHERE company_id = $1 AND id = $2
	`, companyID, projectID).Scan(&budget)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return budget, 0, true, nil
}

func (r *pgBudgetRepo) Projects(ctx context.Context) ([]types.Project, error) {
	companyID := store.CompanyID(ctx)

	rows, err := r.db.Query(ctx, `
		SELECT id, name, budget, owner_department_id
		FROM projects WHERE company_id = $1 ORDER BY id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]types.Project, 0)
	projectIndex := make(map[string]int)
	for rows.Next() {
		var p types.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Budget, &p.OwnerDepartmentID); err != nil {
			return nil, err
		}
		p.Consumed = 0
		p.MemberIDs = []string{}
		projectIndex[p.ID] = len(projects)
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return projects, nil
	}

	memberRows, err := r.db.Query(ctx, `
		SELECT project_id, member_id FROM project_members
		WHERE company_id = $1 ORDER BY project_id, member_id
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()
	for memberRows.Next() {
		var projectID, memberID string
		if err := memberRows.Scan(&projectID, &memberID); err != nil {
			return nil, err
		}
		if idx, ok := projectIndex[projectID]; ok {
			projects[idx].MemberIDs = append(projects[idx].MemberIDs, memberID)
		}
	}
	return projects, memberRows.Err()
}

func (r *pgBudgetRepo) SetProjects(ctx context.Context, projects []types.Project) error {
	companyID := store.CompanyID(ctx)
	cloned := cloneProjects(projects)
	ids := make([]string, len(cloned))
	for i, project := range cloned {
		ids[i] = project.ID
		if _, err := r.db.Exec(ctx, `
			INSERT INTO projects (id, company_id, name, budget, owner_department_id, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (company_id, id) DO UPDATE SET
				name = EXCLUDED.name,
				budget = EXCLUDED.budget,
				owner_department_id = EXCLUDED.owner_department_id,
				updated_at = NOW()
		`, project.ID, companyID, project.Name, project.Budget, project.OwnerDepartmentID); err != nil {
			return fmt.Errorf("upsert project %s: %w", project.ID, err)
		}
		if _, err := r.db.Exec(ctx, `
			DELETE FROM project_members WHERE company_id = $1 AND project_id = $2
		`, companyID, project.ID); err != nil {
			return err
		}
		for _, memberID := range project.MemberIDs {
			if _, err := r.db.Exec(ctx, `
				INSERT INTO project_members (company_id, project_id, member_id) VALUES ($1, $2, $3)
			`, companyID, project.ID, memberID); err != nil {
				return err
			}
		}
	}
	if len(ids) == 0 {
		if _, err := r.db.Exec(ctx, `DELETE FROM project_members WHERE company_id = $1`, companyID); err != nil {
			return err
		}
		_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE company_id = $1`, companyID)
		return err
	}
	if err := pruneByColumnForCompany(ctx, r.db, "project_members", "project_id", companyID, ids); err != nil {
		return err
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE platform_keys SET project_id = NULL, updated_at = NOW()
		WHERE company_id = $1 AND project_id IS NOT NULL AND NOT (project_id = ANY($2))
	`, companyID, ids); err != nil {
		return fmt.Errorf("detach platform keys from pruned projects: %w", err)
	}
	return pruneByIDForCompany(ctx, r.db, "projects", companyID, ids)
}
