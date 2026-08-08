//go:build testhook

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

func createDiscountTestCompany(t *testing.T, st store.Store) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	ctx := context.Background()
	err := st.Company().Create(ctx, store.Company{
		ID:     id,
		Name:   "discount-test-" + id.String()[:8],
		Type:   store.CompanyTypeSaas,
		Status: store.CompanyStatusActive,
	})
	if err != nil {
		t.Fatalf("create test company: %v", err)
	}
	return id
}

func TestModelDiscountInsertAndQuery(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := context.Background()
	companyID := createDiscountTestCompany(t, st)

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
	if len(rows) != 2 {
		t.Fatalf("expected 2 discount rows, got %d", len(rows))
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
	companyID := createDiscountTestCompany(t, st)

	now := time.Now()

	// Insert first discount (older)
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID:     companyID,
		ModelType:     "deepseek-chat",
		Discount:      0.7,
		EffectiveFrom: now.Add(-time.Minute),
	})

	// Insert newer discount for same model
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID:     companyID,
		ModelType:     "deepseek-chat",
		Discount:      0.85,
		EffectiveFrom: now,
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

func TestModelDiscountDeleteByCompanyAndModel(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := context.Background()
	companyID := createDiscountTestCompany(t, st)

	// Insert discounts for two models
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID, ModelType: "gpt-4o", Discount: 0.8,
	})
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID, ModelType: "claude-3", Discount: 0.9,
	})

	// Delete only gpt-4o
	if err := st.ModelDiscount().DeleteByCompanyAndModel(ctx, companyID, "gpt-4o"); err != nil {
		t.Fatalf("DeleteByCompanyAndModel: %v", err)
	}

	rows, err := st.ModelDiscount().CurrentDiscounts(ctx, companyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row remaining, got %d", len(rows))
	}
	if rows[0].ModelType != "claude-3" {
		t.Errorf("expected claude-3 remaining, got %s", rows[0].ModelType)
	}
}

func TestModelDiscountDeleteAllByCompany(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := context.Background()
	companyID := createDiscountTestCompany(t, st)

	// Insert several discounts
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID, ModelType: "gpt-4o", Discount: 0.7,
	})
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID, ModelType: "claude-3", Discount: 0.85,
	})
	_ = st.ModelDiscount().Insert(ctx, store.ModelDiscountRow{
		CompanyID: companyID, ModelType: "deepseek-chat", Discount: 1.2,
	})

	// Delete all
	if err := st.ModelDiscount().DeleteAllByCompany(ctx, companyID); err != nil {
		t.Fatalf("DeleteAllByCompany: %v", err)
	}

	rows, err := st.ModelDiscount().CurrentDiscounts(ctx, companyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows after DeleteAll, got %d", len(rows))
	}
}
