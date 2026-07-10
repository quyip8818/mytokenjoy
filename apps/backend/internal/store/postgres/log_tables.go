package postgres

import "github.com/tokenjoy/backend/internal/config"

type logTables struct {
	logs             string
	ingestJobs       string
	reconcileCursors string
}

func logTablesFor(cfg config.Config) logTables {
	if cfg.LogSchemaIsolated {
		return logTables{
			logs:             "logs",
			ingestJobs:       "ingest_jobs",
			reconcileCursors: "reconcile_cursors",
		}
	}
	return logTables{
		logs:             "newapi.logs",
		ingestJobs:       "backend.ingest_jobs",
		reconcileCursors: "backend.reconcile_cursors",
	}
}
