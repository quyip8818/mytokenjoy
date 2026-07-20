package credentials

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	AuthenticateMember(ctx context.Context, companyID uuid.UUID, email, password string) (types.Member, error)
	BootstrapPlatformIfNeeded(ctx context.Context) error
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

// BootstrapPlatformIfNeeded creates the first platform admin as a member of the
// super company (TokenJoyCompanyID). Idempotent: skips if the member already exists.
func (s *service) BootstrapPlatformIfNeeded(ctx context.Context) error {
	if s.cfg.PlatformBootstrapEmail == "" || s.cfg.PlatformBootstrapPassword == "" {
		return nil
	}
	// Idempotent check: already bootstrapped?
	existing, _, _ := s.store.Org().MemberByEmail(ctx, s.cfg.TokenJoyCompanyID, s.cfg.PlatformBootstrapEmail)
	if existing != nil {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.PlatformBootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bootstrap platform: hash password: %w", err)
	}

	now := time.Now().UTC()
	userID := uuid.MustParse("90000000-0000-0000-0000-000000000001")
	memberID := uuid.MustParse("90000000-0000-0000-0000-000000000002")

	// Create user (idempotent via GetByEmail check)
	existingUser, _ := s.store.User().GetByEmail(ctx, s.cfg.PlatformBootstrapEmail)
	if existingUser != nil {
		userID = existingUser.ID
	} else {
		if err := s.store.User().Create(ctx, store.User{
			ID:           userID,
			Email:        s.cfg.PlatformBootstrapEmail,
			PasswordHash: string(hash),
			Status:       types.MemberStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		}); err != nil {
			return fmt.Errorf("bootstrap platform: create user: %w", err)
		}
	}

	// PlatformAdmin is a global preset role (seeded by bootstrap), no need to insert it.
	// Only need to create the member with the role assignment.
	companyCtx := ctxcompany.With(ctx, ctxcompany.Info{CompanyID: s.cfg.TokenJoyCompanyID})

	member := types.Member{
		ID:        memberID,
		CompanyID: s.cfg.TokenJoyCompanyID,
		UserID:    userID,
		Name:      s.cfg.PlatformBootstrapEmail,
		Status:    types.MemberStatusActive,
		Roles:     []string{grants.RolePlatformAdmin},
	}
	if err := s.store.Org().SetMembers(companyCtx, []types.Member{member}); err != nil {
		return fmt.Errorf("bootstrap platform: set members: %w", err)
	}
	return nil
}

func (s *service) AuthenticateMember(ctx context.Context, companyID uuid.UUID, email, password string) (types.Member, error) {
	member, hash, err := s.store.Org().MemberByEmail(ctx, companyID, email)
	if err != nil {
		return types.Member{}, err
	}
	if member == nil || hash == "" {
		return types.Member{}, domain.NewDomainError(401, "Invalid credentials")
	}
	if member.Status != types.MemberStatusActive {
		return types.Member{}, domain.NewDomainError(401, "Invalid credentials")
	}
	if err := verifyPassword(hash, password); err != nil {
		return types.Member{}, domain.NewDomainError(401, "Invalid credentials")
	}
	return *member, nil
}

func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
