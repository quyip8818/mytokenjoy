//go:build testhook

package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// readDBVersion reads the template version from COMMENT ON DATABASE.
func readDBVersion(ctx context.Context, conn *pgx.Conn, dbName string) (int, bool) {
	var comment *string
	err := conn.QueryRow(ctx, `
		SELECT pg_catalog.shobj_description(oid, 'pg_database')
		FROM pg_database
		WHERE datname = $1
	`, dbName).Scan(&comment)
	if err != nil || comment == nil {
		return 0, false
	}
	var version int
	if _, err := fmt.Sscanf(*comment, "version:%d", &version); err != nil {
		return 0, false
	}
	return version, true
}

// markDBVersion stamps the database with a version comment.
func markDBVersion(ctx context.Context, conn *pgx.Conn, dbName string, version int) error {
	_, err := conn.Exec(ctx, fmt.Sprintf(
		"COMMENT ON DATABASE %s IS 'version:%d'",
		pgx.Identifier{dbName}.Sanitize(), version,
	))
	return err
}
