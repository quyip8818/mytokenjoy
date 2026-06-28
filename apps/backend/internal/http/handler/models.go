package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/permission"
)

type ModelsHandler struct {
	service    domainmodels.Service
	sessionSvc session.Service
}

func NewModelsHandler(service domainmodels.Service, sessionSvc session.Service) *ModelsHandler {
	return &ModelsHandler{service: service, sessionSvc: sessionSvc}
}

func (h *ModelsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/routing", h.RoutingList)
	r.Get("/routing/resolve", h.RoutingResolve)

	sessionWrite := r.With(httpmiddleware.RequireSession(h.sessionSvc))

	manageWrite := sessionWrite.With(httpmiddleware.RequireAnyPermission(permission.ModelManage))
	manageWrite.Post("/", h.Create)
	manageWrite.Put("/{id}/toggle", h.Toggle)

	whitelistWrite := sessionWrite.With(httpmiddleware.RequireAnyPermission(permission.ModelWhitelist))
	whitelistWrite.Put("/routing/{id}", h.RoutingUpdate)
}

func (h *ModelsHandler) List(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListModels())
}

func (h *ModelsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body types.CreateModelInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	model, err := h.service.CreateModel(r.Context(), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, model)
}

func (h *ModelsHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	var body types.ToggleModelInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.service.ToggleModel(r.Context(), chi.URLParam(r, "id"), body.Enabled); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *ModelsHandler) RoutingList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListRoutingRules())
}

func (h *ModelsHandler) RoutingResolve(w http.ResponseWriter, r *http.Request) {
	deptID := r.URL.Query().Get("deptId")
	response.JSON(w, http.StatusOK, h.service.ResolveRouting(deptID))
}

func (h *ModelsHandler) RoutingUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.UpdateRoutingRuleInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	rule, err := h.service.UpdateRoutingRule(r.Context(), chi.URLParam(r, "id"), body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, rule)
}
