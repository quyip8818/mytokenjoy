package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

// PRD 10.3 API 调用校验顺序:
// 1. Key 有效性 2. Key 启用状态 3. 模型白名单 4. 额度充足 5. 转发供应商

func TestGatewayCheckOrder_InvalidKey(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 1000,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-nonexistent-key")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("invalid key should return 403, got %d", w.Code)
	}
}

func TestGatewayCheckOrder_DisabledKey(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 1000,
	})
	ctx := testutil.Ctx()

	mapping, err := scenario.Store.PlatformKeyMappings().GetMappingByKeyHash(ctx, store.HashPlatformKey(scenario.FullKey))
	if err != nil || mapping == nil {
		t.Fatal("expected platform key mapping")
	}

	keys, err := scenario.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	disabled := false
	for i := range keys {
		if keys[i].ID == mapping.PlatformKeyID {
			keys[i].Status = "disabled"
			disabled = true
		}
	}
	if !disabled {
		t.Fatalf("platform key %s not found", mapping.PlatformKeyID)
	}
	if err := scenario.Store.Keys().SetPlatformKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}

	req := gatewaytf.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("disabled key should not pass gateway check")
	}
}

func TestGatewayCheckOrder_ModelNotInWhitelist(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget: 1000,
	})

	req := gatewaytf.GatewayRequestWithModel(scenario.FullKey, "unknown-model-xyz")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("model not in whitelist should not pass gateway, got %d", w.Code)
	}
}

func TestGatewayCheckOrder_AllowsWhenDeptBudgetZero(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:   0,
		Consumed: 0,
	})

	req := gatewaytf.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("gateway should not block on dept budget, got %d", w.Code)
	}
}

func TestGatewayCheckOrder_SuccessfulProxy(t *testing.T) {
	t.Parallel()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header to be forwarded")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}]}`))
	}))
	defer backend.Close()

	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:          1000,
		ProxyBackendURL: backend.URL,
	})

	req := gatewaytf.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("valid request should proxy to backend, got %d: %s", w.Code, w.Body.String())
	}
}
