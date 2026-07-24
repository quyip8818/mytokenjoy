//go:build testhook

package provision_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	newapisynctf "github.com/tokenjoy/backend/tests/testutil/newapisync"
)

func TestBootstrapLeavesAllPlatformKeysReady(t *testing.T) {
	t.Parallel()
	var nextTokenID int64 = 600
	stub := &mock.StubAdminClient{
		CreateTokenFn: func(_ context.Context, _ adminport.CreateTokenInput) (adminport.TokenResult, error) {
			nextTokenID++
			return adminport.TokenResult{ID: nextTokenID, Key: fmt.Sprintf("sk-readiness-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (adminport.TokenResult, error) {
			return adminport.TokenResult{ID: tokenID, Key: fmt.Sprintf("sk-readiness-%d", tokenID), RemainQuota: 1000}, nil
		},
	}
	sync, _ := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	unready, err := sync.UnreadyPlatformKeyIDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(unready) != 0 {
		t.Fatalf("expected no unready keys after bootstrap, got %v", unready)
	}
}

func TestDevReadinessOKAfterBootstrap(t *testing.T) {
	t.Parallel()
	var nextTokenID int64 = 610
	stub := &mock.StubAdminClient{
		CreateTokenFn: func(_ context.Context, _ adminport.CreateTokenInput) (adminport.TokenResult, error) {
			nextTokenID++
			return adminport.TokenResult{ID: nextTokenID, Key: fmt.Sprintf("sk-readiness-http-%d", nextTokenID), RemainQuota: 1000}, nil
		},
		GetTokenFn: func(_ context.Context, tokenID int64) (adminport.TokenResult, error) {
			return adminport.TokenResult{ID: tokenID, Key: fmt.Sprintf("sk-readiness-http-%d", tokenID), RemainQuota: 1000}, nil
		},
	}
	sync, st := newapisynctf.NewLocalTestService(t, stub)
	ctx := testutil.CtxForCompany(contract.LocalCompanyID)
	if err := sync.Bootstrap(ctx, contract.LocalCompanyID); err != nil {
		t.Fatal(err)
	}
	_ = st

	router := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithDeployEnv(config.DeployEnvLocal)(cfg)
	}).Router
	rec := testhttp.ServeAuthz(t, router, http.MethodGet, "/api/dev/readiness", "", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected readiness 200 after bootstrap, got %d body=%s", rec.Code, rec.Body.String())
	}
}
