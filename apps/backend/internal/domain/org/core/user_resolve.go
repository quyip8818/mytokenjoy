package core

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/store"
)

// ResolveOrCreateUser finds an existing user by phone or email, or creates a new one.
// Returns the user ID.
func ResolveOrCreateUser(ctx context.Context, st Store, phone, email string) (string, error) {
	if phone != "" {
		user, err := st.User().GetByPhone(ctx, phone)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.ID, nil
		}
	}
	if email != "" {
		user, err := st.User().GetByEmail(ctx, email)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.ID, nil
		}
	}

	// Create new user.
	now := time.Now().UTC()
	userID := fmt.Sprintf("u-%d", now.UnixNano())
	newUser := store.User{
		ID:        userID,
		Phone:     phone,
		Email:     email,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := st.User().Create(ctx, newUser); err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return userID, nil
}
