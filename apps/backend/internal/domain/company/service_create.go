package company

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error) {
	companies, err := s.store.Company().List(ctx)
	if err != nil {
		return CreateCompanyResult{}, err
	}
	var nextID int64 = 1
	for _, t := range companies {
		if t.ID >= nextID {
			nextID = t.ID + 1
		}
	}
	now := time.Now().UTC()
	var result CreateCompanyResult
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		company := store.Company{
			ID:        nextID,
			Slug:      req.Slug,
			Name:      req.Name,
			Status:    store.CompanyStatusActive,
			PackageID: req.PackageID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Company().Create(ctx, company); err != nil {
			return err
		}
		companyCtx := WithContext(ctx, Context{CompanyID: company.ID, Slug: company.Slug, Status: company.Status})
		if s.cfg.NewAPIEnabled && s.client != nil {
			user, err := s.client.CreateUser(ctx, newapi.CreateUserRequest{
				Username:    fmt.Sprintf("company-%d", company.ID),
				DisplayName: req.Name,
				Password:    randomPassword(),
				Quota:       0,
			})
			if err != nil {
				return fmt.Errorf("create newapi wallet account: %w", err)
			}
			if err := tx.Company().UpdateNewAPIWalletUserID(ctx, company.ID, user.ID); err != nil {
				return err
			}
			walletID := user.ID
			company.NewAPIWalletUserID = &walletID
		}
		rootDeptID := fmt.Sprintf("dept-root-%d", company.ID)
		nodes, err := tx.Org().Nodes().Tree(companyCtx)
		if err != nil {
			return err
		}
		nodes = append(nodes, types.OrgNode{
			ID: rootDeptID, Name: req.Name, MemberCount: 0,
			Budget: 0, Consumed: 0, Period: pkgbudget.PeriodMonthly,
		})
		if err := tx.Org().Nodes().SetTree(companyCtx, nodes); err != nil {
			return err
		}
		if err := tx.Org().SetRoles(companyCtx, defaultCompanyRoles(company.ID)); err != nil {
			return err
		}
		if err := tx.Company().UpdateRootDeptID(ctx, company.ID, rootDeptID); err != nil {
			return err
		}
		company.RootDeptID = &rootDeptID
		inviteToken, err := randomToken()
		if err != nil {
			return err
		}
		inviteID := fmt.Sprintf("invite-%d-%d", company.ID, time.Now().UnixNano())
		if err := tx.Invite().CreateInvite(ctx, store.CompanyInvite{
			ID:        inviteID,
			CompanyID: company.ID,
			Email:     req.SuperAdminEmail,
			Role:      store.InviteRoleSuperAdmin,
			Token:     inviteToken,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
			CreatedAt: now,
		}); err != nil {
			return err
		}
		result = CreateCompanyResult{Company: company, InviteToken: inviteToken}
		return nil
	})
	if err != nil {
		return CreateCompanyResult{}, err
	}
	_ = AppendPlatformOperationLog(ctx, s.store, result.Company.ID, "platform.company.create", "platform", result.Company.Slug,
		fmt.Sprintf("created company %d invite for %s", result.Company.ID, req.SuperAdminEmail))
	return result, nil
}

func defaultCompanyRoles(companyID int64) []types.Role {
	prefix := fmt.Sprintf("%d", companyID)
	roles := []types.Role{
		{ID: "role-1-" + prefix, Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}},
		{ID: "role-2-" + prefix, Name: permission.RoleOrgAdmin, Type: "preset", Permissions: []string{"org:*"}},
		{ID: "role-3-" + prefix, Name: permission.RoleMember, Type: "preset", Permissions: []string{"self:*"}},
		{ID: "role-4-" + prefix, Name: permission.RoleAuditor, Type: "preset", Permissions: []string{"audit:read"}},
		{ID: "role-5-" + prefix, Name: permission.RoleAPICaller, Type: "preset", Permissions: []string{"api:call"}},
	}
	for i := range roles {
		ids, err := permission.RoleGrantIDs(roles[i].Type, roles[i].Name, roles[i].Permissions)
		if err != nil {
			panic(err)
		}
		roles[i].Permissions = ids
	}
	return roles
}
