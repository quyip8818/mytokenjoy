package jobs

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
)

// IngestArgs represents a single consume-log ingest job.
type IngestArgs struct {
	LogID  int64  `json:"log_id" river:"unique"`
	Source string `json:"source"`
}

func (IngestArgs) Kind() string { return KindIngest }

func (IngestArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueCritical,
		MaxAttempts: 20,
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

// IngestReconcileArgs triggers a reconcile pass that scans consume logs
// after the stored cursor and enqueues any missing ingest jobs.
type IngestReconcileArgs struct{}

func (IngestReconcileArgs) Kind() string { return KindIngestReconcile }

func (IngestReconcileArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       config.RiverQueueDefault,
		MaxAttempts: 3,
	}
}
