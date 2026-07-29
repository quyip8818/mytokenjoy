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
	IsContract  bool    `json:"isContract"` // true = company-specific override
}

// CatalogPricing returns merged global + contract pricing for the authenticated sync company.
// Contract prices override global prices for the same modelType.
func (h *Handler) CatalogPricing(w http.ResponseWriter, r *http.Request) {
	companyID, ok := httpmiddleware.SyncCompanyIDFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, "missing sync identity")
		return
	}

	ctx := r.Context()

	globalPrices, err := h.p.PricingSvc.ListGlobalPricing(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	contractPrices, err := h.p.PricingSvc.ListContractPricing(ctx, companyID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Build contract lookup: modelType → price
	contractMap := make(map[string]catalogPricingDTO, len(contractPrices))
	for _, cp := range contractPrices {
		contractMap[cp.ModelType] = catalogPricingDTO{
			ModelType:   cp.ModelType,
			InputPrice:  cp.InputPrice,
			OutputPrice: cp.OutputPrice,
			IsContract:  true,
		}
	}

	// Merge: contract overrides global for the same modelType
	merged := make([]catalogPricingDTO, 0, len(globalPrices))
	for _, gp := range globalPrices {
		if cp, ok := contractMap[gp.ModelType]; ok {
			merged = append(merged, cp)
			delete(contractMap, gp.ModelType)
		} else {
			merged = append(merged, catalogPricingDTO{
				ModelType:   gp.ModelType,
				InputPrice:  gp.InputPrice,
				OutputPrice: gp.OutputPrice,
				IsContract:  false,
			})
		}
	}
	// Append contract-only prices (no global equivalent)
	for _, cp := range contractMap {
		merged = append(merged, cp)
	}

	response.JSON(w, http.StatusOK, map[string]any{"version": h.pricingVersion(ctx), "data": merged})
}

func (h *Handler) pricingVersion(ctx context.Context) int {
	vStr, _ := h.p.SystemSettings.Get(ctx, catalogPricingVersionKey)
	v, _ := strconv.Atoi(vStr)
	return v
}
