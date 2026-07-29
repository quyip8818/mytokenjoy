package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ModelPricingRow represents a single price point in the append-only timeline.
type ModelPricingRow struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	ModelType     string
	InputPrice    float64 // 元/1M tokens
	OutputPrice   float64 // 元/1M tokens
	EffectiveFrom time.Time
	Note          string
	CreatedAt     time.Time
}

// ModelPricingRepository manages the model_pricing append-only table.
type ModelPricingRepository interface {
	// CurrentPrice returns the effective price for a company+model at time `at`.
	// Returns nil if no pricing row exists.
	CurrentPrice(ctx context.Context, companyID uuid.UUID, modelType string, at time.Time) (*ModelPricingRow, error)

	// CurrentPricesBatch returns all effective prices for a company at time `at`
	// (one row per model_type, the latest effective_from <= at).
	CurrentPricesBatch(ctx context.Context, companyID uuid.UUID, at time.Time) ([]ModelPricingRow, error)

	// History returns the full price timeline for a company+model (effective_from DESC).
	History(ctx context.Context, companyID uuid.UUID, modelType string) ([]ModelPricingRow, error)

	// Insert appends a new price row. Changing price = new INSERT.
	Insert(ctx context.Context, row ModelPricingRow) error
}
