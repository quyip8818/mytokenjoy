package company

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error) {
	companyType := req.Type
	if companyType == "" {
		companyType = store.CompanyTypeStandard
	}
	now := time.Now().UTC()
	var result CreateCompanyResult
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		company := store.Company{
			ID:        uuid.Must(uuid.NewV7()),
			Name:      req.Name,
			Type:      companyType,
			Status:    store.CompanyStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Company().Create(ctx, company); err != nil {
			return err
		}
		if err := tx.TenantBackgroundState().EnsureRow(ctx, company.ID); err != nil {
			return err
		}
		companyCtx := WithContext(ctx, Context{CompanyID: company.ID, Status: company.Status})
		if s.client == nil {
			return fmt.Errorf("newapi admin client required")
		}
		user, err := s.client.CreateUser(ctx, adminport.CreateUserInput{
			Username:    fmt.Sprintf("company-%s", company.ID),
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
		rootDeptID := uuid.Must(uuid.NewV7())
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
		if err := tx.Org().SetRoles(companyCtx, defaultCompanyRoles(s.grants)); err != nil {
			return err
		}
		if err := tx.Company().UpdateRootDeptID(ctx, company.ID, rootDeptID); err != nil {
			return err
		}
		company.RootDeptID = &rootDeptID
		inviteCode, err := randomInviteCode()
		if err != nil {
			return err
		}
		inviteID := uuid.Must(uuid.NewV7())
		if err := tx.Invite().CreateInvite(ctx, store.CompanyInvite{
			ID:         inviteID,
			CompanyID:  company.ID,
			Email:      req.SuperAdminEmail,
			Role:       store.InviteRoleSuperAdmin,
			InviteCode: inviteCode,
			ExpiresAt:  now.Add(7 * 24 * time.Hour),
			CreatedAt:  now,
		}); err != nil {
			return err
		}
		result = CreateCompanyResult{Company: company, InviteCode: inviteCode}
		return nil
	})
	if err != nil {
		return CreateCompanyResult{}, err
	}
	_ = AppendPlatformOperationLog(ctx, s.store, result.Company.ID, "platform.company.create", uuid.Nil,
		result.Company.ID.String(),
		fmt.Sprintf("created company %s invite for %s", result.Company.ID, req.SuperAdminEmail))
	return result, nil
}

func defaultCompanyRoles(normalizer grants.Normalizer) []types.Role {
	roles := []types.Role{
		{ID: uuid.Must(uuid.NewV7()), Name: grants.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}},
		{ID: uuid.Must(uuid.NewV7()), Name: grants.RoleOrgAdmin, Type: "preset", Permissions: []string{"org:*"}},
		{ID: uuid.Must(uuid.NewV7()), Name: grants.RoleMember, Type: "preset", Permissions: []string{"self:*"}},
		{ID: uuid.Must(uuid.NewV7()), Name: grants.RoleAuditor, Type: "preset", Permissions: []string{"audit:read"}},
		{ID: uuid.Must(uuid.NewV7()), Name: grants.RoleAPICaller, Type: "preset", Permissions: []string{"api:call"}},
	}
	for i := range roles {
		ids, err := normalizer.RoleGrantIDs(roles[i].Type, roles[i].Name, roles[i].Permissions)
		if err != nil {
			panic(err)
		}
		roles[i].Permissions = ids
	}
	return roles
}
