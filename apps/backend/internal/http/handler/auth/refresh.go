package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/store"
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

	sess, err := h.store.Session().GetActive(r.Context(), sid)
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
	sid := sessiontoken.NewSessionID()
	refreshRaw := sid + "." + sessiontoken.RandomHex(32)

	ttl := time.Duration(h.pub.Cfg.SessionTTLSec) * time.Second
	accessToken, err := sessiontoken.IssueAccessToken(
		h.pub.SessionToken.Secret(), ttl,
		companyID, memberID, userID, sid,
	)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	refreshTTL := time.Duration(h.pub.Cfg.RefreshTokenTTLSec) * time.Second
	sess := store.Session{
		ID:        sid,
		UserID:    userID,
		MemberID:  memberID,
		CompanyID: companyID,
		TokenHash: sessiontoken.SHA256Hex(refreshRaw),
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		CreatedAt: now,
		ExpiresAt: now.Add(refreshTTL),
	}
	if err := h.store.Session().Create(r.Context(), sess); err != nil {
		return "", err
	}

	httpx.SetSessionCookie(w, accessToken, h.pub.SecureCookie)
	httpx.SetRefreshCookie(w, refreshRaw, h.pub.SecureCookie, h.pub.Cfg.RefreshTokenTTLSec)
	return accessToken, nil
}

// splitRefreshToken parses "sid.random" into (sid, random, ok).
func splitRefreshToken(raw string) (string, string, bool) {
	idx := strings.IndexByte(raw, '.')
	if idx <= 0 || idx >= len(raw)-1 {
		return "", "", false
	}
	return raw[:idx], raw[idx+1:], true
}
