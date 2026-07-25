package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store 是所有仓储方法的载体。各 domain service 声明自己的接口子集，
// 此 struct 通过实现对应方法来隐式满足。
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// newID 生成 UUID v7（时间有序）
func newID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
