package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	gatewaytf "github.com/tokenjoy/backend/tests/testutil/gateway"
)

func TestGatewayRejectsNoAuth(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGatewayRejectsInvalidToken(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-invalid-token")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGatewayAcceptsNewAPIStyleToken(t *testing.T) {
	t.Parallel()
	const newAPIKey = "pKozjXrW57MlfHG27zHRLVuVeDLCpkGPCCXtdNSRFkliAGDQ"

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-newapi-key"}`))
	}))
	defer backend.Close()

	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:          1000,
		ProxyBackendURL: backend.URL,
		FullKey:         newAPIKey,
	})

	req := gatewaytf.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for NewAPI-style token, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestGatewayRejectsNonBearerAuth(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
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

	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:          1000,
		ProxyBackendURL: backend.URL,
	})

	req := gatewaytf.GatewayRequest(scenario.FullKey)
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestGatewayAllowsModelsListing(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+scenario.FullKey)

	rec := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /v1/models, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGatewayAllowsDevModelInLocal(t *testing.T) {
	t.Parallel()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"chatcmpl-dev"}`))
	}))
	defer backend.Close()

	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
		Budget:          1000,
		ProxyBackendURL: backend.URL,
		DeployEnv:       config.DeployEnvLocal,
	})

	req := gatewaytf.GatewayRequestWithModel(scenario.FullKey, "test-model")
	w := httptest.NewRecorder()
	scenario.Gateway.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for test model in local, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestGatewayRejectsDevModelOutsideLocal(t *testing.T) {
	t.Parallel()
	envs := []string{config.DeployEnvStaging, config.DeployEnvProduction}
	for _, env := range envs {
		t.Run(env, func(t *testing.T) {
			t.Parallel()
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("proxy must not be reached for test model outside local")
				w.WriteHeader(http.StatusOK)
			}))
			defer backend.Close()

			scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{
				Budget:          1000,
				ProxyBackendURL: backend.URL,
				DeployEnv:       env,
			})

			req := gatewaytf.GatewayRequestWithModel(scenario.FullKey, "test-model")
			w := httptest.NewRecorder()
			scenario.Gateway.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Errorf("expected 403 for test model in %s, got %d; body: %s", env, w.Code, w.Body.String())
			}
		})
	}
}

func TestGatewaySingleStoreCall(t *testing.T) {
	t.Parallel()
	scenario := gatewaytf.BuildGatewayScenario(t, gatewaytf.GatewayScenarioOpts{Budget: 1000})
	gw, counter := gatewaytf.BuildGatewayWithCountingPrecheck(t, scenario)

	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, gatewaytf.GatewayRequest(scenario.FullKey))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if counter.Calls() != 1 {
		t.Fatalf("expected exactly 1 LoadPrecheckContext call, got %d", counter.Calls())
	}
}
