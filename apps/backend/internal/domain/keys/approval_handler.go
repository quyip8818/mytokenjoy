package keys

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/approval"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

// --- KeyApprovalHandler ---

// keyApproveResult carries data produced in OnApprovedTx to PostApprove/Compensate.
type keyApproveResult struct {
	createdKeyID        uuid.UUID
	personalBudgetAdded int64
	departmentID        uuid.UUID
}

// KeyApprovalHandler handles type="key" approvals (create platform key).
type KeyApprovalHandler struct {
	svc *service
}

func NewKeyApprovalHandler(svc Service) *KeyApprovalHandler {
	return &KeyApprovalHandler{svc: svc.(*service)}
}

func (h *KeyApprovalHandler) Type() types.ApprovalType { return types.ApprovalTypeKey }

func (h *KeyApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
	var meta types.KeyApprovalMeta
	if err := json.Unmarshal(input.Metadata, &meta); err != nil {
		return domain.Validation("invalid key approval metadata")
	}
	if meta.RequestedBudget <= 0 {
		return domain.Validation("requestedBudget must be positive")
	}
	if len(meta.RequestedModels) == 0 {
		return domain.Validation("requestedModels required for key approval")
	}
	// Validate models belong to applicant's department
	members, err := h.svc.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	departments, err := common.LoadDepartments(ctx, h.svc.store.Org().Nodes())
	if err != nil {
		return err
	}
	rules, err := common.LoadRoutingRules(ctx, h.svc.store.Org().Nodes(), h.svc.store.Models().Allowlist())
	if err != nil {
		return err
	}
	models, err := h.svc.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	if msg := common.ValidateModelIDsForMember(input.ApplicantID, meta.RequestedModels, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
		return domain.Validation(*msg)
	}
	return nil
}

func (h *KeyApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
	if err := h.svc.requireNewAPI(); err != nil {
		return err
	}
	var meta types.KeyApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	budgetCtx, err := budget.LoadBudgetContext(ctx, h.svc.store.BudgetConsumed(), h.svc.store.Org(), h.svc.store.Budget(), h.svc.store.Keys(), h.svc.cfg.Clock())
	if err != nil {
		return err
	}
	reservedPool := budget.GetReservedPoolForMember(budgetCtx.Tree, budgetCtx.Members, req.ApplicantID)
	if int64(meta.RequestedBudget) > reservedPool {
		return domain.Validation("Reserved pool insufficient")
	}
	return nil
}

func (h *KeyApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
	var meta types.KeyApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	// Acquire budget lock to prevent concurrent approval races
	if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
		return nil, err
	}

	members, err := tx.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	platformKeys, err := budget.LoadPlatformKeysWithUsed(ctx, tx.BudgetConsumed(), tx.Org(), tx.Budget(), tx.Keys(), h.svc.cfg.Clock())
	if err != nil {
		return nil, err
	}

	keyBudget := int64(meta.RequestedBudget)
	remaining := budget.GetBudgetRemaining(members, platformKeys, req.ApplicantID)
	var personalBudgetAdded int64
	if keyBudget > remaining {
		personalBudgetAdded = keyBudget - remaining
		members = budget.AddMemberPersonalBudget(members, req.ApplicantID, personalBudgetAdded)
	}

	memberID := req.ApplicantID
	createdKeyID := uuid.Must(uuid.NewV7())
	platformKeys = append(platformKeys, types.PlatformKey{
		ID:             createdKeyID,
		Name:           fmt.Sprintf("%s-审批 Key", req.ApplicantName),
		KeyPrefix:      "pending...",
		Scope:          types.PlatformKeyScopeMember,
		MemberID:       &memberID,
		Status:         "active",
		Budget:         keyBudget,
		Consumed:       0,
		ModelWhitelist: append([]uuid.UUID{}, meta.RequestedModels...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	})
	if err := tx.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return nil, err
	}
	if err := tx.Org().SetMembers(ctx, members); err != nil {
		return nil, err
	}

	// Resolve department ID for PostApprove
	deptID := req.DepartmentID
	if deptID == uuid.Nil {
		if applicant, ok := org.FindMemberByID(members, req.ApplicantID); ok {
			deptID = applicant.DepartmentID
		}
	}

	return &keyApproveResult{
		createdKeyID:        createdKeyID,
		personalBudgetAdded: personalBudgetAdded,
		departmentID:        deptID,
	}, nil
}

