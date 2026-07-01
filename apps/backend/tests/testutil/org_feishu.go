package testutil

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type FeishuOrgEnv struct {
	Cfg       config.Config
	Store     store.Store
	Svc       org.Service
	ServerURL string
}

func SetupFeishuConnected(t *testing.T) FeishuOrgEnv {
	t.Helper()
	server := StartFeishuMockServer(t)
	cfg := TestConfig()
	cfg.FeishuBaseURL = server.URL
	st := NewMemoryStore(t, cfg)
	ConnectFeishuDataSource(t, &cfg, st, server.URL)
	return FeishuOrgEnv{
		Cfg: cfg, Store: st,
		Svc:       NewOrgService(t, cfg, st),
		ServerURL: server.URL,
	}
}

func ImportFeishuOrg(t *testing.T, env FeishuOrgEnv) types.ImportResult {
	t.Helper()
	result, err := env.Svc.ImportDataSource(Ctx())
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
	cfg := TestConfig()
	cfg.FeishuBaseURL = serverURL
	st := NewMemoryStore(t, cfg)
	ConnectFeishuDataSource(t, &cfg, st, serverURL)
	env := FeishuOrgEnv{
		Cfg: cfg, Store: st,
		Svc:       NewOrgService(t, cfg, st),
		ServerURL: serverURL,
	}
	ImportFeishuOrg(t, env)
	return env
}

func WithSyncConfig(t *testing.T, env FeishuOrgEnv, cfg types.SyncConfig) FeishuOrgEnv {
	t.Helper()
	if err := env.Store.Org().SetSyncConfig(Ctx(), cfg); err != nil {
		t.Fatal(err)
	}
	env.Svc = NewOrgService(t, env.Cfg, env.Store)
	return env
}
