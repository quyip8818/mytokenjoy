package ingest

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

// Mount registers the ingest handler on the given router under /internal.
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Config, d.IngestEnqueuer, d.IngestMetrics, d.Logger)
	r.Route("/internal", h.RegisterRoutes)
}
