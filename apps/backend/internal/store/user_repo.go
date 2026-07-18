package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Phone        string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdatePhone(ctx context.Context, id uuid.UUID, phone string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	HasAnyMember(ctx context.Context, userID uuid.UUID) (bool, error)
}
