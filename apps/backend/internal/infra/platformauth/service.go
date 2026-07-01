package platformauth

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Authenticate(ctx context.Context, email, password string) (string, error)
	BootstrapIfNeeded(ctx context.Context) error
}

type service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{cfg: cfg, store: st}
}

func (s *service) BootstrapIfNeeded(ctx context.Context) error {
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
		Status:       "active",
	})
}

func (s *service) Authenticate(ctx context.Context, email, password string) (string, error) {
	op, err := s.store.Platform().GetOperatorByEmail(ctx, email)
	if err != nil || op == nil || op.Status != "active" {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(op.PasswordHash), []byte(password)); err != nil {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	return op.ID, nil
}
