package relayfix

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	orgfix "github.com/tokenjoy/backend/tests/testutil/org"
	"github.com/tokenjoy/backend/tests/testutil/saas"
)

type GatewayScenarioOpts struct {
	CompanyID          int64
	WalletQuota        int64
	NewAPIWalletUserID int64
	DepartmentID       string
	Budget             float64
	Consumed           float64
	RemainQuota        int64
	CompanyStatus      string
	UseRealWallet      bool
	NewAPIMock         *saas.NewAPIMock
	ProxyBackendURL    string
}

type GatewayScenario struct {
	Gateway domainrelay.GatewayService
	Store   store.Store
	Cfg     config.Config
	FullKey string
}

func ConfigureGatewayStore(t *testing.T, st store.Store, opts GatewayScenarioOpts) string {
	t.Helper()
	if opts.CompanyID == 0 {
		opts.CompanyID = seed.DefaultCompanyID
	}
	if opts.DepartmentID == "" {
		opts.DepartmentID = seed.IDDept3
	}
	if opts.NewAPIWalletUserID == 0 {
		opts.NewAPIWalletUserID = 99
	}
	if opts.RemainQuota == 0 {
		opts.RemainQuota = 10000
	}
	if opts.CompanyStatus == "" {
		opts.CompanyStatus = store.CompanyStatusActive
	}

	ctx := testutil.CtxForCompany(opts.CompanyID)
	if err := st.Company().UpdateNewAPIWalletUserID(ctx, opts.CompanyID, opts.NewAPIWalletUserID); err != nil {
		t.Fatal(err)
	}
	if opts.CompanyStatus != store.CompanyStatusActive {
		if err := st.Company().UpdateStatus(ctx, opts.CompanyID, opts.CompanyStatus); err != nil {
			t.Fatal(err)
		}
	}

	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if !setBudgetOnTree(tree, opts.DepartmentID, opts.Budget, opts.Consumed) {
		t.Fatalf("department %s not found in budget tree", opts.DepartmentID)
	}
	if err := orgfix.PersistBudgetTree(ctx, st, tree); err != nil {
		t.Fatal(err)
	}

	fullKey := "sk-test-gateway-key"
	keys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	platformKeyID := seed.IDPlatformKey1
	if len(keys) == 0 {
		keys = []types.PlatformKey{{
			ID:        "plk-gateway-test",
			Name:      "Gateway Test Key",
			KeyPrefix: "sk-test",
			FullKey:   &fullKey,
			Status:    "active",
			CreatedAt: "2026-06-19",
		}}
		platformKeyID = keys[0].ID
	} else {
		found := false
		for i := range keys {
			if keys[i].ID == seed.IDPlatformKey1 {
				keys[i].FullKey = &fullKey
				keys[i].Status = "active"
				found = true
			}
		}
		if !found {
			keys[0].FullKey = &fullKey
			keys[0].Status = "active"
			platformKeyID = keys[0].ID
		}
	}
	if err := st.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	tokenID := int64(42)
	remain := opts.RemainQuota
	if err := st.Relay().UpsertMapping(ctx, store.RelayMapping{
		CompanyID:              opts.CompanyID,
		PlatformKeyID:          platformKeyID,
		NewAPITokenID:          &tokenID,
		MemberID:               testutil.StrPtr(seed.IDMember1),
		DepartmentID:           opts.DepartmentID,
		SyncStatus:             store.RelaySyncStatusSynced,
		RelayGroup:             "dept-dept-3",
		NewAPITokenRemainQuota: &remain,
	}); err != nil {
		t.Fatal(err)
	}
	return fullKey
}

func BuildGatewayScenario(t *testing.T, opts GatewayScenarioOpts) GatewayScenario {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	fullKey := ConfigureGatewayStore(t, st, opts)

	backendURL := opts.ProxyBackendURL
	if backendURL == "" {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(backend.Close)
		backendURL = backend.URL
	}
	cfg.NewAPIBaseURL = backendURL
	cfg.RelayGatewayEnabled = true

	wallet := gatewayWallet(cfg, opts)
	precheck := domainrelay.NewPrecheckService(st.Org().Nodes(), st.Keys(), wallet)
	gw, err := domainrelay.NewGatewayService(cfg, st.Relay(), st.Company(), precheck)
	if err != nil {
		t.Fatal(err)
	}
	return GatewayScenario{Gateway: gw, Store: st, Cfg: cfg, FullKey: fullKey}
}

func gatewayWallet(cfg config.Config, opts GatewayScenarioOpts) company.WalletService {
	if opts.UseRealWallet && opts.NewAPIMock != nil {
		opts.NewAPIMock.SetQuota(opts.NewAPIWalletUserID, opts.WalletQuota)
		opts.NewAPIMock.ApplyToConfig(&cfg)
		client := newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
		return company.NewWalletService(cfg, client)
	}
	return NewStubWallet(opts.WalletQuota)
}

func setBudgetOnTree(nodes []types.BudgetNode, deptID string, budget, consumed float64) bool {
	for i := range nodes {
		if nodes[i].ID == deptID {
			nodes[i].Budget = budget
			nodes[i].Consumed = consumed
			return true
		}
		if len(nodes[i].Children) > 0 && setBudgetOnTree(nodes[i].Children, deptID, budget, consumed) {
			return true
		}
	}
	return false
}

func GatewayRequest(fullKey string) *http.Request {
	body, _ := json.Marshal(map[string]string{"model": "gpt-4o"})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+fullKey)
	return req
}