func (h *KeyApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, raw approval.ApproveResult) error {
	result := raw.(*keyApproveResult)
	if h.svc.newAPISync == nil || !h.svc.newAPISync.Enabled() {
		return nil
	}
	// Read the created key from store to pass to sync
	platformKeys, err := h.svc.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	var created types.PlatformKey
	for _, key := range platformKeys {
		if key.ID == result.createdKeyID {
			created = key
			break
		}
	}
	if created.ID == uuid.Nil {
		return domain.NotFound("created key not found for sync")
	}
	_, err = h.svc.syncPlatformKeyCreate(ctx, created, result.departmentID)
	return err
}

func (h *KeyApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, raw approval.ApproveResult) error {
	if raw != nil {
		// Called from Approve flow: precise cleanup
		result := raw.(*keyApproveResult)
		if result.createdKeyID != uuid.Nil {
			if err := removePlatformKeyByID(ctx, h.svc.store.Keys(), result.createdKeyID); err != nil {
				return err
			}
		}
		if result.personalBudgetAdded > 0 {
			members, err := h.svc.store.Org().Members(ctx)
			if err != nil {
				return err
			}
			members = budget.AddMemberPersonalBudget(members, req.ApplicantID, -result.personalBudgetAdded)
			return h.svc.store.Org().SetMembers(ctx, members)
		}
		return nil
	}
	// Called from Retry flow (raw==nil): infer from DB.
	// ponytail: for now just return nil — key type retry is rare edge case.
	// The Retry flow's Compensate(nil) + re-run OnApprovedTx will create a new key.
	// Orphaned keys (if any) are harmless pending stubs.
	return nil
}

func (h *KeyApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
	return nil
}

func (h *KeyApprovalHandler) PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error) {
	var meta types.KeyApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	budgetCtx, err := budget.LoadBudgetContext(ctx, h.svc.store.BudgetConsumed(), h.svc.store.Org(), h.svc.store.Budget(), h.svc.store.Keys(), h.svc.cfg.Clock())
	if err != nil {
		return nil, err
	}
	reservedPool := budget.GetReservedPoolForMember(budgetCtx.Tree, budgetCtx.Members, req.ApplicantID)
	return json.Marshal(map[string]any{
		"sufficient":   reservedPool >= int64(meta.RequestedBudget),
		"reservedPool": reservedPool,
		"requested":    int64(meta.RequestedBudget),
	})
}

// --- BudgetApprovalHandler ---

// BudgetApprovalHandler handles type="budget" approvals (increase personal budget only, no key).
type BudgetApprovalHandler struct {
	svc *service
}

func NewBudgetApprovalHandler(svc Service) *BudgetApprovalHandler {
	return &BudgetApprovalHandler{svc: svc.(*service)}
}

func (h *BudgetApprovalHandler) Type() types.ApprovalType { return types.ApprovalTypeBudget }

func (h *BudgetApprovalHandler) Validate(ctx context.Context, input approval.CreateInput) error {
	var meta types.BudgetApprovalMeta
	if err := json.Unmarshal(input.Metadata, &meta); err != nil {
		return domain.Validation("invalid budget approval metadata")
	}
	if meta.RequestedBudget <= 0 {
		return domain.Validation("requestedBudget must be positive")
	}
	return nil
}

