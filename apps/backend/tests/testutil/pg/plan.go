//go:build testhook

package pg

type ClonePlan struct {
	TablesInOrder  []string
	ColumnLists    map[string]string
	TablesWithData map[string]struct{}
	ForeignKeys    []ForeignKeyClone
	SerialColumns  []SerialTarget
}

type ForeignKeyClone struct {
	ChildTable     string
	ConstraintName string
	Definition     string
}

type SerialTarget struct {
	Table  string
	Column string
}
