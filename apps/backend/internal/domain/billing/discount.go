package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

// BatchSetDiscounts atomically replaces discount entries for a company.
// Entries with Discount == 1.0 are treated as deletion.
func (s *service) BatchSetDiscounts(ctx context.Context, companyID uuid.UUID, entries []DiscountSetEntry) error {
	err := s.store.WithTx(ctx, func(tx store.Store) error {
		repo := tx.ModelDiscount()
		for _, item := range entries {
			if item.ModelType == "" || item.Discount <= 0 {
				continue
			}
			if item.Discount == 1.0 {
				if err := repo.DeleteByCompanyAndModel(ctx, companyID, item.ModelType); err != nil {
					return err
				}
			} else {
				row := store.ModelDiscountRow{
					CompanyID: companyID,
					ModelType: item.ModelType,
					Discount:  item.Discount,
					Note:      item.Note,
				}
				if err := repo.Insert(ctx, row); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, _ = s.store.SyncVersions().Increment(ctx, companyID, "discounts")
	return nil
}
