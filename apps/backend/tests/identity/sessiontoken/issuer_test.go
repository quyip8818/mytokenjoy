package sessiontoken_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

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
