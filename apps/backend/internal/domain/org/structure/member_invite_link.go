package structure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/invitetoken"
	"github.com/tokenjoy/backend/internal/store"
)

// GetMemberInviteLink returns an encrypted invite URL for the given member.
// Reuses existing invite record (or creates one if missing). Only works for pending members.
func (s *LocalService) GetMemberInviteLink(ctx context.Context, memberID uuid.UUID) (string, error) {
	target, err := s.d.Store.Org().MemberByID(ctx, memberID)
	if err != nil || target == nil {
		return "", domain.NotFound("member not found")
	}
	if target.Status != types.MemberStatusPending {
		return "", domain.BadRequest("只能为待注册成员获取邀请链接")
	}

	if s.d.InviteIssuer == nil {
		return "", domain.BadRequest("invite secret not configured")
	}

	// Try to find existing invite for this member.
	invite, err := s.d.Store.Invite().GetInviteByMemberID(ctx, memberID)
	if err != nil || invite == nil {
		// Create one if not found.
		code, err := randomHexCode()
		if err != nil {
			return "", err
		}
		now := time.Now().UTC()
		newInvite := store.CompanyInvite{
			ID:         uuid.Must(uuid.NewV7()),
			CompanyID:  store.CompanyID(ctx),
			MemberID:   memberID,
			UserID:     target.UserID,
			Role:       "member",
			InviteCode: code,
			ExpiresAt:  now.Add(s.d.Cfg.InviteExpireDuration()),
			CreatedAt:  now,
		}
		if err := s.d.Store.Invite().CreateInvite(ctx, newInvite); err != nil {
			return "", err
		}
		invite = &newInvite
	}

	// If invite expired, renew it.
	if time.Now().After(invite.ExpiresAt) {
		newExpiry := time.Now().UTC().Add(s.d.Cfg.InviteExpireDuration())
		if err := s.d.Store.Invite().UpdateInviteExpiry(ctx, invite.ID, newExpiry); err != nil {
			return "", err
		}
		invite.ExpiresAt = newExpiry
	}

	// Encrypt with ch=admin_link.
	token, err := s.d.InviteIssuer.Encrypt(invite.InviteCode, invitetoken.ChannelAdminLink, invite.ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("encrypt invite token: %w", err)
	}

	inviteURL := fmt.Sprintf("%s/invite/accept?code=%s", s.d.Cfg.FrontendURL, token)
	return inviteURL, nil
}
