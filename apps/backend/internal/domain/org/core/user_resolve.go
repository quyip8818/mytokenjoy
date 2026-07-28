package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

// ResolveOrCreateUser finds an existing user by phone or email, or creates a new one.
// Returns the user ID. When phone and email are both empty (e.g. dingtalk without
// mobile permission), creates a placeholder user identified by name only.
func ResolveOrCreateUser(ctx context.Context, st Store, phone, email, name string) (uuid.UUID, error) {
	if phone != "" {
		user, err := st.User().GetByPhone(ctx, phone)
		if err != nil {
			return uuid.Nil, err
		}
		if user != nil {
			return user.ID, nil
		}
	}
	if email != "" {
		user, err := st.User().GetByEmail(ctx, email)
		if err != nil {
			return uuid.Nil, err
		}
		if user != nil {
			return user.ID, nil
		}
	}

	// Create new user (phone/email may be empty for platforms that don't expose them).
	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	newUser := store.User{
		ID:        userID,
		Name:      name,
		Phone:     phone,
		Email:     email,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := st.User().Create(ctx, newUser); err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}
