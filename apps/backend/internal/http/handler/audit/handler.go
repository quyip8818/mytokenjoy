package audit

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type Handler struct {
	shared.SessionHandlerBase
	service        domainaudit.Service
	callLogQuerier domainusage.CallLogQuerier
}

func NewHandler(cfg config.Config, service domainaudit.Service, callLogQuerier domainusage.CallLogQuerier, sessionSvc session.Service) *Handler {
	return &Handler{
		SessionHandlerBase: shared.NewSessionHandlerBase(cfg, sessionSvc),
		service:            service,
		callLogQuerier:     callLogQuerier,
	}
}

func (h *Handler) SettingsGet(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	httputil.WriteJSON(w, http.StatusOK, settings, err)
}

func (h *Handler) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AuditSettings
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	settings, err := h.service.UpdateSettings(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, settings, err)
}

func (h *Handler) OperationsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := types.AuditOperationsQueryParams{
		Page:       common.ParseIntParam(query.Get("page"), 1),
		PageSize:   common.ParseIntParam(query.Get("pageSize"), 20),
		Action:     query.Get("action"),
		OperatorID: query.Get("operatorId"),
		Keyword:    query.Get("keyword"),
		From:       query.Get("from"),
		To:         query.Get("to"),
	}
	result, err := h.service.ListOperations(r.Context(), params)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) CallsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	params := types.AuditCallsQueryParams{
		Page:     common.ParseIntParam(query.Get("page"), 1),
		PageSize: common.ParseIntParam(query.Get("pageSize"), 20),
		Model:    query.Get("model"),
		Status:   query.Get("status"),
		CallerID: query.Get("callerId"),
		Keyword:  query.Get("keyword"),
		From:     query.Get("from"),
		To:       query.Get("to"),
	}
	result, err := h.callLogQuerier.ListCalls(r.Context(), params)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.Cfg, r, h.SessionSvc, permission.AuditRead)
	read.Get("/settings", h.SettingsGet)
	read.Get("/operations", h.OperationsList)
	read.Get("/calls", h.CallsList)

	httpmiddleware.WriteRoutes(r, h.Cfg, h.SessionSvc).
		With(httpmiddleware.RequireAnyPermission(permission.AuditRead)).
		Put("/settings", h.SettingsUpdate)
}
