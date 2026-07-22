package me

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	shared.ProtectedHandlerBase
	memberAnalytics domainmemberanalytics.Service
	users           store.UserRepository
	sessions        store.SessionRepository
	verifyCode      *verifycode.Service
}

func NewHandler(p httpdeps.Protected, memberAnalytics domainmemberanalytics.Service,
	users store.UserRepository, sessions store.SessionRepository, vc *verifycode.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		memberAnalytics:      memberAnalytics,
		users:                users,
		sessions:             sessions,
		verifyCode:           vc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.SelfKeys)
	read.Get("/dashboard", h.GetDashboard)

	// Account settings — only require authenticated session, no specific permission.
	session := httpmiddleware.SessionRoutes(r, h.Protected)
	session.Get("/profile", h.GetProfile)
	session.Post("/change-password", h.ChangePassword)
	session.Post("/change-phone", h.ChangePhone)
	session.Post("/change-email", h.ChangeEmail)
	session.Post("/revoke-sessions", h.RevokeSessions)
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	view, err := h.memberAnalytics.GetDashboard(r.Context(), sessionCtx.Member.ID)
	httputil.WriteJSON(w, http.StatusOK, view, err)
}

// --- Profile ---

type profileCompany struct {
	CompanyID   uuid.UUID `json:"companyId"`
	CompanyName string    `json:"companyName"`
	Role        string    `json:"role"`
	Current     bool      `json:"current"`
}

type profileResponse struct {
	Phone       string           `json:"phone"`
	Email       string           `json:"email"`
	Name        string           `json:"name"`
	HasPassword bool             `json:"hasPassword"`
	Companies   []profileCompany `json:"companies"`
}

func maskPhone(phone string) string {
	// Expects formats like +8613812341234 or 13812341234
	if len(phone) >= 7 {
		return phone[:3] + "****" + phone[len(phone)-4:]
	}
	return "****"
}

func maskEmail(email string) string {
	for i, ch := range email {
		if ch == '@' {
			if i <= 2 {
				return email[:1] + "**" + email[i:]
			}
			return email[:2] + "**" + email[i:]
		}
	}
	return "****"
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	sessionCtx, _ := httpmiddleware.SessionFromContext(r.Context())

	ctx := r.Context()
	user, err := h.users.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	companies, err := h.users.ListMemberCompanies(ctx, claims.UserID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	resp := profileResponse{
		Phone:       maskPhone(user.Phone),
		Email:       maskEmail(user.Email),
		Name:        sessionCtx.Member.Name,
		HasPassword: user.PasswordHash != "",
		Companies:   make([]profileCompany, 0, len(companies)),
	}
	for _, c := range companies {
		resp.Companies = append(resp.Companies, profileCompany{
			CompanyID:   c.CompanyID,
			CompanyName: c.CompanyName,
			Role:        c.Role,
			Current:     c.CompanyID == claims.CompanyID,
		})
	}

	httputil.WriteOK(w, resp)
}

// --- Change Password ---

type changePasswordBody struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var body changePasswordBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if len(body.NewPassword) < 8 {
		httputil.WriteStatus(w, http.StatusBadRequest, "password too short (min 8)")
		return
	}

	ctx := r.Context()
	user, err := h.users.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// If user already has a password, verify old password.
	if user.PasswordHash != "" {
		if body.OldPassword == "" {
			httputil.WriteStatus(w, http.StatusBadRequest, "old password required")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.OldPassword)); err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, map[string]string{"code": "wrong_password", "message": "旧密码错误"}, nil)
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if err := h.users.UpdatePassword(ctx, claims.UserID, string(hash)); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Change Phone ---

type changePhoneBody struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

func (h *Handler) ChangePhone(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var body changePhoneBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Phone == "" || body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "phone and code required")
		return
	}

	address := verifycode.FormatPhone(body.Phone)
	ctx := r.Context()

	// Verify code.
	result := h.verifyCode.Verify(ctx, domainnotification.ChannelSMS, address, body.Code)
	if !result.OK {
		httputil.WriteStatus(w, http.StatusBadRequest, result.Reason)
		return
	}

	// Check uniqueness.
	existing, err := h.users.GetByPhone(ctx, address)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if existing != nil && existing.ID != claims.UserID {
		httputil.WriteStatus(w, http.StatusConflict, "该手机号已被其他账户绑定")
		return
	}

	if err := h.users.UpdatePhone(ctx, claims.UserID, address); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Change Email ---

type changeEmailBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *Handler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var body changeEmailBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Email == "" || body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "email and code required")
		return
	}

	ctx := r.Context()

	// Verify code.
	result := h.verifyCode.Verify(ctx, domainnotification.ChannelEmail, body.Email, body.Code)
	if !result.OK {
		httputil.WriteStatus(w, http.StatusBadRequest, result.Reason)
		return
	}

	// Check uniqueness.
	existing, err := h.users.GetByEmail(ctx, body.Email)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if existing != nil && existing.ID != claims.UserID {
		httputil.WriteStatus(w, http.StatusConflict, "该邮箱已被其他账户绑定")
		return
	}

	if err := h.users.UpdateEmail(ctx, claims.UserID, body.Email); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Revoke Sessions ---

func (h *Handler) RevokeSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	if err := h.sessions.RevokeAllByUserExcept(r.Context(), claims.UserID, claims.Sid); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
