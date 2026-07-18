package sessiontoken_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

func TestIssueAccessToken_RoundTrip(t *testing.T) {
	secret := []byte("test-secret-key-for-access-token")
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	memberID := uuid.MustParse("00000000-0000-7000-0000-000000000e01")
	userID := uuid.MustParse("00000000-0000-7000-0000-000000000f01")
	sid := sessiontoken.NewSessionID()

	token, err := sessiontoken.IssueAccessToken(secret, 15*time.Minute, companyID, memberID, userID, sid)
	if err != nil {
		t.Fatalf("IssueAccessToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse back using the same secret
	issuer, err := sessiontoken.NewIssuer(string(secret), 900)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}
	claims, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.CompanyID != companyID {
		t.Errorf("CompanyID = %v, want %v", claims.CompanyID, companyID)
	}
	if claims.Subject != memberID.String() {
		t.Errorf("Subject = %v, want %v", claims.Subject, memberID)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Sid != sid {
		t.Errorf("Sid = %v, want %v", claims.Sid, sid)
	}
}

func TestNewSessionID_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		id := sessiontoken.NewSessionID()
		if id == "" {
			t.Fatal("empty session ID")
		}
		if seen[id] {
			t.Fatalf("duplicate session ID: %s", id)
		}
		seen[id] = true
	}
}
