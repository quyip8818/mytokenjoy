package identity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
)

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
