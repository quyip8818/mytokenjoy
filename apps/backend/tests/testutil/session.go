package testutil

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/seed/contract"
)

func SessionIssuer(t *testing.T) sessiontoken.Issuer {
	t.Helper()
	issuer, err := sessiontoken.NewIssuer(TestSessionSecret, 86400)
	if err != nil {
		t.Fatalf("session issuer: %v", err)
	}
	return issuer
}

func IssueSessionJWT(t *testing.T, companyID uuid.UUID, memberID uuid.UUID) string {
	t.Helper()
	token, err := SessionIssuer(t).Issue(companyID, memberID)
	if err != nil {
		t.Fatalf("issue session jwt: %v", err)
	}
	return token
}

func SessionCookie(t *testing.T, memberID uuid.UUID) string {
	t.Helper()
	return SessionCookieForCompany(t, contract.DefaultCompanyID, memberID)
}

func SessionCookieForCompany(t *testing.T, companyID uuid.UUID, memberID uuid.UUID) string {
	t.Helper()
	token := IssueSessionJWT(t, companyID, memberID)
	return httpx.SessionCookie + "=" + token
}

func SessionCookieAdmin(t *testing.T) string {
	t.Helper()
	return SessionCookie(t, contract.IDMemberAdmin)
}

func WithSessionConfig(cfg config.Config) config.Config {
	cfg.SessionSecret = TestSessionSecret
	if cfg.SessionTTLSec == 0 {
		cfg.SessionTTLSec = 900
	}
	return cfg
}

func SetSessionAuth(req *http.Request, t *testing.T, memberID uuid.UUID) {
	t.Helper()
	req.Header.Set("Cookie", SessionCookie(t, memberID))
}
