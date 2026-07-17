//go:build testhook

package middleware_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
)

type stubCompanyService struct {
	resolve func(ctx context.Context, companyID uuid.UUID) (domaincompany.Context, error)
}

func (s *stubCompanyService) ResolveCompanyContext(ctx context.Context, companyID uuid.UUID) (domaincompany.Context, error) {
	if s.resolve != nil {
		return s.resolve(ctx, companyID)
	}
	return domaincompany.Context{}, domain.NotFound("company not found")
}

type stubRevisionReader struct {
	revision int64
	err      error
}

func (s *stubRevisionReader) GetAuthzRevision(_ context.Context, _ uuid.UUID) (int64, error) {
	return s.revision, s.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func injectCompanyCtx(companyID uuid.UUID, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := domaincompany.WithContext(r.Context(), domaincompany.Context{CompanyID: companyID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
