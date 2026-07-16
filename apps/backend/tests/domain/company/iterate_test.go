package company_test

import (
	"context"
	"testing"
	"time"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestForEachActiveCompanySkipsTesting(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	// Create a testing company.
	if err := st.Company().Create(ctx, store.Company{
		ID: 9100, Name: "Testing Co", Type: store.CompanyTypeTesting,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	// Create a standard company.
	if err := st.Company().Create(ctx, store.Company{
		ID: 9101, Name: "Standard Co", Type: store.CompanyTypeStandard,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	var visited []int64
	err := domaincompany.ForEachActiveCompany(ctx, st.Company(), func(_ context.Context, co store.Company) error {
		visited = append(visited, co.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range visited {
		if id == 9100 {
			t.Fatal("ForEachActiveCompany should skip testing companies")
		}
	}

	found := false
	for _, id := range visited {
		if id == 9101 {
			found = true
		}
	}
	if !found {
		t.Fatal("ForEachActiveCompany should visit standard companies")
	}
}

func TestForEachActiveCompanySkipsSuspended(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	if err := st.Company().Create(ctx, store.Company{
		ID: 9200, Name: "Suspended Co", Type: store.CompanyTypeStandard,
		Status: store.CompanyStatusSuspended, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	var visited []int64
	err := domaincompany.ForEachActiveCompany(ctx, st.Company(), func(_ context.Context, co store.Company) error {
		visited = append(visited, co.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range visited {
		if id == 9200 {
			t.Fatal("ForEachActiveCompany should skip suspended companies")
		}
	}
}

func TestContextFromStoreIncludesType(t *testing.T) {
	t.Parallel()
	co := store.Company{
		ID:     42,
		Name:   "Test",
		Type:   store.CompanyTypeDemo,
		Status: store.CompanyStatusActive,
	}
	info := domaincompany.ContextFromStore(co)
	if info.Type != store.CompanyTypeDemo {
		t.Fatalf("expected type=%s, got %s", store.CompanyTypeDemo, info.Type)
	}
	if info.CompanyID != 42 {
		t.Fatalf("expected companyID=42, got %d", info.CompanyID)
	}
}
