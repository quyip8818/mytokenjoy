package structure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/pkg/invitetoken"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) CreateMember(ctx context.Context, input types.CreateMemberInput) (types.Member, error) {
	if input.User.Phone == "" && input.User.Email == "" {
		return types.Member{}, domain.NewDomainError(domain.StatusBadRequest, "phone or email is required")
	}

	departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
	if err != nil {
		return types.Member{}, err
	}
	dept := pkgorg.FindDepartment(departments, input.Member.DepartmentID)
	if dept == nil {
		return types.Member{}, domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
	}

	deptName := dept.Name
	if path := pkgorg.GetDeptPath(departments, input.Member.DepartmentID); path != nil {
		deptName = *path
	}

	// Resolve or create user for this member.
	userID, err := s.resolveOrCreateUser(ctx, input.User.Phone, input.User.Email, input.User.Name)
	if err != nil {
		return types.Member{}, err
	}

	alias := input.Member.Alias
	if alias == "" {
		alias = input.User.Name
	}

	member := types.Member{
		ID:             generateID(),
		UserID:         userID,
		Alias:          alias,
		EmployeeID:     input.Member.EmployeeID,
		JobTitle:       input.Member.JobTitle,
		HireDate:       input.Member.HireDate,
		DepartmentID:   input.Member.DepartmentID,
		DepartmentName: deptName,
		Status:         types.MemberStatusPending,
		Roles:          []string{grants.RoleMember},
		Source:         types.MemberSourceManual,
		PersonalBudget: common.DefaultPersonalBudget,
	}

	// Generate invite code.
	inviteCode, err := randomHexCode()
	if err != nil {
		return types.Member{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.d.Cfg.InviteExpireDuration())
	invite := store.CompanyInvite{
		ID:         uuid.Must(uuid.NewV7()),
		CompanyID:  store.CompanyID(ctx),
		MemberID:   member.ID,
		Email:      input.User.Email,
		Phone:      input.User.Phone,
		UserID:     userID,
		Role:       "member",
		InviteCode: inviteCode,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
	}

	err = s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		if err := s.checkTrialMemberLimit(ctx, members); err != nil {
			return err
		}
		members = append(members, member)
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		if err := st.Invite().CreateInvite(ctx, invite); err != nil {
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

	// Best-effort: send invite notifications after commit.
	s.sendInviteNotifications(ctx, inviteCode, input.User.Phone, input.User.Email)

	return member, nil
}

// sendInviteNotifications encrypts the invite code with channel info and sends SMS/email.
// Failures are logged but not propagated — the member is already created.
func (s *LocalService) sendInviteNotifications(ctx context.Context, inviteCode string, phone, email string) {
	if s.d.InviteIssuer == nil || s.d.Sender == nil {
		return
	}
	baseURL := s.d.Cfg.FrontendURL
	companyName := ""
	if info, ok := ctxcompany.From(ctx); ok {
		companyName = info.Name
	}
	if companyName == "" {
		companyName = "TokenJoy"
	}

	if phone != "" {
		token, err := s.d.InviteIssuer.Encrypt(inviteCode, invitetoken.ChannelSMS)
		if err == nil {
			inviteURL := fmt.Sprintf("%s/invite/accept?code=%s", baseURL, token)
			msg := domainnotification.RenderedMessage{
				Title: "邀请您加入 " + companyName,
				Body:  fmt.Sprintf("%s邀请您加入TokenJoy平台管理AI用量，点击加入：%s", companyName, inviteURL),
				Payload: map[string]any{
					"eventType": "member_invite",
					"company":   companyName,
					"code":      token,
				},
			}
			if err := s.d.Sender.SendDirect(ctx, "sms", phone, msg); err != nil {
				s.d.Logger.Error("invite sms send failed", "phone", phone, "error", err)
			}
		}
	}

	if email != "" {
		token, err := s.d.InviteIssuer.Encrypt(inviteCode, invitetoken.ChannelEmail)
		if err == nil {
			inviteURL := fmt.Sprintf("%s/invite/accept?code=%s", baseURL, token)
			msg := domainnotification.RenderedMessage{
				Title: "邀请您加入 " + companyName,
				Body:  fmt.Sprintf("您已被邀请加入%s，请点击链接完成注册。", companyName),
				Payload: map[string]any{
					"eventType":   "member_invite",
					"inviteUrl":   inviteURL,
					"companyName": companyName,
				},
			}
			if err := s.d.Sender.SendDirect(ctx, "email", email, msg); err != nil {
				s.d.Logger.Error("invite email send failed", "email", email, "error", err)
			}
		}
	}
}

// randomHexCode generates an 8-byte random hex string for invite codes (16 chars).
func randomHexCode() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *LocalService) UpdateMember(ctx context.Context, id uuid.UUID, input types.UpdateMemberInput) (types.Member, error) {
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
			// Merge: only overwrite non-zero fields from input.
			// Track user-owned field changes in OverrideFields.
			if input.Alias != "" && input.Alias != existing.Alias {
				existing.OverrideFields = core.TrackOverride(existing.OverrideFields, "alias")
				existing.Alias = input.Alias
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
			if input.DepartmentID != uuid.Nil && input.DepartmentID != existing.DepartmentID {
				departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
				if err != nil {
					return types.Member{}, err
				}
				dept := pkgorg.FindDepartment(departments, input.DepartmentID)
				if dept == nil {
					return types.Member{}, domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
				}
				existing.DepartmentID = input.DepartmentID
				deptName := dept.Name
				if path := pkgorg.GetDeptPath(departments, input.DepartmentID); path != nil {
					deptName = *path
				}
				existing.DepartmentName = deptName
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

func (s *LocalService) UpdateMemberUser(ctx context.Context, memberID uuid.UUID, input types.UpdateMemberUserInput) error {
	member, err := s.d.Store.Org().MemberByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member == nil {
		return domain.NewDomainError(404, "types.Member not found")
	}

	if input.Name != "" {
		if err := s.d.Store.User().UpdateName(ctx, member.UserID, input.Name); err != nil {
			return err
		}
	}
	if input.Phone != "" {
		if err := s.d.Store.User().UpdatePhone(ctx, member.UserID, input.Phone); err != nil {
			return mapMemberUniqueError(err)
		}
	}
	if input.Email != "" {
		if err := s.d.Store.User().UpdateEmail(ctx, member.UserID, input.Email); err != nil {
			return mapMemberUniqueError(err)
		}
	}
	return nil
}

func (s *LocalService) UpdateMemberStatus(ctx context.Context, ids []uuid.UUID, status string) error {
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
		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			// pending 成员不允许手动激活，必须通过注册流程
			if members[i].Status == types.MemberStatusPending && status == types.MemberStatusActive {
				return domain.BadRequest("待激活用户需完成注册后自动激活，不可手动启用")
			}
			members[i].Status = status
			if status == types.MemberStatusDisabled {
				for j := range keys {
					if keys[j].MemberID != nil && *keys[j].MemberID == members[i].ID {
						keys[j].Status = "disabled"
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

func (s *LocalService) TransferMembers(ctx context.Context, ids []uuid.UUID, departmentID uuid.UUID) error {
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

		// Load target department's member_avg_budget for personal budget reassignment
		targetBudgetRow, found, err := st.Budget().OrgNodeBudget().Get(ctx, departmentID)
		if err != nil {
			return err
		}
		targetAvgBudget := float64(0)
		if found && targetBudgetRow.MemberAvgBudget > 0 {
			targetAvgBudget = targetBudgetRow.MemberAvgBudget
		}

		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			members[i].DepartmentID = departmentID
			members[i].DepartmentName = deptName

			// Update personal budget to target department's average
			if targetAvgBudget > 0 {
				members[i].PersonalBudget = targetAvgBudget
			}

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
