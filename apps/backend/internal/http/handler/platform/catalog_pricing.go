package platform

import (
	"context"
	"net/http"
	"strconv"

	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
)

// catalogPricingDTO is a model price entry in the sync response.
type catalogPricingDTO struct {
	ModelType   string  `json:"modelType"`
	InputPrice  float64 `json:"inputPrice"`
	OutputPrice float64 `json:"outputPrice"`
}

// CatalogPricing returns global pricing for the authenticated sync company.
func (h *Handler) CatalogPricing(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	ctx := r.Context()

	models, err := h.p.PricingSvc.ListGlobalPricing(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	out := make([]catalogPricingDTO, 0, len(models))
	for _, m := range models {
		out = append(out, catalogPricingDTO{
			ModelType:   m.Type,
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
		})
	}

	response.JSON(w, http.StatusOK, map[string]any{"version": h.pricingVersion(ctx), "data": out})
}

func (h *Handler) pricingVersion(ctx context.Context) int {
	vStr, _ := h.p.SystemSettings.Get(ctx, catalogPricingVersionKey)
	v, _ := strconv.Atoi(vStr)
	return v
}
