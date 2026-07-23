package billing

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
)

type Handler struct {
	shared.ProtectedHandlerBase
	billingSvc domainbilling.Service
}

func NewHandler(p httpdeps.Protected, billingSvc domainbilling.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		billingSvc:           billingSvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.BillingRead)
	read.Get("/billing/wallet", h.GetWallet)
	read.Get("/billing/recharge-records", h.ListRechargeRecords)
	write := httpmiddleware.ReadRoutes(r, h.Protected, permission.BillingRecharge)
	write.Post("/billing/recharge", h.CreateRecharge)
	write.Post("/billing/recharge/{id}/confirm", h.ConfirmRecharge)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	view, err := h.billingSvc.GetWallet(r.Context())
	httputil.WriteJSON(w, http.StatusOK, view, err)
}

func (h *Handler) ListRechargeRecords(w http.ResponseWriter, r *http.Request) {
	records, err := h.billingSvc.ListRechargeRecords(r.Context())
	httputil.WriteJSON(w, http.StatusOK, records, err)
}

type rechargeBody struct {
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotencyKey"`
}

func (h *Handler) CreateRecharge(w http.ResponseWriter, r *http.Request) {
	if isTrialOrDemoCompany(r) {
		httputil.WriteError(w, domain.Forbidden("试用环境不支持充值，升级后可使用"))
		return
	}
	var body rechargeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	memberID := uuid.Nil
	if sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context()); ok {
		memberID = sessionCtx.Member.ID
	}
	order, err := h.billingSvc.CreateSelfRecharge(r.Context(), body.Amount, body.IdempotencyKey, memberID)
	httputil.WriteJSON(w, http.StatusAccepted, order, err)
}

func (h *Handler) ConfirmRecharge(w http.ResponseWriter, r *http.Request) {
	if isTrialOrDemoCompany(r) {
		httputil.WriteError(w, domain.Forbidden("试用环境不支持充值，升级后可使用"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.billingSvc.ConfirmPayment(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// isTrialOrDemoCompany checks if the current request belongs to a trial or demo tenant.
func isTrialOrDemoCompany(r *http.Request) bool {
	info, ok := ctxcompany.From(r.Context())
	return ok && (info.Type == store.CompanyTypeTrial || info.Type == store.CompanyTypeDemo)
}
