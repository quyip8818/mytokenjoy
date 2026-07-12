//go:build testhook

package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// cloneDataExcludedTables are runtime ingest tables seeded empty like production;
// clone structure only so per-test writes never inherit template pollution.
var cloneDataExcludedTables = map[string]struct{}{
	"logs":              {},
	"ingest_jobs":       {},
	"reconcile_cursors": {},
}

func cloneIncludesData(table string, hasRows bool) bool {
	if !hasRows {
		return false
	}
	_, excluded := cloneDataExcludedTables[table]
	return !excluded
}

func LoadClonePlan(ctx context.Context, pool *pgxpool.Pool, schema string) (ClonePlan, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return ClonePlan{}, err
	}
	defer tx.Rollback(ctx)

	tables, err := listSchemaTablesInCloneOrder(ctx, tx, schema)
	if err != nil {
		return ClonePlan{}, err
	}

	plan := ClonePlan{
		TablesInOrder:  tables,
		ColumnLists:    make(map[string]string, len(tables)),
		TablesWithData: make(map[string]struct{}),
	}
	srcSQL := pgx.Identifier{schema}.Sanitize()
	for _, table := range tables {
		columns, err := writableTableColumns(ctx, tx, schema, table)
		if err != nil {
			return ClonePlan{}, fmt.Errorf("columns for %s: %w", table, err)
		}
		if len(columns) > 0 {
			plan.ColumnLists[table] = quoteColumnList(columns)
		}
		tableSQL := pgx.Identifier{table}.Sanitize()
		var hasRows bool
		if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s.%s LIMIT 1)", srcSQL, tableSQL)).Scan(&hasRows); err != nil {
			return ClonePlan{}, fmt.Errorf("count %s: %w", table, err)
		}
		if cloneIncludesData(table, hasRows) {
			plan.TablesWithData[table] = struct{}{}
		}
	}

	fkRows, err := tx.Query(ctx, `
		SELECT child.relname, con.conname, pg_get_constraintdef(con.oid, true)
		FROM pg_constraint con
		JOIN pg_class child ON child.oid = con.conrelid
		JOIN pg_namespace child_ns ON child_ns.oid = child.relnamespace
		WHERE con.contype = 'f'
		  AND child_ns.nspname = $1
		  AND NOT EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = child.oid)
	`, schema)
	if err != nil {
		return ClonePlan{}, fmt.Errorf("list template foreign keys: %w", err)
	}
	defer fkRows.Close()
	for fkRows.Next() {
		var item ForeignKeyClone
		if err := fkRows.Scan(&item.ChildTable, &item.ConstraintName, &item.Definition); err != nil {
			return ClonePlan{}, err
		}
		plan.ForeignKeys = append(plan.ForeignKeys, item)
	}
	if err := fkRows.Err(); err != nil {
		return ClonePlan{}, err
	}

	serialRows, err := tx.Query(ctx, `
		SELECT c.table_name, c.column_name
		FROM information_schema.columns c
		JOIN pg_namespace n ON n.nspname = c.table_schema
		JOIN pg_class cls ON cls.relnamespace = n.oid AND cls.relname = c.table_name
		WHERE c.table_schema = $1
		  AND cls.relkind IN ('r', 'p')
		  AND NOT EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = cls.oid)
		  AND pg_get_serial_sequence(format('%I.%I', c.table_schema, c.table_name), c.column_name) IS NOT NULL
	`, schema)
	if err != nil {
		return ClonePlan{}, fmt.Errorf("list template serial columns: %w", err)
	}
	defer serialRows.Close()
	for serialRows.Next() {
		var target SerialTarget
		if err := serialRows.Scan(&target.Table, &target.Column); err != nil {
			return ClonePlan{}, err
		}
		plan.SerialColumns = append(plan.SerialColumns, target)
	}
	if err := serialRows.Err(); err != nil {
		return ClonePlan{}, err
	}
	return plan, nil
}

