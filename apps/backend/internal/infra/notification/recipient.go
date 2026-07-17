package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

// RecipientInfo holds resolved contact details for a member.
type RecipientInfo struct {
	MemberID uuid.UUID
	Email    string
	Phone    string
	Name     string
}

// RecipientResolver resolves a memberID into contact details (email, phone).
type RecipientResolver struct {
	store store.Store
}

func NewRecipientResolver(st store.Store) *RecipientResolver {
	return &RecipientResolver{store: st}
}

// Resolve looks up a member by ID and returns their contact information.
// Returns a zero-value RecipientInfo with only MemberID set if lookup fails.
func (r *RecipientResolver) Resolve(ctx context.Context, memberID uuid.UUID) RecipientInfo {
	member, err := r.store.Org().MemberByID(ctx, memberID)
	if err != nil || member == nil {
		return RecipientInfo{MemberID: memberID}
	}
	return RecipientInfo{
		MemberID: member.ID,
		Email:    member.Email,
		Phone:    member.Phone,
		Name:     member.Name,
	}
}
