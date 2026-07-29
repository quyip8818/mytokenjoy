// Package catalogsync implements the platform catalog → local sync worker.
// Uses River PeriodicJob + version-based incremental sync.
// Models and pricing are synced independently with separate version counters.
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

// Version keys in system_settings.
const (
	keyModelsVersion  = "catalog.models_version"
	keyPricingVersion = "catalog.pricing_version"
)

// Executor holds the dependencies for catalog sync.
type Executor struct {
	client          *catalog.Client
	port            adminport.Port
	store           store.Store
	globalCompanyID uuid.UUID
	localCompanyID  uuid.UUID // the company registered on SaaS (for contract pricing)
}

func NewExecutor(client *catalog.Client, port adminport.Port, st store.Store, globalCompanyID, localCompanyID uuid.UUID) *Executor {
	return &Executor{
		client:          client,
		port:            port,
		store:           st,
		globalCompanyID: globalCompanyID,
		localCompanyID:  localCompanyID,
	}
}

// Execute performs a single sync cycle with independent models and pricing channels.
func (e *Executor) Execute(ctx context.Context) error {
	remote, err := e.client.FetchVersions(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch versions: %w", err)
	}

	settings := e.store.SystemSettings()

	// --- Models sync ---
	localModelsStr, err := settings.Get(ctx, keyModelsVersion)
	if err != nil {
		return fmt.Errorf("catalogsync get models version: %w", err)
	}
	localModels, _ := strconv.Atoi(localModelsStr)

	if localModels != remote.Models {
		resp, err := e.client.FetchModels(ctx)
		if err != nil {
			return fmt.Errorf("catalogsync fetch models: %w", err)
		}
		if err := e.syncModels(ctx, resp.Data); err != nil {
			return fmt.Errorf("catalogsync sync models: %w", err)
		}
		if err := settings.Set(ctx, keyModelsVersion, strconv.Itoa(resp.Version)); err != nil {
			return fmt.Errorf("catalogsync set models version: %w", err)
		}
		slog.Info("catalogsync: models synced", "version", resp.Version, "count", len(resp.Data))
	}

	// --- Pricing sync (independent channel) ---
	localPricingStr, _ := settings.Get(ctx, keyPricingVersion)
	localPricing, _ := strconv.Atoi(localPricingStr)

	if localPricing != remote.Pricing {
		if err := e.syncPricing(ctx, remote.Pricing); err != nil {
			return fmt.Errorf("catalogsync sync pricing: %w", err)
		}
		slog.Info("catalogsync: pricing synced", "version", remote.Pricing)
	}

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

// syncPricing fetches pricing from the dedicated endpoint and writes to model_pricing.
// isContract determines which companyID the row belongs to.
func (e *Executor) syncPricing(ctx context.Context, remoteVersion int) error {
	resp, err := e.client.FetchPricing(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch pricing: %w", err)
	}

	now := time.Now()
	for _, p := range resp.Data {
		companyID := e.globalCompanyID
		if p.IsContract {
			companyID = e.localCompanyID
		}

		row := store.ModelPricingRow{
			CompanyID:     companyID,
			ModelType:     p.ModelType,
			InputPrice:    p.InputPrice,
			OutputPrice:   p.OutputPrice,
			EffectiveFrom: now,
		}
		_ = e.store.ModelPricing().Insert(ctx, row) // ON CONFLICT skip

		// Global prices → best-effort push to local NewAPI (gateway cache).
		if !p.IsContract {
			if err := e.port.UpsertModelRatio(ctx, p.ModelType, p.InputPrice, p.OutputPrice); err != nil {
				slog.Warn("catalogsync: newapi pricing push failed", "model", p.ModelType, "error", err)
			}
		}
	}

	return e.store.SystemSettings().Set(ctx, keyPricingVersion, strconv.Itoa(remoteVersion))
}
