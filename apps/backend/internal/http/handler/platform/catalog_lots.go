package platform

import (
	"net/http"
	"strconv"

	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
)

// catalogLotDTO is a lot entry in the sync response.
type catalogLotDTO struct {
	ID              string  `json:"id"`
	OrderID         string  `json:"orderId"`
	LotKind         string  `json:"lotKind"`
	BillingCurrency string  `json:"billingCurrency"`
	QuotaPerUnit    int64   `json:"quotaPerUnit"`
	QuotaGranted    int64   `json:"quotaGranted"`
	QuotaRemaining  int64   `json:"quotaRemaining"`
	PaidAmount      float64 `json:"paidAmount"`
	Status          string  `json:"status"`
	CreatedAt       int64   `json:"createdAt"` // unix seconds
}

// catalogOrderDTO is a recharge order entry in the sync response.
type catalogOrderDTO struct {
	ID             string  `json:"id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	QuotaPerUnit   int64   `json:"quotaPerUnit"`
	QuotaGranted   int64   `json:"quotaGranted"`
	Source         string  `json:"source"`
	LotKind        string  `json:"lotKind"`
	Status         string  `json:"status"`
	DisplayOrderID string  `json:"displayOrderId"`
	PaymentMethod  string  `json:"paymentMethod"`
	CreatedAt      int64   `json:"createdAt"` // unix seconds
}

// CatalogWalletLots returns active lots + wallet balance for the authenticated sync company.
// GET /api/platform/sync/catalog/wallet_lots (requires sync token)
func (h *Handler) CatalogWalletLots(w http.ResponseWriter, r *http.Request) {
	companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	ctx := r.Context()

	co, err := h.p.Companies.GetByID(ctx, companyID)
	if err != nil || co == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, "company not found")
		return
	}

	// List all active lots (FIFO order).
	lots, err := h.p.Billing.ListActiveLotsFIFO(ctx, companyID, co.FIFOHeadLotID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Fetch orders for these lots.
	orders, err := h.p.Billing.ListRechargeOrders(ctx, companyID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	orderMap := make(map[string]catalogOrderDTO, len(orders))
	for _, o := range orders {
		orderMap[o.ID.String()] = catalogOrderDTO{
			ID:             o.ID.String(),
			Amount:         o.Amount,
			Currency:       o.Currency,
			QuotaPerUnit:   o.QuotaPerUnit,
			QuotaGranted:   o.QuotaGranted,
			Source:         o.Source,
			LotKind:        o.LotKind,
			Status:         o.Status,
			DisplayOrderID: o.DisplayOrderID,
			PaymentMethod:  o.PaymentMethod,
			CreatedAt:      o.CreatedAt.Unix(),
		}
	}

	data := make([]catalogLotDTO, 0, len(lots))
	for _, lot := range lots {
		data = append(data, catalogLotDTO{
			ID:              lot.ID.String(),
			OrderID:         lot.RechargeOrderID.String(),
			LotKind:         lot.LotKind,
			BillingCurrency: lot.BillingCurrency,
			QuotaPerUnit:    lot.QuotaPerUnit,
			QuotaGranted:    lot.QuotaGranted,
			QuotaRemaining:  lot.QuotaRemaining,
			PaidAmount:      lot.PaidAmount,
			Status:          lot.Status,
			CreatedAt:       lot.CreatedAt.Unix(),
		})
	}

	// Only include orders that have active lots (avoid sending historical exhausted orders).
	activeOrders := make([]catalogOrderDTO, 0, len(lots))
	seen := make(map[string]bool)
	for _, lot := range lots {
		oid := lot.RechargeOrderID.String()
		if seen[oid] {
			continue
		}
		seen[oid] = true
		if o, ok := orderMap[oid]; ok {
			activeOrders = append(activeOrders, o)
		}
	}

	vStr, _ := h.p.SystemSettings.Get(ctx, catalogWalletLotsVersionKey)
	v, _ := strconv.Atoi(vStr)

	response.JSON(w, http.StatusOK, map[string]any{
		"version":           v,
		"data":              data,
		"orders":            activeOrders,
		"walletRemainQuota": co.WalletRemainQuota,
	})
}
