package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid client credentials")

type Client struct {
	ClientID         string
	ClientSecretHash string
	Scope            string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type Store interface {
	GetClientByID(ctx context.Context, clientID string) (*Client, error)
}

type Service struct {
	store     Store
	jwtSecret string
	tokenTTL  time.Duration
}

func NewService(store Store, jwtSecret string, tokenTTL time.Duration) *Service {
	return &Service{store: store, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

func (s *Service) IssueToken(ctx context.Context, clientID, clientSecret string) (TokenResponse, error) {
	client, err := s.store.GetClientByID(ctx, clientID)
	if err != nil {
		return TokenResponse{}, err
	}
	if client == nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   client.ClientID,
		"scope": client.Scope,
		"iat":   now.Unix(),
		"exp":   now.Add(s.tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.tokenTTL.Seconds()),
		Scope:       client.Scope,
	}, nil
}
