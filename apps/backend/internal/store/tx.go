package store

import "github.com/jackc/pgx/v5"

// Tx is a narrow transaction handle for transactional job enqueue.
// Domain code receives store.Store inside WithTx; infra uses type assertion to Tx.
type Tx interface {
	PgxTx() pgx.Tx
}
