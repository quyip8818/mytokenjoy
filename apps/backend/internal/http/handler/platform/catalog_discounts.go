package platform

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
)

// CatalogDiscounts returns per-company discount coefficients for the authenticated sync company.
// GET /api/platform/sync/catalog/discounts (requires sync token)
func (h *Handler) CatalogDiscounts(w http.ResponseWriter, r *http.Request) {
	companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	rows, err := h.p.ModelDiscount.CurrentDiscounts(r.Context(), companyID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	data := make([]discountDTO, len(rows))
	for i, row := range rows {
		data[i] = discountDTO{ModelType: row.ModelType, Discount: row.Discount}
	}

	v, _ := h.p.SyncVersions.Get(r.Context(), companyID, "discounts")
	response.JSON(w, http.StatusOK, map[string]any{"version": v, "data": data})
}
