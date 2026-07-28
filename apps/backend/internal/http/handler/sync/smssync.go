package synchandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// Mount registers sync trigger routes on /sync.
func Mount(r chi.Router, d httpdeps.Deps) {
	r.Route("/sync", func(sync chi.Router) {
		sync.Post("/sms/trigger", handleSMSSyncTrigger(d.IngestEnqueuer))
	})
}

// handleSMSSyncTrigger enqueues an SMS sync job for immediate execution.
func handleSMSSyncTrigger(enqueuer jobs.Enqueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := enqueuer.Insert(r.Context(), jobs.SMSSyncArgs{}, nil); err != nil {
			response.Error(w, http.StatusInternalServerError, "failed to enqueue sms sync: "+err.Error())
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "enqueued"})
	}
}
