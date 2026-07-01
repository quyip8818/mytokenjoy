package middleware

import (
	"context"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const platformSessionCookie = "tokenjoy_platform_session"

type PlatformService interface {
	Authenticate(ctx context.Context, email, password string) (string, error)
	BootstrapIfNeeded(ctx context.Context) error
	GetOperatorID(r *http.Request) (string, bool)
}

type platformContextKey struct{}

func WithPlatformOperator(ctx context.Context, operatorID string) context.Context {
	return context.WithValue(ctx, platformContextKey{}, operatorID)
}

func PlatformOperatorFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(platformContextKey{}).(string)
	return id, ok
}

func PlatformAuth(cfg config.Config, svc PlatformService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/platform/auth/login" && r.Method == http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			operatorID := common.ResolvePlatformOperatorID(r)
			if operatorID == "" {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			ctx := WithPlatformOperator(r.Context(), operatorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type platformAuthService struct {
	cfg   config.Config
	store store.Store
}

func NewPlatformAuthService(cfg config.Config, st store.Store) PlatformService {
	return &platformAuthService{cfg: cfg, store: st}
}

func (s *platformAuthService) BootstrapIfNeeded(ctx context.Context) error {
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

func (s *platformAuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	op, err := s.store.Platform().GetOperatorByEmail(ctx, email)
	if err != nil || op == nil || op.Status != "active" {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(op.PasswordHash), []byte(password)); err != nil {
		return "", domain.NewDomainError(401, "Invalid credentials")
	}
	return op.ID, nil
}

func (s *platformAuthService) GetOperatorID(r *http.Request) (string, bool) {
	id := common.ResolvePlatformOperatorID(r)
	return id, id != ""
}
