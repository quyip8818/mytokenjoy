package dev_test

import (
	"net/http"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	testhttp "github.com/tokenjoy/backend/tests/testutil/http"
)

const devBearerPath = "/api/dev/platform-keys/" + contract.IDPlatformKey1 + "/bearer"

func TestDevPlatformKeyBearerAvailableInLocal(t *testing.T) {
	t.Parallel()
	router := testhttp.NewApp(t, func(cfg *config.Config) {
		testutil.WithDeployEnv(config.DeployEnvLocal)(cfg)
	}).Router

	rec := testhttp.ServeAuthz(t, router, http.MethodGet, devBearerPath, testhttp.AdminCookie(t), "", nil)
	if rec.Code != http.StatusOK && rec.Code != http.StatusConflict && rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected dev bearer route in local (200 or key sync error), got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestDevPlatformKeyBearerNotRegisteredOutsideLocal(t *testing.T) {
	t.Parallel()
	envs := []string{config.DeployEnvStaging, config.DeployEnvProduction}
	for _, env := range envs {
		t.Run(env, func(t *testing.T) {
			t.Parallel()
			router := testhttp.NewApp(t, func(cfg *config.Config) {
				if env == config.DeployEnvProduction {
					testutil.WithProductionContract()(cfg)
				} else {
					testutil.WithDeployEnv(env)(cfg)
				}
			}).Router

			rec := testhttp.ServeAuthz(t, router, http.MethodGet, devBearerPath, testhttp.AdminCookie(t), "", nil)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404 for %s, got %d body=%s", env, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestDevPlatformKeyBearerRequiresKeysAdmin(t *testing.T) {
	t.Parallel()
	router := testhttp.NewRouter(t)

	rec := testhttp.ServeAuthz(t, router, http.MethodGet, devBearerPath, "", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", rec.Code)
	}

	pureCookie := testutil.SessionCookie(t, contract.IDMemberPure)
	rec = testhttp.ServeAuthz(t, router, http.MethodGet, devBearerPath, pureCookie, "", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without KeysAdmin, got %d body=%s", rec.Code, rec.Body.String())
	}
}
