//go:build testhook

package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/infra/ratelimit"
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
			resolve: func(_ context.Context, companyID uuid.UUID) (domaincompany.Context, error) {
				t.Fatal("resolve should not be called when companyID is nil")
				return domaincompany.Context{}, nil
			},
		}
		var nextCalled bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})
		cfg := sessionConfig()
		cfg.LocalCompanyID = uuid.Nil
		handler := httpmiddleware.CompanyResolve(cfg, stub, testutil.SessionIssuer(t))(next)

		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !nextCalled {
			t.Fatal("expected next to be called when no tenant resolved")
		}
	})

	t.Run("M1b company resolve infra error is 500", func(t *testing.T) {
		t.Parallel()
		stub := &stubCompanyService{
			resolve: func(context.Context, uuid.UUID) (domaincompany.Context, error) {
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
			resolve: func(context.Context, uuid.UUID) (domaincompany.Context, error) {
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
					CompanyID: uuid.MustParse("00000000-0000-7000-0000-000000000001"),
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

	t.Run("M8 security headers present", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.SecurityHeaders(true)(next)

		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("X-Content-Type-Options: got %q want nosniff", got)
		}
		if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Fatalf("X-Frame-Options: got %q want DENY", got)
		}
		if got := rec.Header().Get("Strict-Transport-Security"); got == "" {
			t.Fatal("expected HSTS header when secureCookie=true")
		}
	})

	t.Run("M8b security headers no HSTS when not secure", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.SecurityHeaders(false)(next)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
			t.Fatalf("expected no HSTS when secureCookie=false, got %q", got)
		}
	})

	t.Run("M9 request timeout sets context deadline", func(t *testing.T) {
		t.Parallel()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deadline, ok := r.Context().Deadline()
			if !ok {
				t.Fatal("expected context deadline to be set")
			}
			if deadline.IsZero() {
				t.Fatal("deadline should not be zero")
			}
			w.WriteHeader(http.StatusOK)
		})
		handler := httpmiddleware.RequestTimeout(5)(next)

		req := httptest.NewRequest(http.MethodGet, "/api/session", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("M10 rate limit tenant allows when under limit", func(t *testing.T) {
		t.Parallel()
		limiter := ratelimit.NewMemoryLimiter()
		defer limiter.Close()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		handler := injectCompanyCtx(contract.DefaultCompanyID, httpmiddleware.RateLimitTenant(limiter, 100, 200, false, testLogger())(next))

		req := httptest.NewRequest(http.MethodGet, "/api/budget/tree", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Fatal("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("X-RateLimit-Limit") == "" {
			t.Fatal("expected X-RateLimit-Limit header")
		}
	})

	t.Run("M10b rate limit tenant rejects when exhausted", func(t *testing.T) {
		t.Parallel()
		limiter := ratelimit.NewMemoryLimiter()
		defer limiter.Close()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := injectCompanyCtx(contract.DefaultCompanyID, httpmiddleware.RateLimitTenant(limiter, 1, 1, false, testLogger())(next))

		// First request — allowed (uses the 1 token).
		req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("first request: expected 200, got %d", rec.Code)
		}

		// Second request — rejected.
		req2 := httptest.NewRequest(http.MethodGet, "/api/x", nil)
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusTooManyRequests {
			t.Fatalf("second request: expected 429, got %d body=%s", rec2.Code, rec2.Body.String())
		}
	})

	t.Run("M11 rate limit login paths blocks after max", func(t *testing.T) {
		t.Parallel()
		memLimiter := ratelimit.NewMemoryLimiter()
		defer memLimiter.Close()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		paths := []string{"/api/auth/login"}
		handler := httpmiddleware.RateLimitLoginPaths(paths, nil, memLimiter, 2, 60, false, testLogger())(next)

		// First 2 requests pass.
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
			}
		}
		// 3rd request blocked.
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("3rd request: expected 429, got %d", rec.Code)
		}
	})

	t.Run("M11b rate limit login paths ignores non-login", func(t *testing.T) {
		t.Parallel()
		memLimiter := ratelimit.NewMemoryLimiter()
		defer memLimiter.Close()
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})
		paths := []string{"/api/auth/login"}
		handler := httpmiddleware.RateLimitLoginPaths(paths, nil, memLimiter, 1, 60, false, testLogger())(next)

		// Non-login path — should pass without rate limiting.
		req := httptest.NewRequest(http.MethodPost, "/api/budget/tree", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !called {
			t.Fatal("expected next handler for non-login path")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})
}
