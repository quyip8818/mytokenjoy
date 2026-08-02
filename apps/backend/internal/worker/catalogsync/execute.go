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
	keyModelsVersion     = "catalog.models_version"
	keyPricingVersion    = "catalog.pricing_version"
	keyDiscountsVersion  = "catalog.discounts_version"
	keyCurrenciesVersion = "catalog.currencies_version"
	keyWalletLotsVersion = "catalog.wallet_lots_version"
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

	// --- Discounts sync (independent channel) ---
	localDiscountsStr, _ := settings.Get(ctx, keyDiscountsVersion)
	localDiscounts, _ := strconv.Atoi(localDiscountsStr)

	if localDiscounts != remote.Discounts {
		if err := e.syncDiscounts(ctx, remote.Discounts); err != nil {
			return fmt.Errorf("catalogsync sync discounts: %w", err)
		}
		slog.Info("catalogsync: discounts synced", "version", remote.Discounts)
	}

	// --- Currencies sync (independent channel) ---
	localCurrenciesStr, _ := settings.Get(ctx, keyCurrenciesVersion) // key may not exist on first run → "" → 0
	localCurrencies, _ := strconv.Atoi(localCurrenciesStr)

	if localCurrencies != remote.Currencies {
		if err := e.syncCurrencies(ctx); err != nil {
			return fmt.Errorf("catalogsync sync currencies: %w", err)
		}
		slog.Info("catalogsync: currencies synced", "version", remote.Currencies)
	}

	// --- Wallet lots sync (version-gated) ---
	localWalletLotsStr, _ := settings.Get(ctx, keyWalletLotsVersion)
	localWalletLots, _ := strconv.Atoi(localWalletLotsStr)

	if localWalletLots != remote.WalletLots {
		if err := e.syncWalletLots(ctx); err != nil {
			return fmt.Errorf("catalogsync sync wallet_lots: %w", err)
		}
		if err := settings.Set(ctx, keyWalletLotsVersion, strconv.Itoa(remote.WalletLots)); err != nil {
			return fmt.Errorf("catalogsync set wallet_lots version: %w", err)
		}
		slog.Info("catalogsync: wallet_lots synced", "version", remote.WalletLots)
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

// syncPricing fetches global pricing and updates models.input_price/output_price.
func (e *Executor) syncPricing(ctx context.Context, remoteVersion int) error {
	resp, err := e.client.FetchPricing(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch pricing: %w", err)
	}

	for _, p := range resp.Data {
		// Update model's price columns directly.
		_ = e.store.Models().UpdatePrice(ctx, e.globalCompanyID, p.ModelType, p.InputPrice, p.OutputPrice)

		// Best-effort push to local NewAPI (gateway cache).
		if err := e.port.UpsertModelRatio(ctx, p.ModelType, p.InputPrice, p.OutputPrice); err != nil {
			slog.Warn("catalogsync: newapi pricing push failed", "model", p.ModelType, "error", err)
		}
	}

	return e.store.SystemSettings().Set(ctx, keyPricingVersion, strconv.Itoa(remoteVersion))
}

// syncDiscounts fetches per-company discount coefficients and writes to model_discount.
func (e *Executor) syncDiscounts(ctx context.Context, remoteVersion int) error {
	resp, err := e.client.FetchDiscounts(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch discounts: %w", err)
	}

	for _, d := range resp.Data {
		row := store.ModelDiscountRow{
			CompanyID: e.localCompanyID,
			ModelType: d.ModelType,
			Discount:  d.Discount,
		}
		_ = e.store.ModelDiscount().Insert(ctx, row)
	}

	return e.store.SystemSettings().Set(ctx, keyDiscountsVersion, strconv.Itoa(remoteVersion))
}

// syncCurrencies fetches currencies from the platform and replaces local data.
func (e *Executor) syncCurrencies(ctx context.Context) error {
	resp, err := e.client.FetchCurrencies(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch currencies: %w", err)
	}

	currencies := make([]store.Currency, 0, len(resp.Data))
	for _, c := range resp.Data {
		currencies = append(currencies, store.Currency{
			Code:         c.Code,
			QuotaPerUnit: c.QuotaPerUnit,
			Enabled:      true,
		})
	}

	// Wrap in tx for atomicity (upsert + disable stale).
	if err := e.store.WithTx(ctx, func(tx store.Store) error {
		return tx.Billing().ReplaceCurrencies(ctx, currencies)
	}); err != nil {
		return fmt.Errorf("catalogsync replace currencies: %w", err)
	}

	// Use resp.Version (actual data version) rather than remote.Currencies.
	return e.store.SystemSettings().Set(ctx, keyCurrenciesVersion, strconv.Itoa(resp.Version))
}

// syncWalletLots fetches active lots + wallet balance from SaaS and mirrors them locally.
// ponytail: version-gated. Only fetched when SaaS wallet_lots_version changes (recharge or ingest).
// This ensures the Local FIFO chain matches SaaS, so Local Ingest can consume lots identically.
func (e *Executor) syncWalletLots(ctx context.Context) error {
	resp, err := e.client.FetchWalletLots(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch wallet_lots: %w", err)
	}

	companyID := e.localCompanyID

	// Upsert orders first (lots have FK to orders).
	for _, remoteOrder := range resp.Orders {
		orderID, err := uuid.Parse(remoteOrder.ID)
		if err != nil {
			slog.Warn("catalogsync: skip order with invalid ID", "id", remoteOrder.ID)
			continue
		}
		order := store.RechargeOrder{
			ID:             orderID,
			CompanyID:      companyID,
			Amount:         remoteOrder.Amount,
			Currency:       remoteOrder.Currency,
			QuotaPerUnit:   remoteOrder.QuotaPerUnit,
			QuotaGranted:   remoteOrder.QuotaGranted,
			Source:         remoteOrder.Source,
			LotKind:        remoteOrder.LotKind,
			Status:         remoteOrder.Status,
			DisplayOrderID: remoteOrder.DisplayOrderID,
			PaymentMethod:  remoteOrder.PaymentMethod,
			CreatedAt:      time.Unix(remoteOrder.CreatedAt, 0),
		}
		if err := e.store.Billing().UpsertOrderFromSync(ctx, order); err != nil {
			slog.Warn("catalogsync: order upsert failed", "orderId", orderID, "error", err)
		}
	}

	// Upsert lots (referencing the synced orders).
	for _, remoteLot := range resp.Data {
		lotID, err := uuid.Parse(remoteLot.ID)
		if err != nil {
			slog.Warn("catalogsync: skip lot with invalid ID", "id", remoteLot.ID)
			continue
		}
		orderID, _ := uuid.Parse(remoteLot.OrderID)
		if orderID == uuid.Nil {
			orderID = lotID // fallback: use lot ID (backward compat)
		}
		lot := store.RechargeLot{
			ID:              lotID,
			CompanyID:       companyID,
			RechargeOrderID: orderID,
			BillingCurrency: remoteLot.BillingCurrency,
			LotKind:         remoteLot.LotKind,
			PaidAmount:      remoteLot.PaidAmount,
			QuotaPerUnit:    remoteLot.QuotaPerUnit,
			QuotaGranted:    remoteLot.QuotaGranted,
			QuotaRemaining:  remoteLot.QuotaRemaining,
			Status:          remoteLot.Status,
			CreatedAt:       time.Unix(remoteLot.CreatedAt, 0),
		}
		if err := e.store.Billing().UpsertLotFromSync(ctx, lot); err != nil {
			slog.Warn("catalogsync: lot upsert failed", "lotId", lotID, "error", err)
		}
	}

	// Overwrite wallet_remain_quota with SaaS authoritative value.
	if err := e.store.Company().SetWalletRemainQuota(ctx, companyID, resp.WalletRemainQuota, nil); err != nil {
		return fmt.Errorf("catalogsync set wallet: %w", err)
	}

	slog.Info("catalogsync: wallet_lots synced", "lots", len(resp.Data), "orders", len(resp.Orders), "walletRemainQuota", resp.WalletRemainQuota)
	return nil
}
