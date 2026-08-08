package billing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/grants"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/tenant"
)

type Handler struct {
	shared.ProtectedHandlerBase
	billingSvc    domainbilling.Service
	modelDiscount store.ModelDiscountRepository
}

func NewHandler(p httpdeps.Protected, billingSvc domainbilling.Service, modelDiscount store.ModelDiscountRepository) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		billingSvc:           billingSvc,
		modelDiscount:        modelDiscount,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, grants.BillingRead)
	read.Get("/billing/wallet", h.GetWallet)
	read.Get("/billing/recharge-records", h.ListRechargeRecords)
	read.Get("/billing/discounts", h.GetDiscounts)
	read.Get("/billing/lots", h.ListLots)
	write := httpmiddleware.ReadRoutes(r, h.Protected, grants.BillingManage)
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
	if isTestingCompany(r) {
		httputil.WriteError(w, domain.ForbiddenCode("TRIAL_NO_TOPUP", "试用环境不支持充值，升级后可使用"))
		return
	}
	var body rechargeBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if isTestingCompany(r) {
		httputil.WriteError(w, domain.ForbiddenCode("TRIAL_NO_TOPUP", "试用环境不支持充值，升级后可使用"))
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

// isTestingCompany checks if the current request belongs to a testing (trial/demo/testing) tenant.
func isTestingCompany(r *http.Request) bool {
	info, ok := tenant.From(r.Context())
	return ok && company.IsTestingAccount(info.Type)
}

type discountDTO struct {
	ModelType string  `json:"modelType"`
	Discount  float64 `json:"discount"`
	Note      string  `json:"note,omitempty"`
}

func (h *Handler) GetDiscounts(w http.ResponseWriter, r *http.Request) {
	info, ok := tenant.From(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing tenant")
		return
	}
	rows, err := h.modelDiscount.CurrentDiscounts(r.Context(), info.CompanyID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	out := make([]discountDTO, len(rows))
	for i, row := range rows {
		out[i] = discountDTO{ModelType: row.ModelType, Discount: row.Discount, Note: row.Note}
	}
	httputil.WriteJSON(w, http.StatusOK, out, nil)
}

// Mount registers the billing handler on the given router.
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Protected(), d.BillingSvc, d.ModelDiscount())
	h.RegisterRoutes(r)

	// Redemption code endpoint — SaaS only (Local has no redemption_codes table).
	if d.Config.SupportSaas {
		write := httpmiddleware.ReadRoutes(r, d.Protected(), grants.BillingManage)
		write.Post("/billing/redeem", h.Redeem)
	}
}

func (h *Handler) ListLots(w http.ResponseWriter, r *http.Request) {
	lots, err := h.billingSvc.ListLots(r.Context())
	httputil.WriteJSON(w, http.StatusOK, lots, err)
}

type redeemBody struct {
	Code string `json:"code"`
}

func (h *Handler) Redeem(w http.ResponseWriter, r *http.Request) {
	if isTestingCompany(r) {
		httputil.WriteError(w, domain.ForbiddenCode("TRIAL_NO_TOPUP", "试用环境不支持兑换，升级后可使用"))
		return
	}
	var body redeemBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "code required")
		return
	}
	memberID := uuid.Nil
	if sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context()); ok {
		memberID = sessionCtx.Member.ID
	}
	result, err := h.billingSvc.RedeemCode(r.Context(), body.Code, memberID)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}
