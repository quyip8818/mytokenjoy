package company_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestForEachActiveCompanySkipsTesting(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	id9100 := uuid.MustParse("00000000-0000-7000-0000-000000009100")
	id9101 := uuid.MustParse("00000000-0000-7000-0000-000000009101")

	// Create a testing company.
	if err := st.Company().Create(ctx, store.Company{
		ID: id9100, Name: "Testing Co", Type: store.CompanyTypeTesting,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	// Create a standard company.
	if err := st.Company().Create(ctx, store.Company{
		ID: id9101, Name: "Standard Co", Type: store.CompanyTypeStandard,
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	var visited []uuid.UUID
	err := domaincompany.ForEachActiveCompany(ctx, st.Company(), func(_ context.Context, co store.Company) error {
		visited = append(visited, co.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range visited {
		if id == id9100 {
			t.Fatal("ForEachActiveCompany should skip testing companies")
		}
	}

	found := false
	for _, id := range visited {
		if id == id9101 {
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

	id9200 := uuid.MustParse("00000000-0000-7000-0000-000000009200")

	if err := st.Company().Create(ctx, store.Company{
		ID: id9200, Name: "Suspended Co", Type: store.CompanyTypeStandard,
		Status: store.CompanyStatusSuspended, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	var visited []uuid.UUID
	err := domaincompany.ForEachActiveCompany(ctx, st.Company(), func(_ context.Context, co store.Company) error {
		visited = append(visited, co.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range visited {
		if id == id9200 {
			t.Fatal("ForEachActiveCompany should skip suspended companies")
		}
	}
}

func TestContextFromStoreIncludesType(t *testing.T) {
	t.Parallel()
	id42 := uuid.MustParse("00000000-0000-7000-0000-000000000042")
	co := store.Company{
		ID:     id42,
		Name:   "Test",
		Type:   store.CompanyTypeDemo,
		Status: store.CompanyStatusActive,
	}
	info := domaincompany.ContextFromStore(co)
	if info.Type != store.CompanyTypeDemo {
		t.Fatalf("expected type=%s, got %s", store.CompanyTypeDemo, info.Type)
	}
	if info.CompanyID != id42 {
		t.Fatalf("expected companyID=%s, got %s", id42, info.CompanyID)
	}
}
