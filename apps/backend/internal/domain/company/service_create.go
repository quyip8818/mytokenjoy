package company

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error) {
	if req.UserID == uuid.Nil && req.InviteEmail == "" {
		return CreateCompanyResult{}, domain.BadRequest("either userID or inviteEmail is required")
	}
	companyType := req.Type
	if companyType == "" {
		companyType = store.CompanyTypeStandard
	}

	var result CreateCompanyResult
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		company, err := s.provisionCompany(ctx, tx, req.Name, req.Industry, req.Size, companyType)
		if err != nil {
			return err
		}

		if req.UserID != uuid.Nil {
			// UserID mode: creator becomes super-admin immediately.
			alias := req.MemberAlias
			if alias == "" {
				// Fallback: use users.name as alias.
				if user, err := tx.User().GetByID(ctx, req.UserID); err == nil && user != nil {
					alias = user.Name
				}
			}
			member, err := s.addMember(ctx, tx, req.UserID, company.ID, alias, req.MemberAvatar, store.InviteRoleSuperAdmin)
			if err != nil {
				return err
			}
			result = CreateCompanyResult{Company: company, Member: &member}
		} else {
			// InviteEmail mode: generate invite for deferred join.
			inviteCode, err := randomInviteCode()
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			if err := tx.Invite().CreateInvite(ctx, store.CompanyInvite{
				ID:         uuid.Must(uuid.NewV7()),
				CompanyID:  company.ID,
				Email:      req.InviteEmail,
				Role:       store.InviteRoleSuperAdmin,
				InviteCode: inviteCode,
				ExpiresAt:  now.Add(7 * 24 * time.Hour),
				CreatedAt:  now,
			}); err != nil {
				return err
			}
			result = CreateCompanyResult{Company: company, InviteCode: inviteCode}
		}
		return nil
	})
	if err != nil {
		return CreateCompanyResult{}, err
	}

	logDetail := fmt.Sprintf("created company %s", result.Company.ID)
	if req.InviteEmail != "" {
		logDetail = fmt.Sprintf("created company %s invite for %s", result.Company.ID, req.InviteEmail)
		// Send invite email (best-effort, non-fatal).
		s.sendInviteEmail(ctx, req.InviteEmail, result.Company.Name, result.InviteCode)
	}
	_ = AppendPlatformOperationLog(ctx, s.store, result.Company.ID, "platform.company.create", uuid.Nil,
		result.Company.ID.String(), logDetail)
	return result, nil
}

