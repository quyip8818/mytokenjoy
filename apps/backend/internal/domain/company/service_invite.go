package company

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/secrets"
	"github.com/tokenjoy/backend/internal/store"
)

func memberRolesFromInvite(role string) []string {
	switch role {
	case store.InviteRoleSuperAdmin:
		return []string{grants.RoleSuperAdmin}
	default:
		return []string{grants.RoleMember}
	}
}

func (s *service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error) {
	if req.UserID == uuid.Nil {
		return types.Member{}, domain.BadRequest("user id required")
	}

	invite, err := s.store.Invite().GetInviteByCode(ctx, req.InviteCode)
	if err != nil {
		return types.Member{}, domain.NotFound("invite not found")
	}
	if invite.AcceptedAt != nil {
		return types.Member{}, domain.NewDomainError(400, "invite already accepted")
	}
	if time.Now().After(invite.ExpiresAt) {
		return types.Member{}, domain.NewDomainError(400, "invite expired")
	}

	var member types.Member
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		if invite.MemberID != uuid.Nil {
			// Member-invite: member already exists as pending — activate it.
			m, err := s.activateExistingMember(ctx, tx, invite.MemberID, req.UserID, req.RegistrationChannel)
			if err != nil {
				return err
			}
			member = m
		} else {
			// Company-invite (platform provisioning): member doesn't exist yet — create one.
			m, err := s.addMember(ctx, tx, req.UserID, invite.CompanyID, req.Name, "", invite.Role)
			if err != nil {
				return err
			}
			member = m
		}
		return tx.Invite().MarkInviteAccepted(ctx, invite.ID, time.Now().UTC(), req.AcceptedMeta)
	})
	if err != nil {
		return types.Member{}, err
	}
	return member, nil
}

// activateExistingMember transitions a pending member to active and sets registrationChannel.
func (s *service) activateExistingMember(ctx context.Context, tx store.Store, memberID, userID uuid.UUID, channel string) (types.Member, error) {
	// Find the company this member belongs to.
	companyID, err := tx.Org().FindMemberCompanyID(ctx, memberID)
	if err != nil || companyID == uuid.Nil {
		return types.Member{}, domain.NotFound("member company not found")
	}
	company, err := tx.Company().GetByID(ctx, companyID)
	if err != nil || company == nil {
		return types.Member{}, domain.NotFound("company not found")
	}

	companyCtx := WithContext(ctx, ContextFromStore(*company))

	members, err := tx.Org().Members(companyCtx)
	if err != nil {
		return types.Member{}, err
	}

	var found *types.Member
	for i := range members {
		if members[i].ID == memberID {
			members[i].Status = types.MemberStatusActive
			members[i].UserID = userID
			members[i].RegistrationChannel = channel
			found = &members[i]
			break
		}
	}
	if found == nil {
		return types.Member{}, domain.NotFound("member not found")
	}

	if err := tx.Org().SetMembers(companyCtx, members); err != nil {
		return types.Member{}, err
	}
	return *found, nil
}

func (s *service) PendingInvitesForUser(ctx context.Context, req PendingInvitesForUserRequest) ([]PendingInvite, error) {
	invites, err := s.store.Invite().FindPendingInvitesForUser(ctx, req.Email, req.Phone, req.UserID)
	if err != nil {
		return nil, err
	}
	if len(invites) == 0 {
		return nil, nil
	}

	// Batch fetch company names.
	companyIDs := make([]uuid.UUID, 0, len(invites))
	for _, inv := range invites {
		companyIDs = append(companyIDs, inv.CompanyID)
	}
	companies, err := s.store.Company().GetByIDs(ctx, companyIDs)
	if err != nil {
		return nil, err
	}
	nameByID := make(map[uuid.UUID]string, len(companies))
	for _, co := range companies {
		nameByID[co.ID] = co.Name
	}

	result := make([]PendingInvite, 0, len(invites))
	for _, inv := range invites {
		name, ok := nameByID[inv.CompanyID]
		if !ok {
			continue
		}
		result = append(result, PendingInvite{
			InviteCode:  inv.InviteCode,
			CompanyID:   inv.CompanyID,
			CompanyName: name,
			Role:        inv.Role,
			ExpiresAt:   inv.ExpiresAt,
		})
	}
	return result, nil
}

func randomInviteCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomPassword() string {
	return secrets.RandomHex(10) // 20 hex chars — NewAPI max password length is 20
}
