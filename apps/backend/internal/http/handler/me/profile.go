package me

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

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
	Avatar      string           `json:"avatar"`
	HasPassword bool             `json:"hasPassword"`
	Companies   []profileCompany `json:"companies"`
}

func maskPhone(phone string) string {
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
		Name:        user.Name,
		Avatar:      sessionCtx.Member.Avatar,
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

const maxAvatarBytes = 50 * 1024 // 50KB decoded limit

var dicebearStyles = map[string]bool{
	"adventurer": true, "notionists": true, "bottts": true,
	"shapes": true, "lorelei": true, "fun-emoji": true,
}

type updateProfileRequest struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var body updateProfileRequest
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.Name == nil && body.Avatar == nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "nothing to update")
		return
	}

	ctx := r.Context()

	if body.Name != nil {
		name := strings.TrimSpace(*body.Name)
		if err := h.users.UpdateName(ctx, claims.UserID, name); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
	}

	if body.Avatar != nil {
		avatar := *body.Avatar
		if err := ValidateAvatar(avatar); err != nil {
			httputil.WriteError(w, err)
			return
		}
		if err := h.org.UpdateMemberAvatar(ctx, sessionCtx.CompanyID, sessionCtx.Member.ID, avatar); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func ValidateAvatar(avatar string) error {
	if avatar == "" {
		return nil // clear avatar
	}
	if strings.HasPrefix(avatar, "dicebear:") {
		parts := strings.SplitN(avatar, ":", 3)
		if len(parts) != 3 {
			return domain.BadRequest("invalid dicebear format")
		}
		if !dicebearStyles[parts[1]] {
			return domain.BadRequest("unsupported dicebear style")
		}
		if len(parts[2]) > 64 {
			return domain.BadRequest("dicebear seed too long")
		}
		return nil
	}
	if strings.HasPrefix(avatar, "data:image/") {
		// Validate base64 data URI size.
		idx := strings.Index(avatar, ",")
		if idx < 0 {
			return domain.BadRequest("invalid data URI")
		}
		encoded := avatar[idx+1:]
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return domain.BadRequest("invalid base64")
		}
		if len(decoded) > maxAvatarBytes {
			return domain.BadRequest("avatar too large (max 50KB)")
		}
		return nil
	}
	return domain.BadRequest("unsupported avatar format")
}
