package structure

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *LocalService) ListMembers(ctx context.Context, departmentID uuid.UUID, keyword string, directOnly bool, page, pageSize int) (types.MemberPageResult, error) {
	items, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.MemberPageResult{}, err
	}
	if departmentID != uuid.Nil {
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
