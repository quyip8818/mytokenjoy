//go:build testhook

package newapisync

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type TestServiceOpts struct {
	Stub           *mock.StubAdminClient
	SkipWalletSeed bool
}

const TestWalletUserID int64 = 501

func EnsureWalletUserID(t *testing.T, st store.Store, companyID uuid.UUID, walletUserID int64) {
	t.Helper()
	if err := st.Company().UpdateNewAPIWalletUserID(
		testutil.CtxForCompany(companyID),
		companyID,
		walletUserID,
	); err != nil {
		t.Fatal(err)
	}
}

func NewTestService(t *testing.T, opts TestServiceOpts) (*domainnewapisync.NewAPISync, config.Config, store.Store) {
	t.Helper()
	return newTestService(t, opts, nil)
}

func NewLocalTestService(t *testing.T, stub *mock.StubAdminClient, cfgOpts ...testutil.ConfigOption) (*domainnewapisync.NewAPISync, store.Store) {
	t.Helper()
	if stub != nil && stub.User.ID == 0 {
		stub.User.ID = TestWalletUserID
	}
	base := []testutil.ConfigOption{testutil.WithDeployEnv("local")}
	sync, _, st := newTestService(t, TestServiceOpts{Stub: stub, SkipWalletSeed: true}, append(base, cfgOpts...))
	return sync, st
}

func newTestService(t *testing.T, opts TestServiceOpts, cfgOpts []testutil.ConfigOption) (*domainnewapisync.NewAPISync, config.Config, store.Store) {
	t.Helper()
	stub := opts.Stub
	if stub == nil {
		stub = &mock.StubAdminClient{Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: 1000}}
	}
	base := []testutil.ConfigOption{
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	}
	cfg, st := testutil.NewTestStore(t, append(base, cfgOpts...)...)
	if !opts.SkipWalletSeed {
		EnsureWalletUserID(t, st, contract.DefaultCompanyID, TestWalletUserID)
	}
	sync := domainnewapisync.New(
		cfg,
		st,
		newapi.NewAdminPortAdapter(stub),
		policy.NewChannelPolicy(cfg),
		adapter.NewNewAPISyncEnqueuer(riverfix.NewInsertOnlyEnqueuer(t, cfg, st)),
	)
	return sync, cfg, st
}

type PendingPlatformKeyOpts struct {
	ID             uuid.UUID
	Name           string
	MemberID       uuid.UUID
	DepartmentID   uuid.UUID
	Budget         int64
	ModelWhitelist []uuid.UUID
}

func SeedPendingPlatformKey(t *testing.T, st store.Store, opts PendingPlatformKeyOpts) types.PlatformKey {
	t.Helper()
	if opts.ID == uuid.Nil {
		opts.ID = uuid.MustParse("00000000-0000-7000-0000-00000000f099")
	}
	if opts.Name == "" {
		opts.Name = "test-key"
	}
	if opts.MemberID == uuid.Nil {
		opts.MemberID = contract.IDMember1
	}
	if opts.DepartmentID == uuid.Nil {
		opts.DepartmentID = contract.IDDept3
	}
	if opts.Budget == 0 {
		opts.Budget = 1000
	}
	if len(opts.ModelWhitelist) == 0 {
		opts.ModelWhitelist = []uuid.UUID{contract.IDModel1}
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
