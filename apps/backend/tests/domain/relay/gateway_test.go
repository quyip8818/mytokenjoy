package relay_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	relayfix "github.com/tokenjoy/backend/tests/testutil/relay"
)

func TestGatewayRejectsNoAuth(t *testing.T) {
	t.Parallel()
	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGatewayRejectsInvalidToken(t *testing.T) {
	t.Parallel()
	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-invalid-token")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGatewayRejectsNonBearerAuth(t *testing.T) {
	t.Parallel()
	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGatewayProxiesValidRequest(t *testing.T) {
	t.Parallel()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-123"}`))
	}))
	defer backend.Close()

	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{
		Budget:          1000,
		WalletQuota:     999999,
		ProxyBackendURL: backend.URL,
	})

	req := relayfix.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestGatewayRejectsSuspendedCompany(t *testing.T) {
	t.Parallel()
	scenario := relayfix.BuildGatewayScenario(t, relayfix.GatewayScenarioOpts{
		Budget:        1000,
		CompanyStatus: "suspended",
	})

	req := relayfix.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected non-200 for suspended company")
	}
}
