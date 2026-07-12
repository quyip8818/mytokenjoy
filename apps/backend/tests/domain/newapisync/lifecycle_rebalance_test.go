//go:build testhook

package newapisync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

type stubWallet struct {
	quota int64
}

func (s *stubWallet) AvailableQuota(context.Context, int64) (int64, error) {
	return s.quota, nil
}

func TestTrySyncCreateCapsRemainQuotaByWallet(t *testing.T) {
	t.Parallel()

	const walletCap int64 = 50
	var lastRemain int64
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: walletCap},
		CreateTokenFn: func(_ context.Context, req newapi.CreateTokenRequest) (newapi.Token, error) {
			lastRemain = req.RemainQuota
			return newapi.Token{ID: 1, Key: "sk-test", RemainQuota: req.RemainQuota}, nil
		},
	}

	cfg, st := testutil.NewTestStore(t,
		testutil.WithNewAPIEnabled(true),
		testutil.WithNewAPIBaseURL("http://newapi.test"),
		testutil.WithNewAPIAdminToken("token"),
	)
	const walletUserID int64 = 501
	ctx := testutil.CtxForCompany(contract.DefaultCompanyID)
	if err := st.Company().UpdateNewAPIWalletUserID(ctx, contract.DefaultCompanyID, walletUserID); err != nil {
		t.Fatal(err)
	}

	sync := newapisync.New(
		cfg,
		st,
		newapi.NewAdminPortAdapter(stub),
		&stubWallet{quota: walletCap},
		newapisync.NewChannelPolicy(config.Config{}),
		riverfix.NewInsertOnlyEnqueuer(t, cfg, st),
	)

	memberID := contract.IDMember1
	key := types.PlatformKey{
		ID:             "plk-wallet-cap",
		Name:           "wallet-cap-key",
		MemberID:       &memberID,
		Status:         "active",
		Budget:         1000,
		ModelWhitelist: []int64{contract.IDModel1},
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
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.MappingSyncStatusPending,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := sync.TrySyncCreate(ctx, key.ID); err != nil {
		t.Fatalf("TrySyncCreate: %v", err)
	}
	if stub.CreateTokenCalls != 1 {
		t.Fatalf("expected one CreateToken call, got %d", stub.CreateTokenCalls)
	}
	if lastRemain != walletCap {
		t.Fatalf("expected remain quota capped to wallet %d, got %d", walletCap, lastRemain)
	}
}