// provisionCompany creates company infrastructure within tx:
// Company row + NewAPI wallet + preset roles + root org node.
func (s *service) provisionCompany(ctx context.Context, tx store.Store, name, industry, size, companyType string) (store.Company, error) {
	now := time.Now().UTC()
	companyID := uuid.Must(uuid.NewV7())
	company := store.Company{
		ID:        companyID,
		Name:      name,
		Industry:  industry,
		Size:      size,
		Type:      companyType,
		Status:    store.CompanyStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := tx.Company().Create(ctx, company); err != nil {
		return store.Company{}, err
	}
	if err := tx.TenantBackgroundState().EnsureRow(ctx, company.ID); err != nil {
		return store.Company{}, err
	}

	companyCtx := WithContext(ctx, Context{CompanyID: company.ID, Status: company.Status})

	if s.client == nil {
		return store.Company{}, fmt.Errorf("newapi admin client required")
	}
	user, err := s.client.CreateUser(ctx, adminport.CreateUserInput{
		Username:    WalletUsername(companyID),
		DisplayName: name,
		Password:    randomPassword(),
	})
	if err != nil {
		return store.Company{}, fmt.Errorf("create newapi wallet account: %w", err)
	}
	if err := tx.Company().UpdateNewAPIWalletCompanyID(ctx, company.ID, user.ID); err != nil {
		return store.Company{}, err
	}
	// Trial/demo accounts: give NewAPI user a large quota for mock model requests.
	if company.Type == store.CompanyTypeTrial || company.Type == store.CompanyTypeDemo {
		_ = s.client.ManageUser(ctx, user.ID, "add_quota", 500000*500000)
	}
	walletID := user.ID
	company.NewAPIWalletCompanyID = &walletID

	rootDeptID := uuid.Must(uuid.NewV7())
	nodes, err := tx.Org().Nodes().Tree(companyCtx)
	if err != nil {
		return store.Company{}, err
	}
	nodes = append(nodes, types.OrgNode{
		ID: rootDeptID, Name: name, MemberCount: 0,
		Budget: 0, Consumed: 0, Period: budget.PeriodMonthly,
	})
	if err := tx.Org().Nodes().SetTree(companyCtx, nodes); err != nil {
		return store.Company{}, err
	}
	if err := tx.Company().UpdateRootDeptID(ctx, company.ID, rootDeptID); err != nil {
		return store.Company{}, err
	}
	company.RootDeptID = &rootDeptID
	return company, nil
}

// addMember adds a user to a company. Idempotent: if user is already a member, returns existing.
func (s *service) addMember(ctx context.Context, tx store.Store, userID, companyID uuid.UUID, alias, avatar, role string) (types.Member, error) {
	company, err := tx.Company().GetByID(ctx, companyID)
	if err != nil {
		return types.Member{}, err
	}
	if company == nil {
		return types.Member{}, domain.NotFound("company not found")
	}

	companyCtx := WithContext(ctx, ContextFromStore(*company))

	// Check if already a member (idempotent).
	members, err := tx.Org().Members(companyCtx)
	if err != nil {
		return types.Member{}, fmt.Errorf("addMember load members: %w", err)
	}
	for _, m := range members {
		if m.UserID == userID {
			return m, nil
		}
	}

	// Determine department.
	deptID := uuid.Nil
	if company.RootDeptID != nil && *company.RootDeptID != uuid.Nil {
		deptID = *company.RootDeptID
	}
	if deptID == uuid.Nil {
		deptID = uuid.Must(uuid.NewV7())
	}

	memberID := uuid.Must(uuid.NewV7())
	member := types.Member{
		ID:             memberID,
		CompanyID:      company.ID,
		UserID:         userID,
		Alias:          alias,
		Avatar:         avatar,
		DepartmentID:   deptID,
		Status:         types.MemberStatusActive,
		Roles:          memberRolesFromInvite(role),
		PersonalBudget: common.DefaultPersonalBudget,
	}

	members = append(members, member)
	if err := tx.Org().SetMembers(companyCtx, members); err != nil {
		return types.Member{}, fmt.Errorf("addMember SetMembers: %w", err)
	}

	// Set manager on root dept for super-admin.
	if role == store.InviteRoleSuperAdmin {
		nodes, err := tx.Org().Nodes().Tree(companyCtx)
		if err != nil {
			return types.Member{}, err
		}
		for i := range nodes {
			if nodes[i].ID == deptID {
				nodes[i].ManagerID = &memberID
			}
		}
		if err := tx.Org().Nodes().SetTree(companyCtx, nodes); err != nil {
			return types.Member{}, err
		}
	}

	return member, nil
}

// sendInviteEmail sends the company invite email. Best-effort: failures are logged, not propagated.
func (s *service) sendInviteEmail(ctx context.Context, email, companyName, inviteCode string) {
	if s.emailSender == nil {
		return
	}
	inviteURL := fmt.Sprintf("%s/invite?code=%s", s.cfg.FrontendURL, inviteCode)
	msg := domainnotification.RenderedMessage{
		Title: fmt.Sprintf("%s 邀请您加入 TokenJoy", companyName),
		Body:  fmt.Sprintf("%s 邀请您加入 TokenJoy 平台，邀请码：%s", companyName, inviteCode),
		Payload: map[string]any{
			"eventType":   "company_invite",
			"companyName": companyName,
			"inviteCode":  inviteCode,
			"inviteUrl":   inviteURL,
		},
	}
	_ = s.emailSender.SendDirect(ctx, "email", email, msg)
}
