package authz_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

type countingOrgRepo struct {
	store.OrgRepository
	calls int
}

func (r *countingOrgRepo) GetMemberAuthz(ctx context.Context, companyID int64, memberID string) (*store.MemberAuthz, error) {
	r.calls++
	return r.OrgRepository.GetMemberAuthz(ctx, companyID, memberID)
}

type countingStore struct {
	store.Store
	org *countingOrgRepo
}

func (s *countingStore) Org() store.OrgRepository {
	return s.org
}

func newCountingAuthzService(t *testing.T) (authz.Service, *countingOrgRepo) {
	t.Helper()
	cfg := testutil.TestConfig()
	_, base := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	orgRepo := &countingOrgRepo{OrgRepository: base.Org()}
	wrapped := &countingStore{Store: base, org: orgRepo}
	return authz.NewService(cfg, wrapped), orgRepo
}

func TestGetSessionContextCachesByRevision(t *testing.T) {
	svc, orgRepo := newCountingAuthzService(t)
	ctx := testutil.Ctx()
	companyID := seed.DefaultCompanyID
	memberID := seed.IDMemberAdmin

	if _, err := svc.GetSessionContext(ctx, companyID, memberID); err != nil {
		t.Fatalf("first GetSessionContext: %v", err)
	}
	if _, err := svc.GetSessionContext(ctx, companyID, memberID); err != nil {
		t.Fatalf("second GetSessionContext: %v", err)
	}
	if orgRepo.calls != 1 {
		t.Fatalf("expected GetMemberAuthz called once, got %d", orgRepo.calls)
	}
}
