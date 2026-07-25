package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"sms/backend/internal/domain/types"
)

type Store interface {
	GetUserByUsername(ctx context.Context, username string) (*types.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*types.User, error)
	CreateSession(ctx context.Context, session *types.Session) error
	GetSession(ctx context.Context, token string) (*types.Session, error)
	DeleteSession(ctx context.Context, token string) error
	RotateSession(ctx context.Context, oldToken, newToken string, expiresAt time.Time) error
}

type Service struct {
	store           Store
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(store Store, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{store: store, jwtSecret: jwtSecret, accessTokenTTL: accessTTL, refreshTokenTTL: refreshTTL}
}

type LoginResult struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"-"`
	User         types.User `json:"user"`
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, types.ErrUnauthorized
	}
	if user.Status == 0 {
		return nil, types.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, types.ErrUnauthorized
	}

	accessToken := s.issueAccessToken(user)
	refreshToken := uuid.NewString()

	session := &types.Session{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &LoginResult{AccessToken: accessToken, RefreshToken: refreshToken, User: *user}, nil
}

func (s *Service) Refresh(ctx context.Context, oldRefreshToken string) (*LoginResult, error) {
	session, err := s.store.GetSession(ctx, oldRefreshToken)
	if err != nil || session.ExpiresAt.Before(time.Now()) {
		return nil, types.ErrUnauthorized
	}

	user, err := s.store.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, types.ErrUnauthorized
	}

	newRefreshToken := uuid.NewString()
	if err := s.store.RotateSession(ctx, oldRefreshToken, newRefreshToken, time.Now().Add(s.refreshTokenTTL)); err != nil {
		return nil, err
	}

	accessToken := s.issueAccessToken(user)
	user.PasswordHash = ""
	return &LoginResult{AccessToken: accessToken, RefreshToken: newRefreshToken, User: *user}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.store.DeleteSession(ctx, refreshToken)
}

func (s *Service) Profile(ctx context.Context, userID uuid.UUID) (*types.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *Service) JWTSecret() string {
	return s.jwtSecret
}

func (s *Service) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

func (s *Service) issueAccessToken(user *types.User) string {
	claims := jwt.MapClaims{
		"id":       user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(s.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(s.jwtSecret))
	return signed
}
