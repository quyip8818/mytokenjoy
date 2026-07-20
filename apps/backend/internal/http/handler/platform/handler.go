package platform

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

type Handler struct {
	p         httpdeps.Platform
	protected httpdeps.Protected
}

func NewHandler(p httpdeps.Platform, protected httpdeps.Protected) *Handler {
	return &Handler{p: p, protected: protected}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireSession(h.protected))
		r.Use(httpmiddleware.RequirePlatformAdmin(h.p.Cfg.TokenJoyCompanyID))
		r.Get("/companies", h.ListCompanies)
		r.Post("/companies", h.CreateCompany)
		r.Patch("/companies/{id}", h.UpdateCompany)
		r.Post("/companies/{id}/recharge", h.RechargeCompany)
		r.Post("/companies/{id}/gift", h.GiftCompany)
		r.Post("/companies/{id}/adjust", h.AdjustCompany)
		r.Get("/channels", h.ListChannels)
		r.Post("/channels", h.CreateChannel)
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	member, err := h.p.Credentials.AuthenticateMember(r.Context(), h.p.Cfg.TokenJoyCompanyID, body.Email, body.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, err)
		return
	}
	if _, err := httpx.IssueTokenPair(r.Context(), w, r, httpx.TokenPairParams{
		Secret:        h.p.SessionToken.Secret(),
		SessionTTLSec: h.p.Cfg.SessionTTLSec,
		RefreshTTLSec: h.p.Cfg.RefreshTokenTTLSec,
		SecureCookie:  h.p.SecureCookie,
		SessionStore:  h.p.Sessions,
	}, member.CompanyID, member.ID, member.UserID); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"memberId": member.ID.String()}, nil)
}

type createCompanyBody struct {
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
	Email string `json:"email"` // invite email for deferred join
}

func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var body createCompanyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.Email == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "email required")
		return
	}
	companyType := body.Type
	if companyType == "" {
		companyType = "standard"
	}
	result, err := h.p.CompanySvc.CreateCompany(r.Context(), domaincompany.CreateCompanyRequest{
		Name:        body.Name,
		Type:        companyType,
		InviteEmail: body.Email,
	})
	httputil.WriteJSON(w, http.StatusCreated, result, err)
}

func (h *Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.p.CompanySvc.ListCompanies(r.Context())
	httputil.WriteJSON(w, http.StatusOK, companies, err)
}

type updateCompanyBody struct {
	Status *string `json:"status"`
}

func (h *Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body updateCompanyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.Status != nil {
		err = h.p.CompanySvc.UpdateCompany(r.Context(), id, domaincompany.UpdateCompanyPatch{
			Status: body.Status,
		})
	}
	httputil.WriteVoid(w, err)
}

func operatorIDFromSession(r *http.Request) uuid.UUID {
	session, ok := httpx.SessionFromContext(r.Context())
	if !ok {
		return uuid.Nil
	}
	return session.Member.ID
}

type rechargeBody struct {
	Amount float64 `json:"amount"`
}

func (h *Handler) RechargeCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body rechargeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	operatorID := operatorIDFromSession(r)
	err = h.p.BillingSvc.PlatformRecharge(r.Context(), id, body.Amount, operatorID)
	httputil.WriteVoid(w, err)
}

type giftBody struct {
	Amount float64 `json:"amount"`
}

func (h *Handler) GiftCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body giftBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	operatorID := operatorIDFromSession(r)
	err = h.p.BillingSvc.PlatformGift(r.Context(), id, body.Amount, operatorID)
	httputil.WriteVoid(w, err)
}

type adjustBody struct {
	Amount        float64 `json:"amount"`
	AmountDisplay float64 `json:"amountDisplay"`
}

func (h *Handler) AdjustCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body adjustBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	operatorID := operatorIDFromSession(r)
	err = h.p.BillingSvc.PlatformAdjust(r.Context(), id, body.Amount, body.AmountDisplay, operatorID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	keys, err := h.p.KeysSvc.ListProviderKeys(r.Context())
	httputil.WriteJSON(w, http.StatusOK, keys, err)
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var body types.CreateProviderKeyInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	key, err := h.p.KeysSvc.CreateProviderKeyForPlatform(r.Context(), body)
	httputil.WriteJSON(w, http.StatusCreated, key, err)
}
