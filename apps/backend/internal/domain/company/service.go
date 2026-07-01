package company

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domainplatform "github.com/tokenjoy/backend/internal/domain/platform"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error)
	AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error)
	ListCompanies(ctx context.Context) ([]store.Company, error)
	UpdateCompany(ctx context.Context, id int64, patch UpdateCompanyPatch) error
	ResolveCompanyContext(ctx context.Context, companyID int64) (Context, error)
	ResolveFromMember(ctx context.Context, memberID string) (Context, error)
}

type UpdateCompanyPatch struct {
	Status    *string
	PackageID *string
}

type CreateCompanyRequest struct {
	Slug            string
	Name            string
	SuperAdminEmail string
	PackageID       *string
}

type CreateCompanyResult struct {
	Company     store.Company
	InviteToken string
}

type AcceptInviteRequest struct {
	Token    string
	Name     string
	Password string
}

const minInvitePasswordLen = 8

type service struct {
	cfg    config.Config
	store  store.Store
	client newapi.AdminClient
}

func NewService(cfg config.Config, st store.Store, client newapi.AdminClient) Service {
	return &service{cfg: cfg, store: st, client: client}
}

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
			if err := tx.Company().UpdateWalletAccountID(ctx, company.ID, user.ID); err != nil {
				return err
			}
			walletID := user.ID
			company.NewAPIWalletAccountID = &walletID
		}
		rootDeptID := fmt.Sprintf("dept-root-%d", company.ID)
		rootDept := types.Department{
			ID: rootDeptID, Name: req.Name, MemberCount: 0,
		}
		departments, err := tx.Org().Departments(companyCtx)
		if err != nil {
			return err
		}
		departments = append(departments, rootDept)
		if err := tx.Org().SetDepartments(companyCtx, departments); err != nil {
			return err
		}
		if err := tx.Org().SetRoles(companyCtx, defaultCompanyRoles(company.ID)); err != nil {
			return err
		}
		if err := tx.Company().UpdateRootDeptID(ctx, company.ID, rootDeptID); err != nil {
			return err
		}
		company.RootDeptID = &rootDeptID
		budgetTree, err := tx.Budget().Tree(companyCtx)
		if err != nil {
			return err
		}
		budgetTree = append(budgetTree, types.BudgetNode{
			ID: rootDeptID, Name: req.Name, Budget: 0, Consumed: 0, Period: "monthly",
		})
		if err := tx.Budget().SetTree(companyCtx, budgetTree); err != nil {
			return err
		}
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
	_ = domainplatform.AppendAudit(ctx, s.store, "platform.company.create", "platform", result.Company.Slug,
		fmt.Sprintf("created company %d invite for %s", result.Company.ID, req.SuperAdminEmail))
	return result, nil
}

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
		CompanyID:             company.ID,
		Slug:                  company.Slug,
		NewAPIWalletAccountID: walletIDValue(company),
		Status:                company.Status,
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
		ID:           memberID,
		CompanyID:    company.ID,
		Name:         req.Name,
		Email:        invite.Email,
		DepartmentID: deptID,
		Status:       "active",
		Roles:        []string{"超级管理员"},
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

func (s *service) ListCompanies(ctx context.Context) ([]store.Company, error) {
	return s.store.Company().List(ctx)
}

func (s *service) UpdateCompany(ctx context.Context, id int64, patch UpdateCompanyPatch) error {
	if patch.Status != nil {
		if err := s.store.Company().UpdateStatus(ctx, id, *patch.Status); err != nil {
			return err
		}
	}
	if patch.PackageID != nil {
		if err := s.store.Company().UpdatePackageID(ctx, id, patch.PackageID); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) ResolveFromMember(ctx context.Context, memberID string) (Context, error) {
	companies, err := s.store.Company().List(ctx)
	if err != nil {
		return Context{}, err
	}
	for _, company := range companies {
		companyCtx := WithContext(ctx, Context{CompanyID: company.ID})
		members, err := s.store.Org().Members(companyCtx)
		if err != nil {
			continue
		}
		for _, member := range members {
			if member.ID == memberID {
				return Context{
					CompanyID:             company.ID,
					Slug:                  company.Slug,
					NewAPIWalletAccountID: walletIDValue(&company),
					Status:                company.Status,
				}, nil
			}
		}
	}
	return Context{}, domain.NotFound("member not found")
}

func (s *service) ResolveCompanyContext(ctx context.Context, companyID int64) (Context, error) {
	company, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || company == nil {
		return Context{}, domain.NotFound("company not found")
	}
	return Context{
		CompanyID:             company.ID,
		Slug:                  company.Slug,
		NewAPIWalletAccountID: walletIDValue(company),
		Status:                company.Status,
	}, nil
}

func walletIDValue(t *store.Company) int64 {
	if t.NewAPIWalletAccountID != nil {
		return *t.NewAPIWalletAccountID
	}
	return 0
}

func defaultCompanyRoles(companyID int64) []types.Role {
	prefix := fmt.Sprintf("%d", companyID)
	return []types.Role{
		{ID: "role-1-" + prefix, Name: permission.RoleSuperAdmin, Type: "preset", Permissions: []string{"*"}},
		{ID: "role-2-" + prefix, Name: permission.RoleOrgAdmin, Type: "preset", Permissions: []string{"org:*"}},
		{ID: "role-3-" + prefix, Name: permission.RoleMember, Type: "preset", Permissions: []string{"self:*"}},
		{ID: "role-4-" + prefix, Name: permission.RoleAuditor, Type: "preset", Permissions: []string{"audit:read"}},
		{ID: "role-5-" + prefix, Name: permission.RoleAPICaller, Type: "preset", Permissions: []string{"api:call"}},
	}
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
