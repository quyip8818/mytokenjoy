package sessiontoken_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

func TestIssuedJWTHasNoPermissionsClaim(t *testing.T) {
	issuer, err := sessiontoken.NewIssuer("test-secret", 3600)
	if err != nil {
		t.Fatal(err)
	}
	token, err := issuer.Issue(1, "m-admin")
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
	if claims["sub"] != "m-admin" {
		t.Fatalf("expected sub m-admin, got %v", claims["sub"])
	}
	if claims["company_id"] != float64(1) {
		t.Fatalf("expected company_id 1, got %v", claims["company_id"])
	}
}
