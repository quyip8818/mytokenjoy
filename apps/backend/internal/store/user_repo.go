package store

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Phone        string
	Email        string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	UpdatePhone(ctx context.Context, id string, phone string) error
	UpdateEmail(ctx context.Context, id string, email string) error
	UpdateStatus(ctx context.Context, id string, status string) error
}
