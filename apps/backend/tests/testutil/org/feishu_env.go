//go:build testhook

package orgfix

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

type FeishuOrgEnv struct {
	Cfg       config.Config
	Store     store.Store
	Svc       org.Service
	ServerURL string
}

func SetupFeishuConnected(t *testing.T) FeishuOrgEnv {
	t.Helper()
	server := testutil.StartFeishuMockServer(t)
	cfg := testutil.TestConfig()
	cfg.FeishuBaseURL = server.URL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	testutil.ConnectFeishuDataSource(t, &cfg, st, server.URL)
	return FeishuOrgEnv{
		Cfg: cfg, Store: st,
		Svc:       NewService(t, cfg, st),
		ServerURL: server.URL,
	}
}

func ImportFeishuOrg(t *testing.T, env FeishuOrgEnv) types.ImportResult {
	t.Helper()
	result, err := env.Svc.ImportDataSource(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func SetupImportedFeishuOrg(t *testing.T) FeishuOrgEnv {
	t.Helper()
	env := SetupFeishuConnected(t)
	ImportFeishuOrg(t, env)
	return env
}

func SetupImportedFeishuOrgWithServer(t *testing.T, serverURL string) FeishuOrgEnv {
	t.Helper()
	cfg := testutil.TestConfig()
	cfg.FeishuBaseURL = serverURL
	_, st := testutil.NewTestStore(t, testutil.WithConfig(cfg))
	testutil.ConnectFeishuDataSource(t, &cfg, st, serverURL)
	env := FeishuOrgEnv{
		Cfg: cfg, Store: st,
		Svc:       NewService(t, cfg, st),
		ServerURL: serverURL,
	}
	ImportFeishuOrg(t, env)
	return env
}

func WithSyncConfig(t *testing.T, env FeishuOrgEnv, cfg types.SyncConfig) FeishuOrgEnv {
	t.Helper()
	integration, err := env.Store.Org().Integration(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	integration.ApplySyncConfig(cfg)
	if err := env.Store.Org().SetIntegration(testutil.Ctx(), integration); err != nil {
		t.Fatal(err)
	}
	env.Svc = NewService(t, env.Cfg, env.Store)
	return env
}
