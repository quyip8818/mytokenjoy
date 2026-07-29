package platform

import (
	"crypto/subtle"
	"encoding/json"
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

// RegisterLocal handles POST /api/platform/register-local.
// This is a public endpoint (no session auth) guarded by X-Registration-Secret.
// It creates a selfhosted company record and returns the company ID.
// Idempotent: same idempotencyKey returns the same companyId.
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

	// Idempotency: check system_settings for existing registration with this key
	settingsKey := "register_local:" + body.IdempotencyKey
	existing, _ := h.p.SystemSettings.Get(ctx, settingsKey)
	if existing != "" {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"companyId": existing}, nil)
		return
	}

	// Create selfhosted company (minimal: just the companies row, no wallet/org/roles —
	// those are created by the local instance's own bootstrap)
	companyID := uuid.Must(uuid.NewV7())
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

	// Store idempotency mapping (best-effort; failure doesn't roll back company creation)
	_ = h.p.SystemSettings.Set(ctx, settingsKey, companyID.String())

	httputil.WriteJSON(w, http.StatusCreated, map[string]string{"companyId": companyID.String()}, nil)
}
