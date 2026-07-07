package postgres

import "github.com/tokenjoy/backend/internal/config"

type logTables struct {
	logs             string
	ingestFailures   string
	reconcileCursors string
}

func logTablesFor(cfg config.Config) logTables {
	if cfg.LogSchemaIsolated {
		return logTables{
			logs:             "logs",
			ingestFailures:   "ingest_failures",
			reconcileCursors: "reconcile_cursors",
		}
	}
	return logTables{
		logs:             "newapi.logs",
		ingestFailures:   "backend.ingest_failures",
		reconcileCursors: "backend.reconcile_cursors",
	}
}
