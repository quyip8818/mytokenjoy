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
	companyCtx := WithContext(ctx, Context{
		CompanyID:          company.ID,
		Slug:               company.Slug,
		NewAPIWalletUserID: newAPIWalletUserIDValue(company),
		Status:             company.Status,
	})
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
	member := types.Member{
		ID:             memberID,
		CompanyID:      company.ID,
		Name:           req.Name,
		Email:          invite.Email,
		DepartmentID:   deptID,
		Status:         "active",
		Roles:          memberRolesFromInvite(invite.Role),
		PersonalBudget: common.DefaultPersonalBudget,
	}
	members, err := s.store.Org().Members(companyCtx)
	if err != nil {
		return types.Member{}, err
	}
	members = append(members, member)
	if err := s.store.Org().SetMembers(companyCtx, members); err != nil {
		return types.Member{}, err
	}
	if err := s.store.Org().SetMemberPasswordHash(companyCtx, memberID, string(passwordHash)); err != nil {
		return types.Member{}, err
	}
	for i := range nodes {
		if nodes[i].ID == deptID {
			nodes[i].ManagerID = &memberID
		}
	}
	if err := s.store.Org().Nodes().SetTree(companyCtx, nodes); err != nil {
		return types.Member{}, err
	}
	if err := s.store.Invite().MarkInviteAccepted(ctx, invite.ID, time.Now().UTC()); err != nil {
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
