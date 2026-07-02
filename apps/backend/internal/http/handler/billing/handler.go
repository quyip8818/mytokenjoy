package billing

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.SessionHandlerBase
	billingSvc domainbilling.Service
}

func NewHandler(cfg config.Config, billingSvc domainbilling.Service, sessionSvc session.Service) *Handler {
	return &Handler{
		SessionHandlerBase: shared.NewSessionHandlerBase(cfg, sessionSvc),
		billingSvc:         billingSvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Cfg, h.SessionSvc, permission.BillingRead)
	read.Get("/billing/wallet", h.GetWallet)
	write := httpmiddleware.WriteRoutes(r, h.Cfg, h.SessionSvc, permission.BillingRecharge)
	write.Post("/billing/recharge", h.CreateRecharge)
	write.Post("/billing/recharge/{id}/confirm", h.ConfirmRecharge)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	view, err := h.billingSvc.GetWallet(r.Context())
	httputil.WriteJSON(w, http.StatusOK, view, err)
}

type rechargeBody struct {
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotencyKey"`
}

func (h *Handler) CreateRecharge(w http.ResponseWriter, r *http.Request) {
	var body rechargeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	memberID := ""
	if sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context()); ok {
		memberID = sessionCtx.Member.ID
	}
	order, err := h.billingSvc.CreateSelfRecharge(r.Context(), body.Amount, body.IdempotencyKey, memberID)
	httputil.WriteJSON(w, http.StatusAccepted, order, err)
}

func (h *Handler) ConfirmRecharge(w http.ResponseWriter, r *http.Request) {
	err := h.billingSvc.ConfirmPayment(r.Context(), chi.URLParam(r, "id"))
	httputil.WriteVoid(w, err)
}
