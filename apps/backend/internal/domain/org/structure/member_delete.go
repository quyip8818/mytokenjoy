package structure

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) DeleteMembers(ctx context.Context, ids []uuid.UUID, currentMemberID uuid.UUID) error {
	for _, id := range ids {
		if id == currentMemberID {
			return domain.BadRequest("不能删除当前登录的用户")
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

		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		// Disable associated keys for all deleted/disabled members.
		for i := range keys {
			if keys[i].MemberID != nil {
				if _, ok := idSet[*keys[i].MemberID]; ok {
					keys[i].Status = "disabled"
					keys[i].MemberID = nil
				}
			}
		}

		// pending → hard delete (physical remove); others → soft delete (set disabled).
		result := make([]types.Member, 0, len(members))
		for _, m := range members {
			if _, ok := idSet[m.ID]; !ok {
				result = append(result, m)
				continue
			}
			if m.Status == types.MemberStatusPending {
				// Hard delete: don't include in result.
				continue
			}
			// Soft delete: mark as disabled, keep record.
			m.Status = types.MemberStatusDisabled
			result = append(result, m)
		}

		if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
			return err
		}
		if err := st.Org().SetMembers(ctx, result); err != nil {
			return err
		}
		if err := persistRecalculatedMemberCounts(ctx, st, result); err != nil {
			return err
		}
		return core.BumpAuthzRevisionStore(ctx, st)
	})
}
