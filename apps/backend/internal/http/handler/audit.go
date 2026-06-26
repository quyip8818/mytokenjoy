package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/response"
	pkg "github.com/tokenjoy/backend/internal/pkg"
)

type AuditHandler struct {
	service domainaudit.Service
}

func NewAuditHandler(service domainaudit.Service) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) SettingsGet(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetSettings())
}

func (h *AuditHandler) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AuditSettings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	response.JSON(w, http.StatusOK, h.service.UpdateSettings(body))
}

func (h *AuditHandler) OperationsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := types.AuditOperationsQueryParams{
		Page:       pkg.ParseIntParam(query.Get("page"), 1),
		PageSize:   pkg.ParseIntParam(query.Get("pageSize"), 20),
		Action:     query.Get("action"),
		OperatorID: query.Get("operatorId"),
		Keyword:    query.Get("keyword"),
		From:       query.Get("from"),
		To:         query.Get("to"),
	}
	response.JSON(w, http.StatusOK, h.service.ListOperations(params))
}

func (h *AuditHandler) CallsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := types.AuditCallsQueryParams{
		Page:     pkg.ParseIntParam(query.Get("page"), 1),
		PageSize: pkg.ParseIntParam(query.Get("pageSize"), 20),
		Model:    query.Get("model"),
		Status:   query.Get("status"),
		CallerID: query.Get("callerId"),
		Keyword:  query.Get("keyword"),
		From:     query.Get("from"),
		To:       query.Get("to"),
	}
	response.JSON(w, http.StatusOK, h.service.ListCalls(params))
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	r.Get("/settings", h.SettingsGet)
	r.Put("/settings", h.SettingsUpdate)
	r.Get("/operations", h.OperationsList)
	r.Get("/calls", h.CallsList)
}
