package platform

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/store"
)

type registerLocalBody struct {
	Name           string `json:"name"`
	Industry       string `json:"industry"`
	Size           string `json:"size"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// tokenIssueWindow is the minimum interval between token issuances to prevent
// network-retry races from overwriting a just-issued token.
const tokenIssueWindow = 60 * time.Second

// RegisterLocal handles POST /api/platform/register-local.
// This is a public endpoint (no session auth) guarded by X-Registration-Secret.
// It creates a selfhosted company record (idempotent by idempotencyKey) and issues
// a per-company sync token (cst_ prefix). Token issuance has a 60s dedup window.
func (h *Handler) RegisterLocal(w http.ResponseWriter, r *http.Request) {
	// Verify registration secret
	secret := h.p.Cfg.LocalRegistrationSecret
	if secret == "" {
		httputil.WriteStatus(w, http.StatusForbidden, "registration not enabled")
		return
	}
	provided := r.Header.Get("X-Registration-Secret")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
		httputil.WriteStatus(w, http.StatusForbidden, "invalid registration secret")
		return
	}

	var body registerLocalBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.IdempotencyKey == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "idempotencyKey is required")
		return
	}

	ctx := r.Context()

	// --- Company creation (idempotent by idempotencyKey) ---
	settingsKey := "register_local:" + body.IdempotencyKey
	var companyID uuid.UUID

	existing, _ := h.p.SystemSettings.Get(ctx, settingsKey)
	if existing != "" {
		// Company already exists — reuse it
		parsed, err := uuid.Parse(existing)
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, "corrupt idempotency record")
			return
		}
		companyID = parsed
	} else {
		// Create new selfhosted company
		companyID = uuid.Must(uuid.NewV7())
		now := time.Now().UTC()
		company := store.Company{
			ID:        companyID,
			Name:      body.Name,
			Industry:  body.Industry,
			Size:      body.Size,
			Type:      store.CompanyTypeSelfhosted,
			Status:    store.CompanyStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.p.Companies.Create(ctx, company); err != nil {
			httputil.WriteJSON(w, http.StatusInternalServerError, nil, err)
			return
		}
		_ = h.p.SystemSettings.Set(ctx, settingsKey, companyID.String())
	}

	// --- Token issuance (60s dedup window) ---
	co, err := h.p.Companies.GetByID(ctx, companyID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, err)
		return
	}
	if co.TokenIssuedAt != nil && time.Since(*co.TokenIssuedAt) < tokenIssueWindow {
		httputil.WriteStatus(w, http.StatusConflict, "token recently issued, use existing token or retry after 60s")
		return
	}

	// Generate cst_ token: "cst_" + 32 random bytes hex = 68 chars
	syncToken, hash := generateSyncToken()
	now := time.Now().UTC()
	if err := h.p.Companies.UpdateSyncToken(ctx, companyID, hash, now); err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, err)
		return
	}

	status := http.StatusOK
	if existing == "" {
		status = http.StatusCreated
	}
	httputil.WriteJSON(w, status, map[string]string{
		"companyId": companyID.String(),
		"syncToken": syncToken,
	}, nil)
}

// generateSyncToken creates a random sync token and its SHA-256 hash.
// Token format: "cst_" + 32 random bytes hex (68 chars total).
func generateSyncToken() (token string, hash string) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	token = "cst_" + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(h[:])
	return token, hash
}
