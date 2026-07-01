package budget

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *IngestService) applyIngestTx(
	ctx context.Context,
	st store.Store,
	payload newapi.WebhookLogPayload,
	mapping *store.RelayMapping,
	memberID *string,
	modelName string,
	costCNY float64,
) error {
	platformKeys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	keyIdx := -1
	for i := range platformKeys {
		if platformKeys[i].ID == mapping.PlatformKeyID {
			keyIdx = i
			break
		}
	}
	if keyIdx < 0 {
		return domain.NotFound("platform key not found: " + mapping.PlatformKeyID)
	}
	platformKeys[keyIdx].Used += costCNY
	if err := st.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return err
	}

	tree, err := st.Budget().Tree(ctx)
	if err != nil {
		return err
	}
	members, err := st.Org().Members(ctx)
	if err != nil {
		return err
	}
	if memberID != nil {
		if member, ok := pkgorg.FindMemberByID(members, *memberID); ok {
			if err := rollupDepartmentConsumed(tree, member.DepartmentID, costCNY); err != nil {
				return err
			}
			if err := st.Budget().SetTree(ctx, tree); err != nil {
				return err
			}
		}
	}

	if mapping.BudgetGroupID != nil {
		groups, err := st.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for i := range groups {
			if groups[i].ID == *mapping.BudgetGroupID {
				groups[i].Consumed += costCNY
				if err := st.Budget().SetGroups(ctx, groups); err != nil {
					return err
				}
				if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisBudgetGroup, groups[i].ID); err != nil {
					return err
				}
				break
			}
		}
	}

	if memberID != nil {
		if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisMember, *memberID); err != nil {
			return err
		}
	}
	if err := st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisDepartment, mapping.DepartmentID); err != nil {
		return err
	}

	tree, err = st.Budget().Tree(ctx)
	if err != nil {
		return err
	}
	members, err = st.Org().Members(ctx)
	if err != nil {
		return err
	}
	if err := s.evaluateOverrun(ctx, st, tree, members, memberID, mapping); err != nil {
		return err
	}

	memberIDValue := ""
	if memberID != nil {
		memberIDValue = *memberID
	}
	if err := st.Usage().UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart:  usageHourFromPayload(payload),
		DepartmentID: mapping.DepartmentID,
		MemberID:     memberIDValue,
		Model:        modelName,
		CostCNY:      costCNY,
		CallCount:    1,
	}); err != nil {
		return err
	}

	if err := st.Relay().InsertIngestedLogID(ctx, payload.ID); err != nil {
		return err
	}
	if payload.ID > 0 {
		last, err := st.Relay().GetLastLogID(ctx)
		if err != nil {
			return err
		}
		if payload.ID > last {
			if err := st.Relay().SetLastLogID(ctx, payload.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
