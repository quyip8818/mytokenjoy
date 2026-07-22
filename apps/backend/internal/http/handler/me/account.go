package me

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"golang.org/x/crypto/bcrypt"
)

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

	result := h.verifyCode.Verify(ctx, domainnotification.ChannelSMS, address, body.Code)
	if !result.OK {
		httputil.WriteStatus(w, http.StatusBadRequest, result.Reason)
		return
	}

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

	result := h.verifyCode.Verify(ctx, domainnotification.ChannelEmail, body.Email, body.Code)
	if !result.OK {
		httputil.WriteStatus(w, http.StatusBadRequest, result.Reason)
		return
	}

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
