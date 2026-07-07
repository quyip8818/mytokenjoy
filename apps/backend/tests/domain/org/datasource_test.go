package org_test

import (
	"errors"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestTestDataSourceInvalidCredential422(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	server := testutil.StartFeishuAuthErrorServer(t)
	cfg.FeishuBaseURL = server.URL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	svc := orgfix.NewService(t, cfg, st)

	result, err := svc.TestDataSource(testutil.Ctx(), types.Credential{
		Platform: types.PlatformFeishu,
		Feishu: &types.FeishuCredential{
			Platform: types.PlatformFeishu, AppID: "bad", AppSecret: "bad",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("expected failed test result")
	}
	if result.Message == nil || *result.Message == "" {
		t.Fatal("expected failure message for invalid credential")
	}
}

func TestUpdateDataSourcePersistsCredential(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	server := testutil.StartFeishuMockServer(t)
	cfg.FeishuBaseURL = server.URL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	svc := orgfix.NewService(t, cfg, st)

	cred := types.Credential{
		Platform: types.PlatformFeishu,
		Feishu: &types.FeishuCredential{
			Platform: types.PlatformFeishu, AppID: "cli_test", AppSecret: "secret_test",
		},
	}
	if err := svc.UpdateDataSource(testutil.Ctx(), cred, false); err != nil {
		t.Fatal(err)
	}
	stored, err := st.Org().GetIntegrationCredential(testutil.Ctx())
	if err != nil || stored == nil {
		t.Fatalf("expected stored credential, err=%v stored=%v", err, stored)
	}
	status, err := svc.GetDataSourceStatus(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Connected || status.Platform == nil || *status.Platform != types.PlatformFeishu {
		t.Fatalf("unexpected status %+v", status)
	}
}

func TestSearchDataSourceUsesProvider(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	server := testutil.StartFeishuMockServer(t)
	cfg.FeishuBaseURL = server.URL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	testutil.ConnectFeishuDataSource(t, &cfg, st, server.URL)
	svc := orgfix.NewService(t, cfg, st)

	result, err := svc.SearchDataSource(testutil.Ctx(), "Mock")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Mock User" {
		t.Fatalf("unexpected search result %+v", result)
	}
}

func TestUnsupportedPlatformReturns422(t *testing.T) {
	t.Parallel()
	_, _, svc := orgfix.NewServiceFromStore(t)
	_, err := svc.TestDataSource(testutil.Ctx(), types.Credential{
		Platform: types.PlatformDingtalk,
		Dingtalk: &types.DingtalkCredential{
			Platform: types.PlatformDingtalk, CorpID: "c", AppKey: "k", AppSecret: "s",
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %v", err)
	}
}

func TestUpdateDataSourceSwitchPlatformRequiresForce(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	server := testutil.StartFeishuMockServer(t)
	cfg.FeishuBaseURL = server.URL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	testutil.ConnectFeishuDataSource(t, &cfg, st, server.URL)
	svc := orgfix.NewService(t, cfg, st)

	err := svc.UpdateDataSource(testutil.Ctx(), types.Credential{
		Platform: types.PlatformDingtalk,
		Dingtalk: &types.DingtalkCredential{
			Platform: types.PlatformDingtalk, CorpID: "c", AppKey: "k", AppSecret: "s",
		},
	}, false)
	if err == nil {
		t.Fatal("expected platform switch without force to fail")
	}
	var domainErr *domain.DomainError
	if !errors.As(err, &domainErr) || domainErr.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %v", err)
	}
}
