package budgetcheck

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// RefreshPlatformKeys recomputes soft-remain for touched platform keys. No-op
// when cache is disabled or keyIDs is empty.
func RefreshPlatformKeys(
	ctx context.Context,
	cfg config.Config,
	st store.Store,
	cache Store,
	logger *slog.Logger,
	companyID int64,
	keyIDs map[string]struct{},
) {
	if cache == nil || !cache.Enabled() || len(keyIDs) == 0 {
		return
	}
	stores := pkgbudget.MappingStores{
		Consumed: st.BudgetConsumed(),
		OrgNodes: st.Org().Nodes(),
		Org:      st.Org(),
		Budget:   st.Budget(),
		Keys:     st.Keys(),
		Clock:    cfg.Clock(),
	}
	for keyID := range keyIDs {
		refreshOne(ctx, cfg, st, cache, logger, companyID, keyID, stores)
	}
}

func refreshOne(
	ctx context.Context,
	cfg config.Config,
	st store.Store,
	cache Store,
	logger *slog.Logger,
	companyID int64,
	keyID string,
	stores pkgbudget.MappingStores,
) {
	mapping, err := st.PlatformKeyMappings().GetMappingByPlatformKeyID(ctx, keyID)
	if err != nil || mapping == nil || mapping.DepartmentID == "" {
		return
	}
	keyHash, ok, err := st.Keys().PlatformKeyHashByID(ctx, keyID)
	if err != nil || !ok {
		return
	}
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, stores.OrgNodes, mapping.DepartmentID, cfg.Clock())
	if err != nil {
		return
	}
	remain, err := pkgbudget.RemainForMapping(ctx, stores, mapping, open.String())
	if err != nil {
		return
	}
	entry := Entry{PeriodKey: open.String(), SoftRemain: remain, KeyStatus: "active"}
	if err := cache.Set(ctx, companyID, keyHash, entry); err != nil && logger != nil {
		logger.Warn("gateway budget check set failed", "key_id", keyID, "error", err)
	}
}
