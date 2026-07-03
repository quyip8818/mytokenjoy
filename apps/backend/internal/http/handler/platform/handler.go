package platform

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

type Handler struct {
	p httpdeps.Platform
}

func NewHandler(p httpdeps.Platform) *Handler {
	return &Handler{p: p}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Group(func(r chi.Router) {
		r.Use(httpmiddleware.PlatformAuth(h.p.Cfg, h.p.PlatformSessionToken))
		r.Get("/companies", h.ListCompanies)
		r.Post("/companies", h.CreateCompany)
		r.Patch("/companies/{id}", h.UpdateCompany)
		r.Post("/companies/{id}/recharge", h.RechargeCompany)
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
	operatorID, err := h.p.Credentials.AuthenticatePlatform(r.Context(), body.Email, body.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, err)
		return
	}
	token, err := h.p.PlatformSessionToken.Issue(0, operatorID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httpx.SetPlatformSessionCookie(w, token, h.p.SecureCookie)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"operatorId": operatorID}, nil)
}

type createCompanyBody struct {
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	SuperAdminEmail string  `json:"superAdminEmail"`
	PackageID       *string `json:"packageId"`
}

func (h *Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var body createCompanyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	result, err := h.p.CompanySvc.CreateCompany(r.Context(), domaincompany.CreateCompanyRequest{
		Slug: body.Slug, Name: body.Name, SuperAdminEmail: body.SuperAdminEmail, PackageID: body.PackageID,
	})
	httputil.WriteJSON(w, http.StatusCreated, result, err)
}

func (h *Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := h.p.CompanySvc.ListCompanies(r.Context())
	httputil.WriteJSON(w, http.StatusOK, companies, err)
}

type updateCompanyBody struct {
	Status    *string `json:"status"`
	PackageID *string `json:"packageId"`
}

func (h *Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body updateCompanyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.Status != nil || body.PackageID != nil {
		err = h.p.CompanySvc.UpdateCompany(r.Context(), id, domaincompany.UpdateCompanyPatch{
			Status:    body.Status,
			PackageID: body.PackageID,
		})
	}
	httputil.WriteVoid(w, err)
}

type rechargeBody struct {
	Amount float64 `json:"amount"`
}

func (h *Handler) RechargeCompany(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	var body rechargeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	operatorID, _ := httpmiddleware.PlatformOperatorFromContext(r.Context())
	err = h.p.BillingSvc.PlatformRecharge(r.Context(), id, body.Amount, operatorID)
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
	key, err := h.p.KeysSvc.CreatePlatformProviderKey(r.Context(), body)
	httputil.WriteJSON(w, http.StatusCreated, key, err)
}
