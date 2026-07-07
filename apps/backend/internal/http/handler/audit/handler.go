package audit

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service   domainaudit.Service
	readModel domainusage.ReadModel
}

func NewHandler(p httpdeps.Protected, service domainaudit.Service, readModel domainusage.ReadModel) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
		readModel:            readModel,
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
	result, err := h.readModel.ListCalls(r.Context(), params)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.AuditRead)
	read.Get("/settings", h.SettingsGet)
	read.Get("/operations", h.OperationsList)
	read.Get("/calls", h.CallsList)

	httpmiddleware.ReadRoutes(r, h.Protected).
		With(httpmiddleware.RequireAnyPermission(permission.OrgStructure)).
		Put("/settings", h.SettingsUpdate)
}
