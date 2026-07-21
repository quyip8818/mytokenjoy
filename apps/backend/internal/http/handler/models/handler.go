package models

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service domainmodels.Service
}

func NewHandler(p httpdeps.Protected, service domainmodels.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.ModelRead)
	read.Get("/", h.List)
	read.Get("/routing", h.RoutingList)
	read.Get("/routing/resolve", h.RoutingResolve)

	write := httpmiddleware.ReadRoutes(r, h.Protected)

	manageWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelManage))
	manageWrite.Post("/", h.Create)
	manageWrite.Put("/{id}", h.Update)
	manageWrite.Delete("/{id}", h.Delete)
	manageWrite.Put("/{id}/toggle", h.Toggle)

	whitelistWrite := write.With(httpmiddleware.RequireAnyPermission(permission.ModelWhitelist))
	whitelistWrite.Put("/routing/{id}", h.RoutingUpdate)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.ListModelsWithPricing(r.Context())
	httputil.WriteJSON(w, http.StatusOK, models, err)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body types.CreateModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	model, err := h.service.CreateModel(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, model, err)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body types.UpdateModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	model, err := h.service.UpdateModel(r.Context(), id, body)
	httputil.WriteJSON(w, http.StatusOK, model, err)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.service.DeleteModel(r.Context(), id)
	httputil.WriteVoid(w, err)
}

func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body types.ToggleModelInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err = h.service.ToggleModel(r.Context(), id, body.Enabled)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoutingList(w http.ResponseWriter, r *http.Request) {
	rules, err := h.service.ListRoutingRules(r.Context())
	httputil.WriteJSON(w, http.StatusOK, rules, err)
}

func (h *Handler) RoutingResolve(w http.ResponseWriter, r *http.Request) {
	deptID, _ := uuid.Parse(r.URL.Query().Get("deptId"))
	result, err := h.service.ResolveRouting(r.Context(), deptID)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) RoutingUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body types.UpdateRoutingRuleInput
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	rule, err := h.service.UpdateRoutingRule(r.Context(), id, body)
	httputil.WriteJSON(w, http.StatusOK, rule, err)
}
