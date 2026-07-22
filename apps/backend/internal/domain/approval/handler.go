package approval

import (
	"context"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// ApproveResult is the output of OnApprovedTx, passed through to PostApprove/Compensate.
// Each Handler defines its own concrete struct; Engine only passes it through.
// Handlers with no post-transaction side effects return nil.
type ApproveResult interface{}

// Handler defines the business logic for a specific approval type.
// Add a new type by implementing this interface and registering with the Engine.
type Handler interface {
	// Type returns the approval type this handler processes.
	Type() types.ApprovalType

	// Validate checks input legality when creating an approval.
	Validate(ctx context.Context, input CreateInput) error

	// PreApprove runs pre-checks before approval (outside transaction, fail-fast).
	PreApprove(ctx context.Context, req types.ApprovalRequest) error

	// OnApprovedTx executes in-transaction side effects (pure DB ops).
	// Returns ApproveResult to pass to PostApprove/Compensate.
	// Handlers that deduct balances MUST acquire row locks (SELECT ... FOR UPDATE) here.
	OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (ApproveResult, error)

	// PostApprove executes post-transaction side effects (external IO).
	// result is from OnApprovedTx. Return nil if no external IO needed.
	// On error, Engine calls Compensate then marks the approval as failed.
	PostApprove(ctx context.Context, req types.ApprovalRequest, result ApproveResult) error

	// Compensate rolls back business data from OnApprovedTx. MUST be idempotent.
	// Called in two scenarios:
	//   1. Approve flow: PostApprove failed, result is non-nil
	//   2. Retry flow: clearing residual data, result is nil (infer from DB state)
	Compensate(ctx context.Context, req types.ApprovalRequest, result ApproveResult) error

	// OnRejected runs in-transaction side effects on rejection (most types: no-op).
	OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error

	// PreCheck returns a JSON payload for the approver to preview conditions.
	PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error)
}
