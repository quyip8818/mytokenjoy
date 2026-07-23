package structure

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) CreateMember(ctx context.Context, input types.Member) (types.Member, error) {
	departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
	if err != nil {
		return types.Member{}, err
	}
	dept := pkgorg.FindDepartment(departments, input.DepartmentID)
	if dept == nil {
		return types.Member{}, domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
	}

	deptName := dept.Name
	if path := pkgorg.GetDeptPath(departments, input.DepartmentID); path != nil {
		deptName = *path
	}

	// Resolve or create user for this member.
	userID, err := s.resolveOrCreateUser(ctx, input.Phone, input.Email, input.Alias)
	if err != nil {
		return types.Member{}, err
	}

	member := types.Member{
		ID:       generateID(),
		UserID:   userID,
		Alias:    input.Alias,
		Username: input.Username, EmployeeID: input.EmployeeID,
		JobTitle: input.JobTitle, HireDate: input.HireDate,
		DepartmentID: input.DepartmentID, DepartmentName: deptName,
		Status: types.MemberStatusActive, Roles: []string{grants.RoleMember}, Source: "manual",
		PersonalBudget: common.DefaultPersonalBudget,
	}

	err = s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		if err := s.checkTrialMemberLimit(ctx, members); err != nil {
			return err
		}
		members = append(members, member)
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		if err := persistRecalculatedMemberCounts(ctx, st, members); err != nil {
			return err
		}
		return core.BumpAuthzRevisionStore(ctx, st)
	})
	if err != nil {
		return types.Member{}, mapMemberUniqueError(err)
	}
	return member, nil
}

func (s *LocalService) UpdateMember(ctx context.Context, id uuid.UUID, input types.Member) (types.Member, error) {
	if err := validateRolesNotEscalated(input.Roles); err != nil {
		return types.Member{}, err
	}
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.Member{}, err
	}
	for i := range members {
		if members[i].ID == id {
			existing := members[i]
			// Merge: only overwrite non-zero fields from input.
			// Track user-owned field changes in OverrideFields.
			if input.Alias != "" && input.Alias != existing.Alias {
				existing.OverrideFields = core.TrackOverride(existing.OverrideFields, "alias")
				existing.Alias = input.Alias
			}
			if input.Username != "" {
				existing.Username = input.Username
			}
			if input.EmployeeID != "" {
				existing.EmployeeID = input.EmployeeID
			}
			if input.JobTitle != "" {
				existing.JobTitle = input.JobTitle
			}
			if input.HireDate != "" {
				existing.HireDate = input.HireDate
			}
			if input.DepartmentID != uuid.Nil {
				existing.DepartmentID = input.DepartmentID
				existing.DepartmentName = input.DepartmentName
			}
			if len(input.Roles) > 0 {
				rolesChanged := !slices.Equal(existing.Roles, input.Roles)
				existing.Roles = input.Roles
				if rolesChanged {
					if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
						return types.Member{}, err
					}
				}
			}
			if input.Status != "" {
				existing.Status = input.Status
			}

			members[i] = existing
			if err := s.d.Store.Org().SetMembers(ctx, members); err != nil {
				return types.Member{}, err
			}

			// Update phone/email on users table if provided.
			if input.Phone != "" {
				if err := s.d.Store.User().UpdatePhone(ctx, existing.UserID, input.Phone); err != nil {
					return types.Member{}, err
				}
			}
			if input.Email != "" {
				if err := s.d.Store.User().UpdateEmail(ctx, existing.UserID, input.Email); err != nil {
					return types.Member{}, err
				}
			}

			return existing, nil
		}
	}
	return types.Member{}, domain.NewDomainError(404, "types.Member not found")
}

func (s *LocalService) UpdateMemberStatus(ctx context.Context, ids []uuid.UUID, status string) error {
	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		keys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			return err
		}
		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			members[i].Status = status
			if status == "inactive" {
				for j := range keys {
					if keys[j].MemberID != nil && *keys[j].MemberID == members[i].ID {
						keys[j].Status = "disabled"
					}
				}
			}
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
			return err
		}
		return core.BumpAuthzRevisionStore(ctx, st)
	})
}

func (s *LocalService) TransferMembers(ctx context.Context, ids []uuid.UUID, departmentID uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		departments, err := common.LoadDepartments(ctx, st.Org().Nodes())
		if err != nil {
			return err
		}
		target := pkgorg.FindDepartment(departments, departmentID)
		if target == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		deptName := target.Name
		if path := pkgorg.GetDeptPath(departments, departmentID); path != nil {
			deptName = *path
		}

		// Load target department's member_avg_budget for personal budget reassignment
		targetBudgetRow, found, err := st.Budget().OrgNodeBudget().Get(ctx, departmentID)
		if err != nil {
			return err
		}
		targetAvgBudget := int64(0)
		if found && targetBudgetRow.MemberAvgBudget > 0 {
			targetAvgBudget = targetBudgetRow.MemberAvgBudget
		}

		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			members[i].DepartmentID = departmentID
			members[i].DepartmentName = deptName

			// Update personal budget to target department's average
			if targetAvgBudget > 0 {
				members[i].PersonalBudget = targetAvgBudget
			}

			mappings, err := st.PlatformKeyMappings().ListMappingsByMemberID(ctx, members[i].ID)
			if err != nil {
				return err
			}
			for _, mapping := range mappings {
				mapping.DepartmentID = departmentID
				if err := st.PlatformKeyMappings().UpsertMapping(ctx, mapping); err != nil {
					return err
				}
			}
		}

		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		return persistRecalculatedMemberCounts(ctx, st, members)
	})
}
