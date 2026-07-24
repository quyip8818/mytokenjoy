package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	if input.Scope == "" {
		return types.PlatformKey{}, domain.Validation("scope is required")
	}
	if err := budget.ValidatePlatformKeyScope(input.Scope, input.MemberID, input.ProjectID); err != nil {
		return types.PlatformKey{}, domain.Validation(err.Error())
	}

	budgetCtx, err := budget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := budgetCtx.PlatformKeys
	projects := budgetCtx.Projects
	members := budgetCtx.Members
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.PlatformKey{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return types.PlatformKey{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}

	// Trial/demo accounts: whitelist must only contain test-only models.
	if info, ok := ctxcompany.From(ctx); ok && (info.Type == store.CompanyTypeTrial || info.Type == store.CompanyTypeDemo) {
		if err := validateTestOnlyModels(input.ModelWhitelist, models); err != nil {
			return types.PlatformKey{}, err
		}
	}

	switch input.Scope {
	case types.PlatformKeyScopeMember:
		if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if msg := budget.ValidateMemberScopeKeyBudget(members, platformKeys, *input.MemberID, int64(input.Budget), uuid.Nil); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	case types.PlatformKeyScopeProject:
		project, ok := budget.FindProject(projects, *input.ProjectID)
		if !ok {
			return types.PlatformKey{}, domain.NotFound("Project not found")
		}
		if msg := budget.ValidateProjectScopeKeyBudget(input.Scope, project, platformKeys, input.MemberID, int64(input.Budget), uuid.Nil); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.MemberID != nil {
			if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		}
	case types.PlatformKeyScopeProjectMember:
		project, ok := budget.FindProject(projects, *input.ProjectID)
		if !ok {
			return types.PlatformKey{}, domain.NotFound("Project not found")
		}
		if err := budget.ValidateProjectMemberRoster(project, *input.MemberID); err != nil {
			return types.PlatformKey{}, domain.Validation(err.Error())
		}
		if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if msg := budget.ValidateProjectScopeKeyBudget(input.Scope, project, platformKeys, input.MemberID, int64(input.Budget), uuid.Nil); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	}

	if err := s.requireNewAPI(); err != nil {
		return types.PlatformKey{}, err
	}

	created := types.PlatformKey{
		ID:   uuid.Must(uuid.NewV7()),
		Name: input.Name, KeyPrefix: "pending...", Scope: input.Scope,
		MemberID: input.MemberID, ProjectID: input.ProjectID,
		Status: "active", Budget: int64(input.Budget), Consumed: 0,
		ModelWhitelist: append([]uuid.UUID{}, input.ModelWhitelist...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	}
	platformKeys = append(platformKeys, created)
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}

	departmentID, err := s.resolvePlatformKeyDepartmentID(input, members, projects)
	if err != nil {
		return types.PlatformKey{}, err
	}
	result, err := s.syncPlatformKeyCreate(ctx, created, departmentID)
	if err != nil {
		return types.PlatformKey{}, err
	}
	if err := domainbudget.RefreshPlatformKeyCombined(ctx, s.store, created.ID, s.cfg.Clock(), nil); err != nil {
		return types.PlatformKey{}, err
	}
	s.appendKeyCreateLog(ctx, input, result)
	return result, nil
}

// validateTestOnlyModels ensures all model IDs in the whitelist are test-only models.
// Used to enforce that trial/demo accounts cannot create keys with real model access.
func validateTestOnlyModels(whitelist []uuid.UUID, catalog []types.ModelInfo) error {
	byID := make(map[uuid.UUID]types.ModelInfo, len(catalog))
	for _, m := range catalog {
		byID[m.ID] = m
	}
	for _, id := range whitelist {
		m, ok := byID[id]
		if !ok {
			return domain.Validation("model not found in catalog")
		}
		if !modelcatalog.IsTestOnlyCallType(m.Type) {
			return domain.Validation("试用账户只能使用 test-model，升级后可使用全部模型")
		}
	}
	return nil
}

func (s *service) appendKeyCreateLog(ctx context.Context, input types.CreatePlatformKeyInput, key types.PlatformKey) {
	if input.OperatorID == uuid.Nil {
		return
	}
	_ = s.store.Audit().AppendOperationLog(ctx, types.OperationLog{
		ID:         uuid.Must(uuid.NewV7()),
		Action:     "key_create",
		Operator:   input.OperatorName,
		OperatorID: input.OperatorID,
		ActorType:  store.ActorTypeMember,
		Target:     fmt.Sprintf("Platform Key: %s", key.Name),
		Detail:     fmt.Sprintf("创建平台凭证，额度 %d 元", key.Budget),
		IP:         input.IP,
		CreatedAt:  time.Now().Format("2006-01-02 15:04"),
	})
}