func CloneSchema(ctx context.Context, pool *pgxpool.Pool, src, dst string, plan ClonePlan) error {
	srcSQL := pgx.Identifier{src}.Sanitize()
	dstSQL := pgx.Identifier{dst}.Sanitize()

	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+dstSQL); err != nil {
		return fmt.Errorf("create schema %s: %w", dst, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin clone tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL session_replication_role = replica"); err != nil {
		return fmt.Errorf("disable fk checks: %w", err)
	}

	// Clone parent/plain tables only. Partition children are skipped: LIKE on a
	// partitioned parent yields a plain table that still accepts any occurred_at,
	// which matches prior test semantics and avoids 36 extra DDL round-trips.
	var ddl strings.Builder
	for _, table := range plan.TablesInOrder {
		tableSQL := pgx.Identifier{table}.Sanitize()
		fmt.Fprintf(
			&ddl,
			"CREATE TABLE %s.%s (LIKE %s.%s INCLUDING ALL);\n",
			dstSQL, tableSQL, srcSQL, tableSQL,
		)
	}
	if ddl.Len() > 0 {
		if _, err := tx.Exec(ctx, ddl.String()); err != nil {
			return fmt.Errorf("clone table structures: %w", err)
		}
	}

	var inserts strings.Builder
	for table := range plan.TablesWithData {
		tableSQL := pgx.Identifier{table}.Sanitize()
		columnList := plan.ColumnLists[table]
		fmt.Fprintf(
			&inserts,
			"INSERT INTO %s.%s (%s) SELECT %s FROM %s.%s;\n",
			dstSQL, tableSQL, columnList, columnList, srcSQL, tableSQL,
		)
	}
	if inserts.Len() > 0 {
		if _, err := tx.Exec(ctx, inserts.String()); err != nil {
			return fmt.Errorf("clone table data: %w", err)
		}
	}

	var fkDDL strings.Builder
	for _, item := range plan.ForeignKeys {
		childSQL := pgx.Identifier{item.ChildTable}.Sanitize()
		nameSQL := pgx.Identifier{item.ConstraintName}.Sanitize()
		definition := remapForeignKeyDefinition(item.Definition, src, dst)
		fmt.Fprintf(
			&fkDDL,
			"ALTER TABLE %s.%s ADD CONSTRAINT %s %s;\n",
			dstSQL, childSQL, nameSQL, definition,
		)
	}
	if fkDDL.Len() > 0 {
		if _, err := tx.Exec(ctx, fkDDL.String()); err != nil {
			return fmt.Errorf("clone foreign keys: %w", err)
		}
	}

	if err := syncSchemaSequences(ctx, tx, dst, plan.SerialColumns); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "SET LOCAL session_replication_role = DEFAULT"); err != nil {
		return fmt.Errorf("restore fk checks: %w", err)
	}
	return tx.Commit(ctx)
}

func syncSchemaSequences(ctx context.Context, tx pgx.Tx, dstSchema string, targets []SerialTarget) error {
	// Cloned test schemas may lack serial sequences (e.g. river_job_id_seq); create and sync them here.
	if len(targets) == 0 {
		return nil
	}
	var stmts strings.Builder
	for _, target := range targets {
		columnSQL := pgx.Identifier{target.Column}.Sanitize()
		seqName := target.Table + "_" + target.Column + "_seq"
		fmt.Fprintf(
			&stmts,
			`DO $clone$ BEGIN
IF pg_get_serial_sequence('%[1]s.%[2]s', '%[3]s') IS NULL THEN
  CREATE SEQUENCE %[1]s.%[4]s OWNED BY %[1]s.%[2]s.%[3]s;
  ALTER TABLE %[1]s.%[2]s ALTER COLUMN %[3]s SET DEFAULT nextval('%[1]s.%[4]s'::regclass);
END IF;
PERFORM setval(pg_get_serial_sequence('%[1]s.%[2]s', '%[3]s'), COALESCE((SELECT MAX(%[5]s) FROM %[1]s.%[2]s), 1), true);
END $clone$;
`,
			dstSchema, target.Table, target.Column, seqName, columnSQL,
		)
	}
	if _, err := tx.Exec(ctx, stmts.String()); err != nil {
		return fmt.Errorf("sync schema sequences: %w", err)
	}
	return nil
}

