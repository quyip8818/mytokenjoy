package me

import (
	"net/http"

	"github.com/google/uuid"
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
