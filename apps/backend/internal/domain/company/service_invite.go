package company

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/pkg/secrets"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const minInvitePasswordLen = 8

func memberRolesFromInvite(role string) []string {
	switch role {
	case store.InviteRoleSuperAdmin:
		return []string{grants.RoleSuperAdmin}
	default:
		return []string{grants.RoleMember}
	}
}

func (s *service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error) {
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
	if len(req.Password) < minInvitePasswordLen {
		return types.Member{}, domain.NewDomainError(400, "password too short")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return types.Member{}, fmt.Errorf("hash password: %w", err)
	}
	company, err := s.store.Company().GetByID(ctx, invite.CompanyID)
	if err != nil {
		return types.Member{}, err
	}
	if company == nil {
		return types.Member{}, domain.NotFound("company not found")
	}

	// Find or create user by email — then create member in same transaction.
	user, err := s.store.User().GetByEmail(ctx, invite.Email)
	if err != nil {
		return types.Member{}, err
	}

	companyCtx := WithContext(ctx, ContextFromStore(*company))
	nodes, err := s.store.Org().Nodes().Tree(companyCtx)
	if err != nil {
		return types.Member{}, err
	}
	deptID := rootOrgNodeID(nodes)
	if deptID == "" {
		deptID = fmt.Sprintf("dept-root-%d", company.ID)
	}
	if company.RootDeptID != nil && *company.RootDeptID != "" {
		deptID = *company.RootDeptID
	}
	memberID := fmt.Sprintf("member-%d-%d", company.ID, time.Now().UnixNano())

	var member types.Member
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		// Create or update user within the transaction.
		if user == nil {
			userID := fmt.Sprintf("u-%d", time.Now().UnixNano())
			user = &store.User{
				ID:           userID,
				Email:        invite.Email,
				PasswordHash: string(passwordHash),
				Status:       "active",
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			}
			if err := tx.User().Create(ctx, *user); err != nil {
				return fmt.Errorf("accept-invite create user (id=%s): %w", user.ID, err)
			}
		} else {
			if err := tx.User().UpdatePassword(ctx, user.ID, string(passwordHash)); err != nil {
				return fmt.Errorf("accept-invite update password: %w", err)
			}
		}

		member = types.Member{
			ID:             memberID,
			CompanyID:      company.ID,
			UserID:         user.ID,
			Name:           req.Name,
			Email:          invite.Email,
			Phone:          user.Phone,
			DepartmentID:   deptID,
			Status:         "active",
			Roles:          memberRolesFromInvite(invite.Role),
			PersonalBudget: common.DefaultPersonalBudget,
		}

		txCompanyCtx := WithContext(ctx, ContextFromStore(*company))
		members, err := tx.Org().Members(txCompanyCtx)
		if err != nil {
			return fmt.Errorf("accept-invite load members: %w", err)
		}
		members = append(members, member)
		if err := tx.Org().SetMembers(txCompanyCtx, members); err != nil {
			return fmt.Errorf("accept-invite SetMembers: %w", err)
		}

		for i := range nodes {
			if nodes[i].ID == deptID {
				nodes[i].ManagerID = &memberID
			}
		}
		if err := tx.Org().Nodes().SetTree(txCompanyCtx, nodes); err != nil {
			return fmt.Errorf("accept-invite SetTree: %w", err)
		}
		return tx.Invite().MarkInviteAccepted(ctx, invite.ID, time.Now().UTC())
	})
	if err != nil {
		return types.Member{}, err
	}
	return member, nil
}

func rootOrgNodeID(nodes []types.OrgNode) string {
	for _, node := range org.FlattenOrgNodeTree(nodes) {
		if node.ParentID == nil || *node.ParentID == "" {
			return node.ID
		}
	}
	return ""
}

func randomInviteCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomPassword() string {
	return secrets.RandomHex(16)
}
