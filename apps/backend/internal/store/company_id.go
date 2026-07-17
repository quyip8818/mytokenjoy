package store

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

// CompanyID extracts the tenant company ID from context.
// Logs a warning if missing (zero) — this indicates a programming error where
// a tenant-scoped query is called without proper company context setup.
func CompanyID(ctx context.Context) uuid.UUID {
	id := ctxcompany.ID(ctx)
	if id == uuid.Nil {
		_, file, line, _ := runtime.Caller(1)
		slog.Default().Warn("store.CompanyID called with zero company context",
			"caller", file, "line", line)
	}
	return id
}

// CompanyIDOrZero returns the company ID without any warning.
// Use for operations that legitimately work without tenant context
// (e.g., cross-tenant lookups like FindMemberCompanyID).
func CompanyIDOrZero(ctx context.Context) uuid.UUID {
	return ctxcompany.ID(ctx)
}
