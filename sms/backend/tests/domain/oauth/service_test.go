package oauth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sms/backend/internal/domain/oauth"
)

// --- mock store ---

type mockClientStore struct {
	clients map[string]*oauth.Client
}

func newMockStore() *mockClientStore {
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-secret"), bcrypt.MinCost)
	return &mockClientStore{
		clients: map[string]*oauth.Client{
			"tokenjoy-sync": {
				ClientID:         "tokenjoy-sync",
				ClientSecretHash: string(hash),
				Scope:            "sync:read",
			},
		},
	}
}

func (m *mockClientStore) GetClientByID(_ context.Context, clientID string) (*oauth.Client, error) {
	c, ok := m.clients[clientID]
	if !ok {
		return nil, nil
	}
	return c, nil
}

// --- tests ---

func TestIssueToken_ValidCredentials(t *testing.T) {
	store := newMockStore()
	svc := oauth.NewService(store, "test-jwt-secret", 10*time.Minute)

	resp, err := svc.IssueToken(context.Background(), "tokenjoy-sync", "test-secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %s", resp.TokenType)
	}
	if resp.ExpiresIn != 600 {
		t.Fatalf("expected expires_in 600, got %d", resp.ExpiresIn)
	}
	if resp.Scope != "sync:read" {
		t.Fatalf("expected scope sync:read, got %s", resp.Scope)
	}

	// Verify the JWT is valid and contains expected claims
	token, err := jwt.Parse(resp.AccessToken, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-jwt-secret"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims")
	}
	if claims["sub"] != "tokenjoy-sync" {
		t.Fatalf("expected sub=tokenjoy-sync, got %v", claims["sub"])
	}
	if claims["scope"] != "sync:read" {
		t.Fatalf("expected scope=sync:read, got %v", claims["scope"])
	}
}

func TestIssueToken_WrongSecret(t *testing.T) {
	store := newMockStore()
	svc := oauth.NewService(store, "test-jwt-secret", 10*time.Minute)

	_, err := svc.IssueToken(context.Background(), "tokenjoy-sync", "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestIssueToken_UnknownClient(t *testing.T) {
	store := newMockStore()
	svc := oauth.NewService(store, "test-jwt-secret", 10*time.Minute)

	_, err := svc.IssueToken(context.Background(), "unknown-client", "any-secret")
	if err == nil {
		t.Fatal("expected error for unknown client")
	}
}
