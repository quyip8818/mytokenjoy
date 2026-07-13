//go:build testhook

package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func sessionConfig() config.Config {
	return testutil.WithSessionConfig(config.Config{})
}

func TestMiddlewareBehaviors(t *testing.T) {
	t.Parallel()

	t.Run("M1 company resolve missing tenant", func(t *testing.T) {
		t.Parallel()
		stub := &stubCompanyService{
			resolve: func(_ context.Context, companyID int64) (domaincompany.Context, error) {
				if companyID != 0 {
					t.Fatalf("expected companyID 0, got %d", companyID)
				}
				return domaincompany.Context{}, domain.NotFound("company not found")
			},
		}
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next should not run")
		})
		cfg := sessionConfig()
		cfg.LocalCompanyID = 0
		handler := httpmiddleware.CompanyResolve(cfg, stub, testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("M1b company resolve infra error is 500", func(t *testing.T) {
		t.Parallel()
		stub := &stubCompanyService{
			resolve: func(context.Context, int64) (domaincompany.Context, error) {
				return domaincompany.Context{}, fmt.Errorf("db unavailable")
			},
		}
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next should not run")
		})
		cfg := sessionConfig()
		cfg.LocalCompanyID = contract.DefaultCompanyID
		handler := httpmiddleware.CompanyResolve(cfg, stub, testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("M2 platform route skips company resolve", func(t *testing.T) {
		t.Parallel()
		stub := &stubCompanyService{
			resolve: func(context.Context, int64) (domaincompany.Context, error) {
				t.Fatal("company resolve should be skipped for platform routes")
				return domaincompany.Context{}, nil
			},
		}
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.CompanyResolve(sessionConfig(), stub, testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodGet, "/api/platform/companies", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Fatal("expected next handler to run")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("M3 platform auth unauthorized", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next should not run")
		})
		handler := httpmiddleware.PlatformAuth(sessionConfig(), testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodGet, "/api/platform/companies", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("M4 platform login bypass", func(t *testing.T) {
		t.Parallel()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		})
		handler := httpmiddleware.PlatformAuth(sessionConfig(), testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodPost, "/api/platform/auth/login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Fatal("expected login path to bypass platform auth")
		}
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("M5 company gate blocks writes when suspended", func(t *testing.T) {
		t.Parallel()
		gate := domaincompany.NewGate(config.Config{})
		injectSuspended := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := domaincompany.WithContext(r.Context(), domaincompany.Context{
					CompanyID: 1,
					Status:    "suspended",
				})
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}
		okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := injectSuspended(httpmiddleware.CompanyReadOnlyMiddleware(gate)(okHandler))

		postRec := httptest.NewRecorder()
		handler.ServeHTTP(postRec, httptest.NewRequest(http.MethodPost, "/api/budget/tree", nil))
		if postRec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for POST, got %d body=%s", postRec.Code, postRec.Body.String())
		}

		getRec := httptest.NewRecorder()
		handler.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/api/budget/tree", nil))
		if getRec.Code != http.StatusOK {
			t.Fatalf("expected 200 for GET, got %d body=%s", getRec.Code, getRec.Body.String())
		}
	})

	t.Run("M6 authz revision header from session", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.AuthzRevisionHeader(&stubRevisionReader{})(next)

		req := httptest.NewRequest(http.MethodGet, "/api/org/departments/tree", nil)
		req = req.WithContext(httpx.WithSessionContext(req.Context(), types.SessionContext{
			AuthzRevision: 42,
		}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Header().Get(httpx.HeaderAuthzRevision); got != "42" {
			t.Fatalf("expected revision header 42, got %q", got)
		}
	})

	t.Run("M6 authz revision header from company repo", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.AuthzRevisionHeader(&stubRevisionReader{revision: 77})(next)

		req := httptest.NewRequest(http.MethodGet, "/api/org/departments/tree", nil)
		req = req.WithContext(domaincompany.WithContext(req.Context(), domaincompany.Context{
			CompanyID: contract.DefaultCompanyID,
		}))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Header().Get(httpx.HeaderAuthzRevision); got != "77" {
			t.Fatalf("expected revision header 77, got %q", got)
		}
	})

	t.Run("M7 require session rejects tampered jwt", func(t *testing.T) {
		t.Parallel()
		cfg, st := testutil.NewTestStore(t)
		protected := httpdeps.Protected{
			Cfg:          cfg,
			AuthzSvc:     authz.NewService(cfg, st),
			SessionToken: testutil.SessionIssuer(t),
		}
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.RequireSession(protected)(next)

		token := testutil.IssueSessionJWT(t, contract.DefaultCompanyID, contract.IDMemberAdmin)
		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		req.Header.Set("Cookie", httpx.SessionCookie+"="+token+"tampered")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
		}
	})
}
