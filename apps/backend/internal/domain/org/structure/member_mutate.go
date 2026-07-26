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

func (s *LocalService) CreateMember(ctx context.Context, input types.Member) (types.Member, error) {
	departments, err := common.LoadDepartments(ctx, s.d.Store.Org().Nodes())
	if err != nil {
		return types.Member{}, err
	}
	dept := pkgorg.FindDepartment(departments, input.DepartmentID)
	if dept == nil {
		return types.Member{}, domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
	}

	deptName := dept.Name
	if path := pkgorg.GetDeptPath(departments, input.DepartmentID); path != nil {
		deptName = *path
	}

	// Resolve or create user for this member.
	userID, err := s.resolveOrCreateUser(ctx, input.Phone, input.Email, input.Alias)
	if err != nil {
		return types.Member{}, err
	}

	member := types.Member{
		ID:       generateID(),
		UserID:   userID,
		Alias:    input.Alias,
		Username: input.Username, EmployeeID: input.EmployeeID,
		JobTitle: input.JobTitle, HireDate: input.HireDate,
		DepartmentID: input.DepartmentID, DepartmentName: deptName,
		Status: types.MemberStatusPending, Roles: []string{grants.RoleMember}, Source: types.MemberSourceManual,
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
		Email:      input.Email,
		Phone:      input.Phone,
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
	s.sendInviteNotifications(ctx, inviteCode, expiresAt, input.Phone, input.Email)

	return member, nil
}

// sendInviteNotifications encrypts the invite code with channel info and sends SMS/email.
// Failures are logged but not propagated — the member is already created.
func (s *LocalService) sendInviteNotifications(ctx context.Context, inviteCode string, expiresAt time.Time, phone, email string) {
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
		token, err := s.d.InviteIssuer.Encrypt(inviteCode, invitetoken.ChannelSMS, expiresAt)
		if err == nil {
			inviteURL := fmt.Sprintf("%s/invite/accept?code=%s", baseURL, token)
			msg := domainnotification.RenderedMessage{
				Title: "邀请您加入 " + companyName,
				Body:  fmt.Sprintf("您已被邀请加入%s，请点击链接完成注册：%s", companyName, inviteURL),
				Payload: map[string]any{
					"eventType":   "member_invite",
					"inviteUrl":   inviteURL,
					"companyName": companyName,
				},
			}
			if err := s.d.Sender.SendDirect(ctx, "sms", phone, msg); err != nil {
				s.d.Logger.Error("invite sms send failed", "phone", phone, "error", err)
			}
		}
	}

	if email != "" {
		token, err := s.d.InviteIssuer.Encrypt(inviteCode, invitetoken.ChannelEmail, expiresAt)
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

// randomHexCode generates a 32-byte random hex string for invite codes.
func randomHexCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *LocalService) UpdateMember(ctx context.Context, id uuid.UUID, input types.Member) (types.Member, error) {
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
			if input.Username != "" {
				existing.Username = input.Username
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
			if input.DepartmentID != uuid.Nil {
				existing.DepartmentID = input.DepartmentID
				existing.DepartmentName = input.DepartmentName
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

			// Update phone/email on users table if provided.
			if input.Phone != "" {
				if err := s.d.Store.User().UpdatePhone(ctx, existing.UserID, input.Phone); err != nil {
					return types.Member{}, err
				}
			}
			if input.Email != "" {
				if err := s.d.Store.User().UpdateEmail(ctx, existing.UserID, input.Email); err != nil {
					return types.Member{}, err
				}
			}

			return existing, nil
		}
	}
	return types.Member{}, domain.NewDomainError(404, "types.Member not found")
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
			members[i].Status = status
			if status == "inactive" {
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