func writableTableColumns(ctx context.Context, tx pgx.Tx, schema, table string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name = $2
		  AND is_generated = 'NEVER'
		ORDER BY ordinal_position
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, rows.Err()
}

func listSchemaTablesInCloneOrder(ctx context.Context, tx pgx.Tx, schema string) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1
		  AND c.relkind IN ('r', 'p')
		  AND NOT EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = c.oid)
		ORDER BY c.relname
	`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableSet := make(map[string]struct{})
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tableSet[name] = struct{}{}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	depRows, err := tx.Query(ctx, `
		SELECT child.relname, parent.relname
		FROM pg_constraint con
		JOIN pg_class child ON child.oid = con.conrelid
		JOIN pg_namespace child_ns ON child_ns.oid = child.relnamespace
		JOIN pg_class parent ON parent.oid = con.confrelid
		JOIN pg_namespace parent_ns ON parent_ns.oid = parent.relnamespace
		WHERE con.contype = 'f'
		  AND child_ns.nspname = $1
		  AND parent_ns.nspname = $1
		  AND NOT EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = child.oid)
		  AND NOT EXISTS (SELECT 1 FROM pg_inherits i WHERE i.inhrelid = parent.oid)
	`, schema)
	if err != nil {
		return nil, err
	}
	defer depRows.Close()

	indegree := make(map[string]int, len(tables))
	dependents := make(map[string][]string, len(tables))
	for name := range tableSet {
		indegree[name] = 0
	}
	for depRows.Next() {
		var child, parent string
		if err := depRows.Scan(&child, &parent); err != nil {
			return nil, err
		}
		if child == parent {
			continue
		}
		if _, ok := tableSet[child]; !ok {
			continue
		}
		if _, ok := tableSet[parent]; !ok {
			continue
		}
		dependents[parent] = append(dependents[parent], child)
		indegree[child]++
	}
	if err := depRows.Err(); err != nil {
		return nil, err
	}

	queue := make([]string, 0, len(tables))
	for _, name := range tables {
		if indegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	ordered := make([]string, 0, len(tables))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		ordered = append(ordered, current)
		for _, child := range dependents[current] {
			indegree[child]--
			if indegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}
	if len(ordered) != len(tables) {
		return nil, fmt.Errorf("schema %s has cyclic foreign keys", schema)
	}
	return ordered, nil
}

func remapForeignKeyDefinition(def, srcSchema, dstSchema string) string {
	quotedDst := pgx.Identifier{dstSchema}.Sanitize()
	def = strings.ReplaceAll(def, srcSchema+".", quotedDst+".")
	quotedSrc := pgx.Identifier{srcSchema}.Sanitize()
	def = strings.ReplaceAll(def, quotedSrc+".", quotedDst+".")
	return rewriteUnqualifiedForeignKeyRef(def, quotedDst)
}

func rewriteUnqualifiedForeignKeyRef(definition, dstSQL string) string {
	const prefix = "REFERENCES "
	idx := strings.Index(definition, prefix)
	if idx < 0 {
		return definition
	}
	rest := definition[idx+len(prefix):]
	if strings.HasPrefix(rest, dstSQL+".") {
		return definition
	}
	spaceIdx := strings.IndexAny(rest, " (")
	if spaceIdx <= 0 {
		return definition
	}
	tableName := rest[:spaceIdx]
	tableSQL := pgx.Identifier{tableName}.Sanitize()
	return definition[:idx] + prefix + dstSQL + "." + tableSQL + rest[spaceIdx:]
}

func quoteColumnList(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = pgx.Identifier{col}.Sanitize()
	}
	return strings.Join(quoted, ", ")
}
