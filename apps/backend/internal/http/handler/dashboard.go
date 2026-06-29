package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
	pkg "github.com/tokenjoy/backend/internal/pkg"
)

type DashboardHandler struct {
	cfg        config.Config
	service    domaindashboard.Service
	sessionSvc session.Service
}

func NewDashboardHandler(cfg config.Config, service domaindashboard.Service, sessionSvc session.Service) *DashboardHandler {
	return &DashboardHandler{cfg: cfg, service: service, sessionSvc: sessionSvc}
}

func (h *DashboardHandler) CostSummary(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		result, err := h.service.CostSummary(ctx, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) DepartmentCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		params := parseCostQueryParams(r)
		result, err := h.service.DepartmentCosts(ctx, query.Get("parentId"), params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) DepartmentMemberCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		result, err := h.service.DepartmentMemberCosts(ctx, chi.URLParam(r, "deptId"), params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) DailyCosts(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		result, err := h.service.DailyCosts(ctx, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) TopConsumers(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		limit := pkg.ParseIntParam(query.Get("limit"), 5)
		params := parseCostQueryParams(r)
		result, err := h.service.TopConsumers(ctx, limit, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) ModelUsage(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		result, err := h.service.ModelUsage(ctx, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) TeamUsage(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		params := parseCostQueryParams(r)
		result, err := h.service.TeamUsage(ctx, params, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) UsageSeries(w http.ResponseWriter, r *http.Request) {
	h.withScope(w, r, func(ctx context.Context, scope domainusage.SessionScope) {
		query := r.URL.Query()
		granularity := query.Get("granularity")
		startRaw := query.Get("start")
		endRaw := query.Get("end")
		if granularity == "" || startRaw == "" || endRaw == "" {
			httputil.WriteError(w, domain.BadRequest("granularity, start and end are required"))
			return
		}
		timezone := domainusage.ResolveTimezone("")
		start, end, err := domainusage.ParseSeriesTimeRange(startRaw, endRaw, granularity, timezone)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		groupBy := query.Get("groupBy")
		if groupBy == "" {
			groupBy = domainusage.GroupByNone
		}
		result, err := h.service.UsageSeries(ctx, types.UsageSeriesQuery{
			Granularity:  granularity,
			Start:        start,
			End:          end,
			GroupBy:      groupBy,
			DepartmentID: query.Get("departmentId"),
			MemberID:     query.Get("memberId"),
			Timezone:     timezone,
		}, scope)
		httputil.WriteJSON(w, http.StatusOK, result, err)
	})
}

func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.cfg, h.sessionSvc, permission.DashboardCost, permission.DashboardUsage)
	read.Get("/cost/summary", h.CostSummary)
	read.Get("/cost/departments", h.DepartmentCosts)
	read.Get("/cost/departments/{deptId}/members", h.DepartmentMemberCosts)
	read.Get("/cost/daily", h.DailyCosts)
	read.Get("/cost/top", h.TopConsumers)
	read.Get("/usage/models", h.ModelUsage)
	read.Get("/usage/teams", h.TeamUsage)
	read.Get("/usage/series", h.UsageSeries)
}

func (h *DashboardHandler) withScope(w http.ResponseWriter, r *http.Request, fn func(context.Context, domainusage.SessionScope)) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	fn(r.Context(), domainusage.SessionScope{
		MemberID:     sessionCtx.Member.ID,
		DepartmentID: sessionCtx.Member.DepartmentID,
		Permissions:  sessionCtx.Permissions,
	})
}

func parseCostQueryParams(r *http.Request) types.CostQueryParams {
	query := r.URL.Query()
	return types.CostQueryParams{
		Period:      query.Get("period"),
		StartDate:   query.Get("startDate"),
		EndDate:     query.Get("endDate"),
		Granularity: query.Get("granularity"),
	}
}
