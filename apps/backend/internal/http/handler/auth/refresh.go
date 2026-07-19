package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

// Refresh validates the refresh cookie and issues a new access token.
// No DB write — SELECT + JWT sign only.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	raw := httpx.ResolveRefreshCookie(r)
	if raw == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sid, _, ok := splitRefreshToken(raw)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sess, err := h.sessions.GetActive(r.Context(), sid)
	if err != nil || sess == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if sessiontoken.SHA256Hex(raw) != sess.TokenHash {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ttl := time.Duration(h.pub.Cfg.SessionTTLSec) * time.Second
	accessToken, err := sessiontoken.IssueAccessToken(
		h.pub.SessionToken.Secret(), ttl,
		sess.CompanyID, sess.MemberID, sess.UserID, sess.ID,
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	httpx.SetSessionCookie(w, accessToken, h.pub.SecureCookie)
	w.WriteHeader(http.StatusNoContent)
}

// issueTokenPair creates a session record and sets both access + refresh cookies.
func (h *Handler) issueTokenPair(w http.ResponseWriter, r *http.Request, companyID, memberID, userID uuid.UUID) (string, error) {
	return httpx.IssueTokenPair(r.Context(), w, r, httpx.TokenPairParams{
		Secret:        h.pub.SessionToken.Secret(),
		SessionTTLSec: h.pub.Cfg.SessionTTLSec,
		RefreshTTLSec: h.pub.Cfg.RefreshTokenTTLSec,
		SecureCookie:  h.pub.SecureCookie,
		SessionStore:  h.sessions,
	}, companyID, memberID, userID)
}

// splitRefreshToken parses "sid.random" into (sid, random, ok).
func splitRefreshToken(raw string) (string, string, bool) {
	idx := strings.IndexByte(raw, '.')
	if idx <= 0 || idx >= len(raw)-1 {
		return "", "", false
	}
	return raw[:idx], raw[idx+1:], true
}
