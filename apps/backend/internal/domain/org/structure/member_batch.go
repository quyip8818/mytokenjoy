package structure

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

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
	if err := s.checkTrialMemberLimitBatch(ctx, members, len(rows)); err != nil {
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
			Status: types.MemberStatusActive, Roles: []string{grants.RoleMember}, Source: "imported",
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
