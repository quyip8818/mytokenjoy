package platform

import (
	"context"
	"net/http"

	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
)

// catalogPricingDTO is a model price entry in the sync response.
type catalogPricingDTO struct {
	ModelType       string  `json:"modelType"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

// CatalogPricing returns global pricing for the authenticated sync company.
// ponytail: reads from NewAPI (SOT) in real-time, no DB cache.
func (h *Handler) CatalogPricing(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	ctx := r.Context()

	ratios, err := h.p.AdminPort.ListModelPricing(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	out := make([]catalogPricingDTO, 0, len(ratios))
	for _, r := range ratios {
		inputPrice, outputPrice, cacheInputPrice := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio, r.CacheRatio)
		out = append(out, catalogPricingDTO{
			ModelType:       r.ModelName,
			InputPrice:      inputPrice,
			OutputPrice:     outputPrice,
			CacheInputPrice: cacheInputPrice,
		})
	}

	response.JSON(w, http.StatusOK, map[string]any{"version": h.pricingVersion(ctx), "data": out})
}

func (h *Handler) pricingVersion(ctx context.Context) int {
	v, _ := h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "pricing")
	return v
}