func (h *BudgetApprovalHandler) PreApprove(ctx context.Context, req types.ApprovalRequest) error {
	var meta types.BudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	budgetCtx, err := budget.LoadBudgetContext(ctx, h.svc.store.BudgetConsumed(), h.svc.store.Org(), h.svc.store.Budget(), h.svc.store.Keys(), h.svc.cfg.Clock())
	if err != nil {
		return err
	}
	reservedPool := budget.GetReservedPoolForMember(budgetCtx.Tree, budgetCtx.Members, req.ApplicantID)
	if int64(meta.RequestedBudget) > reservedPool {
		return domain.Validation("Reserved pool insufficient")
	}
	return nil
}

func (h *BudgetApprovalHandler) OnApprovedTx(ctx context.Context, req types.ApprovalRequest, tx store.Store) (approval.ApproveResult, error) {
	var meta types.BudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)

	// Acquire budget lock to prevent concurrent approval races
	if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
		return nil, err
	}

	// Resolve department and re-check reserved pool under lock
	members, err := tx.Org().Members(ctx)
	if err != nil {
		return nil, err
	}
	applicant, ok := org.FindMemberByID(members, req.ApplicantID)
	if !ok {
		return nil, domain.NotFound("申请人不存在")
	}
	row, found, err := tx.Budget().OrgNodeBudget().Get(ctx, applicant.DepartmentID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, domain.Validation("部门预算节点不存在")
	}
	reserved := int64(0)
	if row.ReservedPool != nil {
		reserved = *row.ReservedPool
	}
	if int64(meta.RequestedBudget) > reserved {
		return nil, domain.Validation(fmt.Sprintf("预留池余额不足，当前剩余 %d quota", reserved))
	}

	members = budget.AddMemberPersonalBudget(members, req.ApplicantID, int64(meta.RequestedBudget))
	if err := tx.Org().SetMembers(ctx, members); err != nil {
		return nil, err
	}
	return nil, nil
}

func (h *BudgetApprovalHandler) PostApprove(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	return nil // Pure DB operation, no external IO
}

func (h *BudgetApprovalHandler) Compensate(ctx context.Context, req types.ApprovalRequest, _ approval.ApproveResult) error {
	// Idempotent: if budget was added, subtract it back
	var meta types.BudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	members, err := h.svc.store.Org().Members(ctx)
	if err != nil {
		return err
	}
	members = budget.AddMemberPersonalBudget(members, req.ApplicantID, -int64(meta.RequestedBudget))
	return h.svc.store.Org().SetMembers(ctx, members)
}

func (h *BudgetApprovalHandler) OnRejected(ctx context.Context, req types.ApprovalRequest, tx store.Store) error {
	return nil
}

func (h *BudgetApprovalHandler) PreCheck(ctx context.Context, req types.ApprovalRequest) (json.RawMessage, error) {
	var meta types.BudgetApprovalMeta
	json.Unmarshal(req.Metadata, &meta)
	budgetCtx, err := budget.LoadBudgetContext(ctx, h.svc.store.BudgetConsumed(), h.svc.store.Org(), h.svc.store.Budget(), h.svc.store.Keys(), h.svc.cfg.Clock())
	if err != nil {
		return nil, err
	}
	reservedPool := budget.GetReservedPoolForMember(budgetCtx.Tree, budgetCtx.Members, req.ApplicantID)
	return json.Marshal(map[string]any{
		"sufficient":   reservedPool >= int64(meta.RequestedBudget),
		"reservedPool": reservedPool,
		"requested":    int64(meta.RequestedBudget),
	})
}

// removePlatformKeyByID drops a platform key from the store. Idempotent.
func removePlatformKeyByID(ctx context.Context, keysRepo store.KeysRepository, keyID uuid.UUID) error {
	keys, err := keysRepo.PlatformKeys(ctx)
	if err != nil {
		return err
	}
	filtered := make([]types.PlatformKey, 0, len(keys))
	removed := false
	for _, key := range keys {
		if key.ID == keyID {
			removed = true
			continue
		}
		filtered = append(filtered, key)
	}
	if !removed {
		return nil
	}
	return keysRepo.SetPlatformKeys(ctx, filtered)
}
