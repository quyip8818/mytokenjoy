package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ModelDiscountRow represents a single discount entry in the append-only timeline.
type ModelDiscountRow struct {
	ID            uuid.UUID
	CompanyID     uuid.UUID
	ModelType     string  // exact model type or "*" for wildcard
	Discount      float64 // multiplier: 0.8 = 20% off, 1.2 = 20% markup
	EffectiveFrom time.Time
	Note          string
	CreatedAt     time.Time
}

// ModelDiscountRepository manages the model_discount append-only table.
type ModelDiscountRepository interface {
	// CurrentDiscounts returns all effective discounts for a company
	// (one row per model_type, the latest effective_from <= now).
	CurrentDiscounts(ctx context.Context, companyID uuid.UUID) ([]ModelDiscountRow, error)

	// Insert appends a new discount row.
	Insert(ctx context.Context, row ModelDiscountRow) error
}
