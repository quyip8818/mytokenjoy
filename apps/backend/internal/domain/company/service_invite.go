package company

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"golang.org/x/crypto/bcrypt"
)

const minInvitePasswordLen = 8

func (s *service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error) {
	invite, err := s.store.Invite().GetInviteByToken(ctx, req.Token)
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
	if err != nil || company == nil {
		return types.Member{}, domain.NotFound("company not found")
	}
	companyCtx := WithContext(ctx, Context{
		CompanyID:          company.ID,
		Slug:               company.Slug,
		NewAPIWalletUserID: newAPIWalletUserIDValue(company),
		Status:             company.Status,
	})
	departments, err := s.store.Org().Departments(companyCtx)
	if err != nil {
		return types.Member{}, err
	}
	deptID := fmt.Sprintf("dept-root-%d", company.ID)
	if company.RootDeptID != nil {
		deptID = *company.RootDeptID
	}
	memberID := fmt.Sprintf("member-%d-%d", company.ID, time.Now().UnixNano())
	member := types.Member{
		ID:            memberID,
		CompanyID:     company.ID,
		Name:          req.Name,
		Email:         invite.Email,
		DepartmentID:  deptID,
		Status:        "active",
		Roles:         []string{"超级管理员"},
		PersonalQuota: common.DefaultPersonalQuota,
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
	for i := range departments {
		if departments[i].ID == deptID {
			departments[i].MemberCount++
			departments[i].ManagerID = &memberID
		}
	}
	if err := s.store.Org().SetDepartments(companyCtx, departments); err != nil {
		return types.Member{}, err
	}
	if err := s.store.Invite().MarkInviteAccepted(ctx, invite.ID, time.Now().UTC()); err != nil {
		return types.Member{}, err
	}
	return member, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func randomPassword() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
