//go:build testhook

package newapisync

import (
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type TestServiceOpts struct {
	Stub   *mock.StubAdminClient
	Wallet company.WalletService
}

func NewTestService(t *testing.T, opts TestServiceOpts) (*domainnewapisync.NewAPISync, config.Config, store.Store) {
	t.Helper()
	stub := opts.Stub
	if stub == nil {
		stub = &mock.StubAdminClient{Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: 1000}}
	}
	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	sync := domainnewapisync.New(
		cfg,
		st,
		newapi.NewAdminPortAdapter(stub),
		opts.Wallet,
		domainnewapisync.NewChannelPolicy(cfg),
		app.NewNewAPISyncEnqueuer(riverfix.NewInsertOnlyEnqueuer(t, cfg, st)),
	)
	return sync, cfg, st
}

type PendingPlatformKeyOpts struct {
	ID             string
	Name           string
	MemberID       string
	DepartmentID   string
	Budget         float64
	ModelWhitelist []int64
}

func SeedPendingPlatformKey(t *testing.T, st store.Store, opts PendingPlatformKeyOpts) types.PlatformKey {
	t.Helper()
	if opts.ID == "" {
		opts.ID = "plk-test"
	}
	if opts.Name == "" {
		opts.Name = "test-key"
	}
	if opts.MemberID == "" {
		opts.MemberID = contract.IDMember1
	}
	if opts.DepartmentID == "" {
		opts.DepartmentID = contract.IDDept3
	}
	if opts.Budget == 0 {
		opts.Budget = 1000
	}
	if len(opts.ModelWhitelist) == 0 {
		opts.ModelWhitelist = []int64{contract.IDModel1}
	}

	ctx := testutil.CtxForCompany(contract.DefaultCompanyID)
	memberID := opts.MemberID
	key := types.PlatformKey{
		ID:             opts.ID,
		Name:           opts.Name,
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Status:         "active",
		Budget:         opts.Budget,
		ModelWhitelist: opts.ModelWhitelist,
		CreatedAt:      "2026-06-19",
	}
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	keys = append(keys, key)
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	if err := st.PlatformKeyMappings().UpsertMapping(ctx, store.PlatformKeyMapping{
		CompanyID:     contract.DefaultCompanyID,
		PlatformKeyID: key.ID,
		MemberID:      &memberID,
		DepartmentID:  opts.DepartmentID,
		SyncStatus:    store.MappingSyncStatusPending,
	}); err != nil {
		t.Fatal(err)
	}
	return key
}
