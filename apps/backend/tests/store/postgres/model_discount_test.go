//go:build testhook

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

func TestModelDiscountInsertAndQuery(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := context.Background()

	companyID := store.CompanyID(ctx)
	if companyID == uuid.Nil {
		// Use the first seeded company
		co, err := st.Company().List(ctx)
		if err != nil || len(co) == 0 {
			t.Fatal("no companies in test db")
		}
		companyID = co[0].ID
	}

	// Insert two discounts for different models
	err := st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID,
		ModelType: "gpt-4o",
		Discount:  0.8,
		Note:      "test discount",
	})
	if err != nil {
		t.Fatalf("insert gpt-4o discount: %v", err)
	}

	err = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID,
		ModelType: "*",
		Discount:  0.9,
		Note:      "wildcard",
	})
	if err != nil {
		t.Fatalf("insert wildcard discount: %v", err)
	}

	// Query current discounts
	rows, err := st.ModelDiscount().CurrentDiscounts(ctx, companyID)
	if err != nil {
		t.Fatalf("CurrentDiscounts: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 discount rows, got %d", len(rows))
	}

	// Verify values
	found := map[string]float64{}
	for _, r := range rows {
		found[r.ModelType] = r.Discount
	}
	if found["gpt-4o"] != 0.8 {
		t.Errorf("gpt-4o discount = %f, want 0.8", found["gpt-4o"])
	}
	if found["*"] != 0.9 {
		t.Errorf("wildcard discount = %f, want 0.9", found["*"])
	}
}

func TestModelDiscountAppendOnlyLatestWins(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := context.Background()

	co, err := st.Company().List(ctx)
	if err != nil || len(co) == 0 {
		t.Fatal("no companies in test db")
	}
	companyID := co[0].ID

	// Insert first discount
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID,
		ModelType: "deepseek-chat",
		Discount:  0.7,
	})

	// Insert newer discount for same model (append-only, later effective_from)
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID,
		ModelType: "deepseek-chat",
		Discount:  0.85,
	})

	// Query should return latest only
	rows, err := st.ModelDiscount().CurrentDiscounts(ctx, companyID)
	if err != nil {
		t.Fatal(err)
	}

	var dsDiscount float64
	for _, r := range rows {
		if r.ModelType == "deepseek-chat" {
			dsDiscount = r.Discount
		}
	}
	if dsDiscount != 0.85 {
		t.Errorf("deepseek-chat discount = %f, want 0.85 (latest)", dsDiscount)
	}
}
