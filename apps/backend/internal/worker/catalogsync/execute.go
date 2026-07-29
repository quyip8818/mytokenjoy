// Package catalogsync implements the platform catalog → local sync worker.
// Uses River PeriodicJob + version-based incremental sync (models only).
package catalogsync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
)

// Version key in system_settings.
const keyModelsVersion = "catalog.models_version"

// Executor holds the dependencies for catalog sync.
type Executor struct {
	client          *catalog.Client
	port            adminport.Port
	store           store.Store
	globalCompanyID uuid.UUID
}

func NewExecutor(client *catalog.Client, port adminport.Port, st store.Store, globalCompanyID uuid.UUID) *Executor {
	return &Executor{client: client, port: port, store: st, globalCompanyID: globalCompanyID}
}

// Execute performs a single sync cycle: check version → pull models → upsert + pricing.
func (e *Executor) Execute(ctx context.Context) error {
	// 1. Fetch remote version.
	remote, err := e.client.FetchVersions(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch versions: %w", err)
	}

	settings := e.store.SystemSettings()

	// 2. Compare local version.
	localStr, err := settings.Get(ctx, keyModelsVersion)
	if err != nil {
		return fmt.Errorf("catalogsync get local version: %w", err)
	}
	local, _ := strconv.Atoi(localStr) // missing/empty → 0

	if local == remote.Models {
		slog.Debug("catalogsync: up to date", "version", local)
		return nil
	}

	// 3. Pull models catalog.
	resp, err := e.client.FetchModels(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch models: %w", err)
	}

	// 4. Sync models to local DB.
	if err := e.syncModels(ctx, resp.Data); err != nil {
		return fmt.Errorf("catalogsync sync models: %w", err)
	}

	// 5. Sync pricing to local NewAPI.
	if err := e.syncPricing(ctx, resp.Data); err != nil {
		return fmt.Errorf("catalogsync sync pricing: %w", err)
	}

	// 6. Update local version from response.
	if err := settings.Set(ctx, keyModelsVersion, strconv.Itoa(resp.Version)); err != nil {
		return fmt.Errorf("catalogsync set local version: %w", err)
	}

	slog.Info("catalogsync: sync complete", "version", resp.Version, "models", len(resp.Data))
	return nil
}

func (e *Executor) syncModels(ctx context.Context, models []catalog.CatalogModel) error {
	infos := make([]types.ModelInfo, 0, len(models))
	for _, m := range models {
		infos = append(infos, types.ModelInfo{
			CompanyID:    e.globalCompanyID,
			Provider:     m.Provider,
			Type:         m.ModelID,
			Name:         m.DisplayName,
			Source:       "platform",
			Active:       true,
			Capabilities: m.Capabilities,
			MaxContext:   m.MaxContext,
		})
	}
	return e.store.Models().SyncFromPlatform(ctx, e.globalCompanyID, infos)
}

func (e *Executor) syncPricing(ctx context.Context, models []catalog.CatalogModel) error {
	now := time.Now()
	for _, m := range models {
		// Write to model_pricing table (TJ SOT).
		row := store.ModelPricingRow{
			CompanyID:     e.globalCompanyID,
			ModelType:     m.ModelID,
			InputPrice:    m.InputPrice,
			OutputPrice:   m.OutputPrice,
			EffectiveFrom: now,
		}
		_ = e.store.ModelPricing().Insert(ctx, row) // ON CONFLICT skips duplicates

		// Best-effort push to local NewAPI (gateway cache).
		if err := e.port.UpsertModelRatio(ctx, m.ModelID, m.InputPrice, m.OutputPrice); err != nil {
			slog.Warn("catalogsync: newapi pricing push failed", "model", m.ModelID, "error", err)
		}
	}
	return nil
}
