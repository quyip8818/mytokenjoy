package credentials

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	AuthenticateMember(ctx context.Context, companyID int64, email, password string) (types.Member, error)
	AuthenticatePlatform(ctx context.Context, email, password string) (string, error)
	BootstrapPlatformIfNeeded(ctx context.Context) error
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

func (s *service) BootstrapPlatformIfNeeded(ctx context.Context) error {
	count, err := s.store.Platform().CountOperators(ctx)
	if err != nil {
		return err
	}
	if count > 0 || s.cfg.PlatformBootstrapEmail == "" || s.cfg.PlatformBootstrapPassword == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.PlatformBootstrapPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.store.Platform().CreateOperator(ctx, store.PlatformOperator{
		ID:           "platform-op-1",
		Email:        s.cfg.PlatformBootstrapEmail,
		PasswordHash: string(hash),
		Status:       types.MemberStatusActive,
	})
}

func (s *service) AuthenticateMember(ctx context.Context, companyID int64, email, password string) (types.Member, error) {
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

func (s *service) AuthenticatePlatform(ctx context.Context, email, password string) (string, error) {
	op, err := s.store.Platform().GetOperatorByEmail(ctx, email)
	if err != nil || op == nil || op.Status != types.MemberStatusActive {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	if err := verifyPassword(op.PasswordHash, password); err != nil {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	return op.ID, nil
}

func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
