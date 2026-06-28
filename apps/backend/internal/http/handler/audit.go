package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
	pkg "github.com/tokenjoy/backend/internal/pkg"
)

type AuditHandler struct {
	cfg        config.Config
	service    domainaudit.Service
	sessionSvc session.Service
}

func NewAuditHandler(cfg config.Config, service domainaudit.Service, sessionSvc session.Service) *AuditHandler {
	return &AuditHandler{cfg: cfg, service: service, sessionSvc: sessionSvc}
}

func (h *AuditHandler) SettingsGet(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetSettings())
}

func (h *AuditHandler) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.AuditSettings
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, h.service.UpdateSettings(body))
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
	httputil.WriteOK(w, h.service.ListOperations(params))
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
	httputil.WriteOK(w, h.service.ListCalls(params))
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.cfg, r, h.sessionSvc, permission.AuditRead)
	read.Get("/settings", h.SettingsGet)
	read.Get("/operations", h.OperationsList)
	read.Get("/calls", h.CallsList)

	httpmiddleware.WriteRoutes(r, h.cfg, h.sessionSvc).
		With(httpmiddleware.RequireAnyPermission(permission.AuditRead)).
		Put("/settings", h.SettingsUpdate)
}
