package structure

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) BatchInvite(ctx context.Context, ids []uuid.UUID) (types.BatchInviteResult, error) {
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.BatchInviteResult{}, err
	}

	// Filter to pending members only.
	targets := make([]types.Member, 0)
	if len(ids) > 0 {
		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		for _, member := range members {
			if _, ok := idSet[member.ID]; ok && member.Status == types.MemberStatusPending {
				targets = append(targets, member)
			}
		}
	} else {
		for _, member := range members {
			if member.Status == types.MemberStatusPending {
				targets = append(targets, member)
			}
		}
	}

	sent := 0
	for _, target := range targets {
		// Find existing invite for this member.
		invite, err := s.d.Store.Invite().GetInviteByMemberID(ctx, target.ID)
		if err != nil || invite == nil {
			continue
		}

		// Renew expiry.
		newExpiry := time.Now().UTC().Add(s.d.Cfg.InviteExpireDuration())
		if err := s.d.Store.Invite().UpdateInviteExpiry(ctx, invite.ID, newExpiry); err != nil {
			s.d.Logger.Warn("batch invite: renew expiry failed", "memberID", target.ID, "error", err)
			continue
		}

		// Lookup user to get phone/email for re-sending.
		user, err := s.d.Store.User().GetByID(ctx, target.UserID)
		if err != nil || user == nil {
			continue
		}

		// Re-send notifications.
		s.sendInviteNotifications(ctx, invite.InviteCode, newExpiry, user.Phone, user.Email)
		sent++
	}

	return types.BatchInviteResult{Sent: sent}, nil
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

	type importedMember struct {
		inviteCode string
		expiresAt  time.Time
		phone      string
		email      string
	}
	var invitesToSend []importedMember

	companyID := store.CompanyID(ctx)
	now := time.Now().UTC()
	expiresAt := now.Add(s.d.Cfg.InviteExpireDuration())

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
		userID, uerr := s.resolveOrCreateUser(ctx, row.Phone, row.Email, row.Name)
		if uerr != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: uerr.Error(),
			})
			continue
		}

		memberID := generateID()
		members = append(members, types.Member{
			ID: memberID, UserID: userID,
			Alias:        row.Name,
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Status: types.MemberStatusPending, Roles: []string{grants.RoleMember}, Source: types.MemberSourceCSV,
			PersonalBudget: common.DefaultPersonalBudget,
		})

		// Create invite for this member.
		code, cerr := randomHexCode()
		if cerr != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "generate invite code failed",
			})
			continue
		}
		invite := store.CompanyInvite{
			ID:         uuid.Must(uuid.NewV7()),
			CompanyID:  companyID,
			MemberID:   memberID,
			Email:      row.Email,
			Phone:      row.Phone,
			UserID:     userID,
			Role:       "member",
			InviteCode: code,
			ExpiresAt:  expiresAt,
			CreatedAt:  now,
		}
		if err := s.d.Store.Invite().CreateInvite(ctx, invite); err != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "create invite failed",
			})
			continue
		}

		invitesToSend = append(invitesToSend, importedMember{
			inviteCode: code, expiresAt: expiresAt, phone: row.Phone, email: row.Email,
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

	// Best-effort: send invite notifications.
	for _, im := range invitesToSend {
		s.sendInviteNotifications(ctx, im.inviteCode, im.expiresAt, im.phone, im.email)
	}

	return types.MemberBatchImportResult{Imported: imported, Failures: failures}, nil
}
