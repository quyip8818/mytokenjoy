package postgres

import "github.com/tokenjoy/backend/internal/config"

type logTables struct {
	logs             string
	reconcileCursors string
}

func logTablesFor(cfg config.Config) logTables {
	if cfg.LogSchemaIsolated {
		return logTables{
			logs:             "logs",
			reconcileCursors: "reconcile_cursors",
		}
	}
	return logTables{
		logs:             "newapi.logs",
		reconcileCursors: "backend.reconcile_cursors",
	}
}
