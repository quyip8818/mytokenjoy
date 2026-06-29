package postgres

import (
	"context"
)

type pgKeysRepo struct {
	ctx context.Context
	db  dbQuerier
}
