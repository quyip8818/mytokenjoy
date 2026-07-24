package identity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
)

// --- register token ---

func TestRegisterTokenRoundTrip(t *testing.T) {
	t.Parallel()
	secret := []byte("test-secret-32-bytes-for-hmac!!")
	issuer := registertoken.NewIssuer(secret)

	userID := uuid.Must(uuid.NewV7())
	token, err := issuer.Issue(userID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected userID=%s, got %s", userID, claims.UserID)
	}
}

func TestRegisterTokenRejectsWrongSecret(t *testing.T) {
	t.Parallel()
	issuer1 := registertoken.NewIssuer([]byte("secret-1"))
	issuer2 := registertoken.NewIssuer([]byte("secret-2"))

	userID := uuid.Must(uuid.NewV7())
	token, _ := issuer1.Issue(userID)

	_, err := issuer2.Parse(token)
	if err == nil {
		t.Fatal("expected parse failure with wrong secret")
	}
}

func TestRegisterTokenRejectsGarbage(t *testing.T) {
	t.Parallel()
	issuer := registertoken.NewIssuer([]byte("secret"))

	_, err := issuer.Parse("not-a-valid-jwt")
	if err == nil {
		t.Fatal("expected parse failure for garbage input")
	}
}

// --- verify code ---

func TestFormatPhone(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"13800138000", "+8613800138000"},
		{"+8613800138000", "+8613800138000"},
		{"+1234567890", "+1234567890"},
		{"86138", "86138"}, // too short, no prefix added
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := verifycode.FormatPhone(tc.input)
			if got != tc.want {
				t.Errorf("FormatPhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNewServiceNilWhenNoRedisURL(t *testing.T) {
	t.Parallel()
	svc, err := verifycode.NewService(verifycode.Config{RedisURL: ""}, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when redisURL is empty")
	}
}

func TestNewServiceErrorOnBadRedisURL(t *testing.T) {
	t.Parallel()
	svc, err := verifycode.NewService(verifycode.Config{RedisURL: "not-a-url"}, nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid redis URL")
	}
	if svc != nil {
		t.Fatal("expected nil service on error")
	}
}
