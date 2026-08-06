package platform

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/httpx"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
)

type Handler struct {
	p         httpdeps.Platform
	protected httpdeps.Protected
}

func NewHandler(p httpdeps.Platform, protected httpdeps.Protected) *Handler {
	return &Handler{p: p, protected: protected}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public read-only — Catalog sync (no auth required)
	r.Get("/sync/versions", h.CatalogVersions)
	r.Get("/sync/catalog/models", h.CatalogModels)
	r.Get("/sync/catalog/currencies", h.CatalogCurrencies)

	// Per-company sync token protected — contains company-isolated data
	r.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireSyncToken(h.p.Companies))
		r.Get("/sync/catalog/pricing", h.CatalogPricing)
		r.Get("/sync/catalog/wallet_lots", h.CatalogWalletLots)
	})

	// Public — Local registration (guarded by X-Registration-Secret header)
	r.Post("/register-local", h.RegisterLocal)

	// Platform admin auth
	r.Post("/auth/login", h.Login)
	r.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireSession(h.protected))
		r.Use(httpmiddleware.RequirePlatformAdmin(h.p.Cfg.TokenJoyCompanyID, h.p.Cfg.SupportSaas))
		r.Get("/companies", h.ListCompanies)
		r.Get("/companies/overview", h.CompaniesOverview)
		r.Post("/companies", h.CreateCompany)
		r.Patch("/companies/{id}", h.UpdateCompany)
		r.Post("/companies/{id}/recharge", h.RechargeCompany)
		r.Post("/companies/{id}/gift", h.GiftCompany)
		r.Post("/companies/{id}/adjust", h.AdjustCompany)
		r.Get("/channels", h.ListChannels)
		r.Post("/channels", h.CreateChannel)
		// Model management
		r.Get("/models", h.ListModels)
		r.Post("/models", h.CreateModel)
		r.Post("/models/sync", h.SyncModelsFromNewAPI)
		r.Put("/models/{id}", h.UpdateModel)
		r.Put("/models/{id}/pricing", h.SetModelPricing)
		r.Post("/catalog/publish", h.PublishCatalog)
		// Pricing management
		r.Get("/pricing", h.ListGlobalPricing)
		r.Put("/pricing", h.SetGlobalPricing)
		r.Get("/companies/{id}/discounts", h.ListCompanyDiscounts)
		r.Put("/companies/{id}/discounts", h.SetCompanyDiscount)
		// Currency management
		r.Get("/currencies", h.ListCurrencies)
		r.Post("/currencies", h.CreateCurrency)
		r.Put("/currencies/{code}", h.UpdateCurrency)
		r.Patch("/currencies/{code}/status", h.ToggleCurrencyStatus)
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	operatorID := operatorIDFromSession(r)
	err = h.p.BillingSvc.PlatformGift(r.Context(), id, body.Amount, operatorID)
	httputil.WriteVoid(w, err)
}

type adjustBody struct {
	Amount     float64 `json:"amount"`
	PaidAmount float64 `json:"paidAmount"`
}

func (h *Handler) AdjustCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body adjustBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	operatorID := operatorIDFromSession(r)
	err = h.p.BillingSvc.PlatformAdjust(r.Context(), id, body.Amount, body.PaidAmount, operatorID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	keys, err := h.p.KeysSvc.ListProviderKeys(r.Context())
	httputil.WriteJSON(w, http.StatusOK, keys, err)
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var body types.CreateProviderKeyInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	key, err := h.p.KeysSvc.CreateProviderKeyForPlatform(r.Context(), body)
	httputil.WriteJSON(w, http.StatusCreated, key, err)
}

// Mount registers the platform handler on the given router under /platform (SaaS only).
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Platform(), d.Protected())
	r.Route("/platform", h.RegisterRoutes)
}
