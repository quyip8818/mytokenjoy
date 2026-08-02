package platform

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type registerLocalBody struct {
	Name           string `json:"name"`
	Industry       string `json:"industry"`
	Size           string `json:"size"`
	AdminEmail     string `json:"adminEmail"`
	AdminPassword  string `json:"adminPassword"`
	AdminName      string `json:"adminName"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// RegisterLocal handles POST /api/platform/register-local.
// This is a public endpoint (no session auth) guarded by X-Registration-Secret.
// Steps:
//  1. Create selfhosted Company (idempotent by idempotencyKey)
//  2. Create admin User + Member (super_admin)
//  3. Create NewAPI wallet user + unlimited token (platformKey)
//  4. Issue sync token (cst_ prefix) for Catalog Sync auth
//
// Returns { companyId, walletUserId, platformKey, syncToken }.
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
	if body.AdminEmail == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "adminEmail is required")
		return
	}
	if body.AdminPassword == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "adminPassword is required")
		return
	}
	if body.IdempotencyKey == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "idempotencyKey is required")
		return
	}

	ctx := r.Context()

	// --- Idempotency check ---
	settingsKey := "register_local:" + body.IdempotencyKey
	existing, _ := h.p.SystemSettings.Get(ctx, settingsKey)
	if existing != "" {
		// Already completed — return stored result.
		var result registerLocalResult
		if err := json.Unmarshal([]byte(existing), &result); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, "corrupt idempotency record")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, result, nil)
		return
	}

	// --- Step 1+2: Create User + Company + Member via CompanySvc ---
	// First, create the admin user (or find existing by email).
	userID, err := h.ensureUser(ctx, body)
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, fmt.Errorf("create admin user: %w", err))
		return
	}

	// Create company (provisions NewAPI wallet user + org tree + member).
	createResult, err := h.p.CompanySvc.CreateCompany(ctx, domaincompany.CreateCompanyRequest{
		UserID:      userID,
		Name:        body.Name,
		Industry:    body.Industry,
		Size:        body.Size,
		Type:        store.CompanyTypeSelfhosted,
		MemberAlias: body.Name,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, fmt.Errorf("create company: %w", err))
		return
	}
	companyID := createResult.Company.ID
	walletUserID := int64(0)
	if createResult.Company.NewAPIWalletCompanyID != nil {
		walletUserID = *createResult.Company.NewAPIWalletCompanyID
	}

	// --- Step 3: Create unlimited token on wallet user (platformKey) ---
	tokenResult, err := h.p.AdminPort.CreateToken(ctx, adminport.CreateTokenInput{
		UserID:         walletUserID,
		Name:           body.Name + " platform key",
		UnlimitedQuota: true,
		Group:          h.p.Cfg.PlatformSharedNewAPIGroup,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, fmt.Errorf("create platform token: %w", err))
		return
	}

	// --- Step 4: Issue sync token ---
	syncToken, hash := generateSyncToken()
	now := time.Now().UTC()
	if err := h.p.Companies.UpdateSyncToken(ctx, companyID, hash, now); err != nil {
		httputil.WriteJSON(w, http.StatusInternalServerError, nil, err)
		return
	}

	// --- Persist idempotent result ---
	result := registerLocalResult{
		CompanyID:    companyID.String(),
		WalletUserID: walletUserID,
		PlatformKey:  tokenResult.Key,
		SyncToken:    syncToken,
	}
	resultJSON, _ := json.Marshal(result)
	_ = h.p.SystemSettings.Set(ctx, settingsKey, string(resultJSON))

	httputil.WriteJSON(w, http.StatusCreated, result, nil)
}

type registerLocalResult struct {
	CompanyID    string `json:"companyId"`
	WalletUserID int64  `json:"walletUserId"`
	PlatformKey  string `json:"platformKey"`
	SyncToken    string `json:"syncToken"`
}

// ensureUser creates an admin user or returns the existing one by email.
func (h *Handler) ensureUser(ctx context.Context, body registerLocalBody) (uuid.UUID, error) {
	// Check if user already exists.
	existing, err := h.p.Users.GetByEmail(ctx, body.AdminEmail)
	if err == nil && existing != nil {
		return existing.ID, nil
	}

	// Create new user.
	hash, err := bcrypt.GenerateFromPassword([]byte(body.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}
	name := body.AdminName
	if name == "" {
		name = body.Name // Default: admin user's name is the company name.
	}
	now := time.Now().UTC()
	user := store.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        body.AdminEmail,
		PasswordHash: string(hash),
		Name:         name,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.p.Users.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
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
