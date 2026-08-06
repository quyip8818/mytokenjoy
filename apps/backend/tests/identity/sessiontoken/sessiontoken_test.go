package sessiontoken_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
)

// --- issue access token ---

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

// --- issuer ---

func TestIssuedJWTHasNoPermissionsClaim(t *testing.T) {
	issuer, err := sessiontoken.NewIssuer("test-secret", 3600)
	if err != nil {
		t.Fatal(err)
	}
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	memberID := uuid.MustParse("00000000-0000-7000-0000-000000000e01")
	token, err := issuer.Issue(companyID, memberID)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"permissions", "roles", "read_only", "readOnly"} {
		if _, ok := claims[forbidden]; ok {
			t.Fatalf("JWT must not contain %q claim", forbidden)
		}
	}
	if claims["sub"] != memberID.String() {
		t.Fatalf("expected sub %s, got %v", memberID, claims["sub"])
	}
	if claims["company_id"] != companyID.String() {
		t.Fatalf("expected company_id %s, got %v", companyID, claims["company_id"])
	}
}
