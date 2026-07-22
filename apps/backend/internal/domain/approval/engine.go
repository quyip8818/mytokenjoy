package approval

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// TxRunner executes fn inside a database transaction.
type TxRunner func(ctx context.Context, fn func(store.Store) error) error

// Engine orchestrates the approval state machine.
type Engine struct {
	repo     Repository
	txRunner TxRunner
	handlers map[types.ApprovalType]Handler
	logger   *slog.Logger
}

// NewEngine creates an Engine with the given handlers registered.
func NewEngine(repo Repository, txRunner TxRunner, logger *slog.Logger, handlers ...Handler) *Engine {
	m := make(map[types.ApprovalType]Handler, len(handlers))
	for _, h := range handlers {
		m[h.Type()] = h
	}
	return &Engine{repo: repo, txRunner: txRunner, logger: logger, handlers: m}
}

// --- Create ---

func (e *Engine) Create(ctx context.Context, input CreateInput) (types.ApprovalRequest, error) {
	handler, ok := e.handlers[input.Type]
	if !ok {
		return types.ApprovalRequest{}, domain.Validation("unknown approval type")
	}
	if err := handler.Validate(ctx, input); err != nil {
		return types.ApprovalRequest{}, err
	}
	req := types.ApprovalRequest{
		ID:             uuid.Must(uuid.NewV7()),
		CompanyID:      store.CompanyID(ctx),
		Type:           input.Type,
		Status:         types.ApprovalPending,
		ApplicantID:    input.ApplicantID,
		ApplicantName:  input.ApplicantName,
		DepartmentID:   input.DepartmentID,
		DepartmentName: input.DepartmentName,
		Metadata:       input.Metadata,
		CreatedAt:      time.Now().UTC(),
	}
	if err := e.repo.Create(ctx, req); err != nil {
		return types.ApprovalRequest{}, err
	}
	return req, nil
}

// --- Approve ---

func (e *Engine) Approve(ctx context.Context, id uuid.UUID, approver ApproverInfo) error {
	req, err := e.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != types.ApprovalPending {
		return domain.Validation("approval already resolved")
	}
	handler, ok := e.handlers[req.Type]
	if !ok {
		return domain.Validation("no handler registered for type: " + string(req.Type))
	}

	// 1. Pre-check (outside tx, fail-fast)
	if err := handler.PreApprove(ctx, req); err != nil {
		return err
	}

	// 2. Transaction: update status + business side effects
	now := time.Now().UTC()
	req.Status = types.ApprovalApproved
	req.ApproverID = &approver.ID
	req.ApproverName = &approver.Name
	req.ResolvedAt = &now

	var result ApproveResult
	if err := e.txRunner(ctx, func(tx store.Store) error {
		if err := tx.Approval().Update(ctx, req); err != nil {
			return err
		}
		var txErr error
		result, txErr = handler.OnApprovedTx(ctx, req, tx)
		return txErr
	}); err != nil {
		return err
	}

	// 3. Post-transaction: external IO
	if err := handler.PostApprove(ctx, req, result); err != nil {
		e.logger.Error("PostApprove failed, compensating",
			"approval_id", id, "type", req.Type, "error", err)

		// 4. Compensate business data (best-effort, idempotent)
		if compErr := handler.Compensate(ctx, req, result); compErr != nil {
			e.logger.Error("Compensate failed",
				"approval_id", id, "error", compErr)
		}

		// 5. Mark failed regardless of compensate outcome
		e.markFailed(ctx, req)
		return err
	}

	return nil
}

// --- Reject ---

func (e *Engine) Reject(ctx context.Context, id uuid.UUID, approver ApproverInfo, reason string) error {
	req, err := e.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != types.ApprovalPending {
		return domain.Validation("approval already resolved")
	}
	handler, ok := e.handlers[req.Type]
	if !ok {
		return domain.Validation("no handler registered for type: " + string(req.Type))
	}

	now := time.Now().UTC()
	req.Status = types.ApprovalRejected
	req.ApproverID = &approver.ID
	req.ApproverName = &approver.Name
	req.RejectReason = &reason
	req.ResolvedAt = &now

	return e.txRunner(ctx, func(tx store.Store) error {
		if err := tx.Approval().Update(ctx, req); err != nil {
			return err
		}
		return handler.OnRejected(ctx, req, tx)
	})
}

// --- Cancel ---

func (e *Engine) Cancel(ctx context.Context, id uuid.UUID, applicantID uuid.UUID) error {
	req, err := e.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != types.ApprovalPending {
		return domain.Validation("approval already resolved")
	}
	if req.ApplicantID != applicantID {
		return domain.Forbidden("only applicant can cancel")
	}
	now := time.Now().UTC()
	req.Status = types.ApprovalCancelled
	req.ResolvedAt = &now
	return e.repo.Update(ctx, req)
}

// --- Retry ---

func (e *Engine) Retry(ctx context.Context, id uuid.UUID) error {
	req, err := e.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if req.Status != types.ApprovalFailed {
		return domain.Validation("only failed approvals can be retried")
	}
	handler, ok := e.handlers[req.Type]
	if !ok {
		return domain.Validation("no handler registered for type: " + string(req.Type))
	}

	// 1. Compensate residual data (idempotent, result=nil means infer from DB)
	if err := handler.Compensate(ctx, req, nil); err != nil {
		e.logger.Error("Retry compensate failed", "approval_id", id, "error", err)
		return err
	}

	// 2. Re-run full chain
	if err := handler.PreApprove(ctx, req); err != nil {
		return err
	}

	var result ApproveResult
	req.Status = types.ApprovalApproved
	if err := e.txRunner(ctx, func(tx store.Store) error {
		if err := tx.Approval().Update(ctx, req); err != nil {
			return err
		}
		var txErr error
		result, txErr = handler.OnApprovedTx(ctx, req, tx)
		return txErr
	}); err != nil {
		return err
	}

	if err := handler.PostApprove(ctx, req, result); err != nil {
		e.logger.Error("Retry PostApprove failed", "approval_id", id, "error", err)
		if compErr := handler.Compensate(ctx, req, result); compErr != nil {
			e.logger.Error("Retry Compensate failed", "approval_id", id, "error", compErr)
		}
		e.markFailed(ctx, req)
		return err
	}

	return nil
}

// --- Query ---

func (e *Engine) List(ctx context.Context, filter store.ApprovalListFilter) ([]types.ApprovalRequest, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return e.repo.List(ctx, filter)
}

func (e *Engine) Get(ctx context.Context, id uuid.UUID) (types.ApprovalRequest, error) {
	return e.repo.Get(ctx, id)
}

// --- PreCheck (delegated to Handler) ---

func (e *Engine) PreCheck(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	req, err := e.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	handler, ok := e.handlers[req.Type]
	if !ok {
		return nil, domain.Validation("no handler registered for type: " + string(req.Type))
	}
	return handler.PreCheck(ctx, req)
}

// --- internal ---

func (e *Engine) markFailed(ctx context.Context, req types.ApprovalRequest) {
	now := time.Now().UTC()
	req.Status = types.ApprovalFailed
	req.ResolvedAt = &now
	if err := e.repo.Update(ctx, req); err != nil {
		e.logger.Error("failed to mark approval as failed", "approval_id", req.ID, "error", err)
	}
}
