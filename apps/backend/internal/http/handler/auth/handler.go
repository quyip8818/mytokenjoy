package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

type Handler struct {
	pub        httpdeps.Public
	companySvc domaincompany.Service
}

func NewHandler(pub httpdeps.Public, companySvc domaincompany.Service) *Handler {
	return &Handler{
		pub:        pub,
		companySvc: companySvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/accept-invite", h.AcceptInvite)
}

type loginBody struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CompanyID uuid.UUID `json:"companyId"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var companyID uuid.UUID
	if h.pub.Cfg.SupportSaas {
		if body.CompanyID == uuid.Nil {
			httputil.WriteJSON(w, http.StatusBadRequest, nil, domain.BadRequest("company id required"))
			return
		}
		companyCtx, err := h.companySvc.ResolveCompanyContext(r.Context(), body.CompanyID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
			return
		}
		companyID = companyCtx.CompanyID
	} else {
		companyCtx, ok := ctxcompany.From(r.Context())
		if !ok {
			httputil.WriteStatus(w, http.StatusBadRequest, "Company not found")
			return
		}
		companyID = companyCtx.CompanyID
	}
	member, err := h.pub.Credentials.AuthenticateMember(r.Context(), companyID, body.Email, body.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, err)
		return
	}
	token, err := h.pub.SessionToken.IssueWithUser(member.CompanyID, member.ID, member.UserID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httpx.SetSessionCookie(w, token, h.pub.SecureCookie)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"memberId": member.ID.String()}, nil)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.ClearSessionCookie(w)
	httputil.WriteStatus(w, http.StatusNoContent, "")
}

type acceptInviteBody struct {
	InviteCode string `json:"inviteCode"`
	Name       string `json:"name"`
	Password   string `json:"password"`
}

func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var body acceptInviteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	member, err := h.companySvc.AcceptInvite(r.Context(), domaincompany.AcceptInviteRequest{
		InviteCode: body.InviteCode, Name: body.Name, Password: body.Password,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}
	token, err := h.pub.SessionToken.IssueWithUser(member.CompanyID, member.ID, member.UserID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httpx.SetSessionCookie(w, token, h.pub.SecureCookie)
	httputil.WriteJSON(w, http.StatusOK, member, nil)
}
