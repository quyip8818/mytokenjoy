package store

import (
	"context"
	"time"
)

type PlatformOperator struct {
	ID           string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlatformRepository interface {
	GetOperatorByEmail(ctx context.Context, email string) (*PlatformOperator, error)
	GetOperatorByID(ctx context.Context, id string) (*PlatformOperator, error)
	CreateOperator(ctx context.Context, op PlatformOperator) error
	CountOperators(ctx context.Context) (int, error)
}
