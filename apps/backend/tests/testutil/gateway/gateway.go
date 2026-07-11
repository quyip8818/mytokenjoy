//go:build testhook

package gatewayfix

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

type GatewayScenarioOpts struct {
	CompanyID          int64
	WalletBalancePoint *float64
	NewAPIWalletUserID int64
	DepartmentID       string
	Budget             float64
	Consumed           float64
	RemainQuota        int64
	CompanyStatus      string
	ProxyBackendURL    string
}

type GatewayScenario struct {
	Gateway domaingateway.GatewayService
	Store   store.Store
	Cfg     config.Config
	FullKey string
}

func BuildGatewayScenario(t *testing.T, opts GatewayScenarioOpts) GatewayScenario {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	fullKey := ConfigureGatewayStore(t, cfg, st, opts)

	backendURL := opts.ProxyBackendURL
	if backendURL == "" {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(backend.Close)
		backendURL = backend.URL
	}
	cfg.NewAPIBaseURL = backendURL
	cfg.GatewayEnabled = true

	precheck := NewPrecheckService(cfg, st)
	gw, err := domaingateway.NewGatewayService(cfg, precheck)
	if err != nil {
		t.Fatal(err)
	}
	return GatewayScenario{Gateway: gw, Store: st, Cfg: cfg, FullKey: fullKey}
}

func NewPrecheckService(cfg config.Config, st store.Store) *domaingateway.PrecheckService {
	return domaingateway.NewPrecheckService(st.GatewayPrecheck(), cfg.Clock())
}

func setBudgetOnTree(nodes []types.BudgetNode, deptID string, budget, consumed float64) bool {
	for i := range nodes {
		if nodes[i].ID == deptID {
			nodes[i].Budget = budget
			if consumed > 0 {
				nodes[i].Consumed = consumed
			}
			return true
		}
		if len(nodes[i].Children) > 0 && setBudgetOnTree(nodes[i].Children, deptID, budget, consumed) {
			return true
		}
	}
	return false
}

func GatewayRequest(fullKey string) *http.Request {
	return GatewayRequestWithModel(fullKey, "gpt-4o")
}

func GatewayRequestWithModel(fullKey, model string) *http.Request {
	body, _ := json.Marshal(map[string]string{"model": model})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+fullKey)
	req.Header.Set("Content-Type", "application/json")
	return req
}
