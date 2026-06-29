package postgres

import (
	"context"
)

type pgBudgetRepo struct {
	ctx context.Context
	db  dbQuerier
}
