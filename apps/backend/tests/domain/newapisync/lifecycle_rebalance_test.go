//go:build testhook

package newapisync_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

type stubWallet struct {
	quota int64
}

func (s *stubWallet) AvailableNewAPIUnits(context.Context, int64) (int64, error) {
	return s.quota, nil
}

func (s *stubWallet) FreshNewAPIUnits(ctx context.Context, id int64) (int64, error) {
	return s.AvailableNewAPIUnits(ctx, id)
}

func (s *stubWallet) InvalidateNewAPIUnits(int64) {}

func TestTrySyncCreateCapsRemainQuotaByWallet(t *testing.T) {
	t.Parallel()

	const walletCap int64 = 50
	var lastRemain int64
	const platformKeyID = "plk-wallet-cap"
	stub := &mock.StubAdminClient{
		Token: newapi.Token{ID: 1, Key: "sk-test", RemainQuota: walletCap},
		CreateTokenFn: func(_ context.Context, req newapi.CreateTokenRequest) (newapi.Token, error) {
			lastRemain = req.RemainQuota
			if req.Name != "tokenjoy:"+platformKeyID {
				t.Fatalf("expected unique token name tokenjoy:%s, got %q", platformKeyID, req.Name)
			}
			return newapi.Token{ID: 1, Key: "sk-test", RemainQuota: req.RemainQuota}, nil
		},
	}

	sync, _, st := newapisynctf.NewTestService(t, newapisynctf.TestServiceOpts{
		Stub:   stub,
		Wallet: &stubWallet{quota: walletCap},
	})

	const walletUserID int64 = 501
	ctx := testutil.CtxForCompany(contract.DefaultCompanyID)
	if err := st.Company().UpdateNewAPIWalletUserID(ctx, contract.DefaultCompanyID, walletUserID); err != nil {
		t.Fatal(err)
	}

	key := newapisynctf.SeedPendingPlatformKey(t, st, newapisynctf.PendingPlatformKeyOpts{
		ID:   platformKeyID,
		Name: "wallet-cap-key",
	})

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
