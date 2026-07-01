package middleware

import (
	"net/http"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func CompanyReadOnlyMiddleware(gate *domaincompany.Gate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if gate.IsSuspended(r.Context()) && !isReadOnlyAllowed(r.Method) {
				httputil.WriteStatus(w, http.StatusForbidden, "Company suspended")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isReadOnlyAllowed(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
