package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"sms/backend/internal/domain/types"
)

type Store interface {
	ListUsers(ctx context.Context) ([]types.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*types.User, error)
	CreateUser(ctx context.Context, u *types.User) error
	UpdateUser(ctx context.Context, u *types.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type CreateInput struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	RealName string  `json:"realName"`
	Email    *string `json:"email"`
	Role     string  `json:"role"`
	Status   *int    `json:"status"`
}

type UpdateInput struct {
	RealName string  `json:"realName"`
	Email    *string `json:"email"`
	Role     string  `json:"role"`
	Password string  `json:"password"`
	Status   *int    `json:"status"`
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]types.User, error) {
	return s.store.ListUsers(ctx)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*types.User, error) {
	if input.Username == "" || input.Password == "" || input.RealName == "" {
		return nil, fmt.Errorf("%w: 用户名、密码和真实姓名不能为空", types.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, err
	}
	u := &types.User{
		ID:           uuid.Must(uuid.NewV7()),
		Username:     input.Username,
		PasswordHash: string(hash),
		RealName:     input.RealName,
		Email:        input.Email,
		Role:         input.Role,
		Status:       1,
	}
	if input.Status != nil {
		u.Status = *input.Status
	}
	if u.Role == "" {
		u.Role = "viewer"
	}
	if err := s.store.CreateUser(ctx, u); err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*types.User, error) {
	existing, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	existing.RealName = input.RealName
	existing.Email = input.Email
	if input.Role != "" {
		existing.Role = input.Role
	}
	if input.Status != nil {
		existing.Status = *input.Status
	}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
		if err != nil {
			return nil, err
		}
		existing.PasswordHash = string(hash)
	}
	if err := s.store.UpdateUser(ctx, existing); err != nil {
		return nil, err
	}
	existing.PasswordHash = ""
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.store.DeleteUser(ctx, id)
}
