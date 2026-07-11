package structure

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

var protectedRoles = map[string]struct{}{
	permission.RoleSuperAdmin: {},
	permission.RoleOrgAdmin:   {},
}

func validateRolesNotEscalated(roles []string) error {
	for _, role := range roles {
		if _, protected := protectedRoles[role]; protected {
			return domain.Forbidden("cannot assign protected role via member update")
		}
	}
	return nil
}

func (s *LocalService) ListMembers(ctx context.Context, departmentID, keyword string, directOnly bool, page, pageSize int) (types.MemberPageResult, error) {
	items, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.MemberPageResult{}, err
	}
	if departmentID != "" {
		departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
		if err != nil {
			return types.MemberPageResult{}, err
		}
		items = pkgorg.FilterMembersByDepartment(items, departments, departmentID, directOnly)
	}
	// Count pending before keyword filtering so count is always accurate.
	pendingCount := 0
	for _, m := range items {
		if m.Status == types.MemberStatusPending {
			pendingCount++
		}
	}
	if keyword != "" {
		filtered := make([]types.Member, 0)
		for _, member := range items {
			if strings.Contains(member.Name, keyword) {
				filtered = append(filtered, member)
			}
		}
		items = filtered
	}
	paged, total, safePage, safeSize := common.Paginate(items, page, pageSize)
	return types.MemberPageResult{
		Items: paged, Total: total, Page: safePage, PageSize: safeSize,
		PendingCount: pendingCount,
	}, nil
}

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

	member := types.Member{
		ID:   generateID("m"),
		Name: input.Name, Phone: input.Phone, Email: input.Email,
		Username: input.Username, EmployeeID: input.EmployeeID,
		JobTitle: input.JobTitle, HireDate: input.HireDate,
		DepartmentID: input.DepartmentID, DepartmentName: deptName,
		Status: types.MemberStatusActive, Roles: []string{permission.RoleMember}, Source: "manual",
		PersonalBudget: common.DefaultPersonalBudget,
	}

	err = s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
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

func mapMemberUniqueError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return domain.Conflict("邮箱已存在")
		}
		if strings.Contains(pgErr.ConstraintName, "phone") {
			return domain.Conflict("手机号已存在")
		}
		return domain.Conflict("成员信息重复")
	}
	return err
}

func persistRecalculatedMemberCounts(ctx context.Context, st store.Store, members []types.Member) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	nodes = core.RecalcDepartmentMemberCounts(nodes, members)
	return st.Org().Nodes().SetTree(ctx, nodes)
}

func (s *LocalService) UpdateMember(ctx context.Context, id string, input types.Member) (types.Member, error) {
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
			// Merge: only overwrite non-zero fields from input
			if input.Name != "" {
				existing.Name = input.Name
			}
			if input.Phone != "" {
				existing.Phone = input.Phone
			}
			if input.Email != "" {
				existing.Email = input.Email
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
			if input.DepartmentID != "" {
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
			return existing, nil
		}
	}
	return types.Member{}, domain.NewDomainError(404, "types.Member not found")
}

func (s *LocalService) DeleteMembers(ctx context.Context, ids []string) error {
	if sessionCtx, ok := httpx.SessionFromContext(ctx); ok {
		for _, id := range ids {
			if id == sessionCtx.Member.ID {
				return domain.BadRequest("不能删除当前登录的用户")
			}
		}
	}
	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		keys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			return err
		}

		idSet := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		// Disable keys belonging to deleted members and detach member reference
		for i := range keys {
			if keys[i].MemberID != nil {
				if _, ok := idSet[*keys[i].MemberID]; ok {
					keys[i].Status = "disabled"
					keys[i].MemberID = nil
				}
			}
		}

		// Remove members from the list
		filtered := make([]types.Member, 0, len(members)-len(ids))
		for _, m := range members {
			if _, ok := idSet[m.ID]; !ok {
				filtered = append(filtered, m)
			}
		}

		if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
			return err
		}
		if err := st.Org().SetMembers(ctx, filtered); err != nil {
			return err
		}
		if err := persistRecalculatedMemberCounts(ctx, st, filtered); err != nil {
			return err
		}
		return core.BumpAuthzRevisionStore(ctx, st)
	})
}

func (s *LocalService) UpdateMemberStatus(ctx context.Context, ids []string, status string) error {
	return s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		keys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			return err
		}
		for _, id := range ids {
			for i := range members {
				if members[i].ID == id {
					members[i].Status = status
					if status == "inactive" {
						for j := range keys {
							if keys[j].MemberID != nil && *keys[j].MemberID == id {
								keys[j].Status = "disabled"
							}
						}
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

func (s *LocalService) TransferMembers(ctx context.Context, ids []string, departmentID string) error {
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

		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		idSet := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			members[i].DepartmentID = departmentID
			members[i].DepartmentName = deptName

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

func (s *LocalService) InviteMember() error {
	return domain.NewDomainError(domain.StatusNotImplemented, "Invite member is not implemented")
}

func (s *LocalService) BatchInvite(ctx context.Context, ids []string) (types.BatchInviteResult, error) {
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.BatchInviteResult{}, err
	}
	targets := make([]types.Member, 0)
	if len(ids) > 0 {
		idSet := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		for _, member := range members {
			if _, ok := idSet[member.ID]; ok {
				targets = append(targets, member)
			}
		}
	} else {
		for _, member := range members {
			if member.Status == types.MemberStatusPending || member.Status == types.MemberStatusInactive {
				targets = append(targets, member)
			}
		}
	}
	return types.BatchInviteResult{Sent: len(targets)}, nil
}

func (s *LocalService) BatchImport(ctx context.Context, rows []types.BatchImportRow) (types.MemberBatchImportResult, error) {
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	flat := pkgorg.FlattenDepartmentTree(departments)
	failures := make([]types.MemberBatchImportFailure, 0)
	imported := 0

	for index, row := range rows {
		var dept *types.Department
		for i := range flat {
			if flat[i].Name == row.DepartmentName {
				dept = &flat[i]
				break
			}
		}
		if dept == nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "types.Department not found",
			})
			continue
		}
		members = append(members, types.Member{
			ID:   generateID("m-import"),
			Name: row.Name, Phone: row.Phone, Email: row.Email,
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Status: types.MemberStatusActive, Roles: []string{permission.RoleMember}, Source: "imported",
		})
		imported++
	}

	if err := s.d.Store.Org().SetMembers(ctx, members); err != nil {
		return types.MemberBatchImportResult{Imported: imported, Failures: append(failures, types.MemberBatchImportFailure{
			Row: 0, Reason: "Failed to persist imported members",
		})}, nil
	}
	if imported > 0 {
		if err := persistRecalculatedMemberCounts(ctx, s.d.Store, members); err != nil {
			return types.MemberBatchImportResult{}, err
		}
		if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
			return types.MemberBatchImportResult{}, err
		}
	}

	return types.MemberBatchImportResult{Imported: imported, Failures: failures}, nil
}
