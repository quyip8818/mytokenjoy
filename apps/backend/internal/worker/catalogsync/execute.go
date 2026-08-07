// Package catalogsync implements the platform catalog → local sync worker.
// Uses River PeriodicJob + version-based incremental sync.
// Models and pricing are synced independently with separate version counters.
package catalogsync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	catalog "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
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
// ponytail: Local stores all versions under GlobalSyncVersion (single-company deployment).
func (e *Executor) Execute(ctx context.Context) error {
	remote, err := e.client.FetchVersions(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch versions: %w", err)
	}

	sv := e.store.SyncVersions()

	// --- Models sync ---
	localModels, err := sv.Get(ctx, store.GlobalSyncVersion, "models")
	if err != nil {
		return fmt.Errorf("catalogsync get models version: %w", err)
	}

	if localModels != remote.Models {
		resp, err := e.client.FetchModels(ctx)
		if err != nil {
			return fmt.Errorf("catalogsync fetch models: %w", err)
		}
		if err := e.syncModels(ctx, resp.Data); err != nil {
			return fmt.Errorf("catalogsync sync models: %w", err)
		}
		if err := sv.Set(ctx, store.GlobalSyncVersion, "models", resp.Version); err != nil {
			return fmt.Errorf("catalogsync set models version: %w", err)
		}
		slog.Info("catalogsync: models synced", "version", resp.Version, "count", len(resp.Data))
	}

	// --- Pricing sync (independent channel) ---
	localPricing, _ := sv.Get(ctx, store.GlobalSyncVersion, "pricing")

	if localPricing != remote.Pricing {
		if err := e.syncPricing(ctx, remote.Pricing); err != nil {
			return fmt.Errorf("catalogsync sync pricing: %w", err)
		}
		slog.Info("catalogsync: pricing synced", "version", remote.Pricing)
	}

	// --- Discounts sync (independent channel) ---
	localDiscounts, _ := sv.Get(ctx, store.GlobalSyncVersion, "discounts")

	if localDiscounts != remote.Discounts {
		if err := e.syncDiscounts(ctx, remote.Discounts); err != nil {
			return fmt.Errorf("catalogsync sync discounts: %w", err)
		}
		slog.Info("catalogsync: discounts synced", "version", remote.Discounts)
	}

	// --- Currencies sync (independent channel) ---
	localCurrencies, _ := sv.Get(ctx, store.GlobalSyncVersion, "currencies")

	if localCurrencies != remote.Currencies {
		if err := e.syncCurrencies(ctx); err != nil {
			return fmt.Errorf("catalogsync sync currencies: %w", err)
		}
		slog.Info("catalogsync: currencies synced", "version", remote.Currencies)
	}

	// --- Wallet lots sync (version-gated) ---
	localWalletLots, _ := sv.Get(ctx, store.GlobalSyncVersion, "wallet_lots")

	if localWalletLots != remote.WalletLots {
		if err := e.syncWalletLots(ctx); err != nil {
			return fmt.Errorf("catalogsync sync wallet_lots: %w", err)
		}
		if err := sv.Set(ctx, store.GlobalSyncVersion, "wallet_lots", remote.WalletLots); err != nil {
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
			Deprecated:   false,
			Capabilities: m.Capabilities,
			MaxContext:   m.MaxContext,
		})
	}
	return e.store.Models().SyncFromPlatform(ctx, e.globalCompanyID, infos)
}

// syncPricing fetches global pricing and replaces local NewAPI ratio maps.
// ponytail: SaaS returns the full pricing snapshot — no read-merge needed. 2 PUTs total.
func (e *Executor) syncPricing(ctx context.Context, remoteVersion int) error {
	resp, err := e.client.FetchPricing(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch pricing: %w", err)
	}
	if len(resp.Data) == 0 {
		return fmt.Errorf("catalogsync pricing: empty response, skipping")
	}

	mrMap := make(map[string]float64, len(resp.Data))
	crMap := make(map[string]float64, len(resp.Data))
	caMap := make(map[string]float64, len(resp.Data))
	for _, p := range resp.Data {
		mr, cr, ca := modelcatalog.RatioFromPrice(p.InputPrice, p.OutputPrice, p.CacheInputPrice)
		mrMap[p.ModelType] = mr
		crMap[p.ModelType] = cr
		caMap[p.ModelType] = ca
	}

	mrJSON, _ := json.Marshal(mrMap)
	if err := e.port.UpdateOption(ctx, "ModelRatio", string(mrJSON)); err != nil {
		return fmt.Errorf("catalogsync write ModelRatio: %w", err)
	}
	crJSON, _ := json.Marshal(crMap)
	if err := e.port.UpdateOption(ctx, "CompletionRatio", string(crJSON)); err != nil {
		return fmt.Errorf("catalogsync write CompletionRatio: %w", err)
	}
	caJSON, _ := json.Marshal(caMap)
	if err := e.port.UpdateOption(ctx, "CacheRatio", string(caJSON)); err != nil {
		return fmt.Errorf("catalogsync write CacheRatio: %w", err)
	}

	return e.store.SyncVersions().Set(ctx, store.GlobalSyncVersion, "pricing", remoteVersion)
}

// syncDiscounts fetches per-company discount coefficients and inserts new rows (id-idempotent).
func (e *Executor) syncDiscounts(ctx context.Context, remoteVersion int) error {
	resp, err := e.client.FetchDiscounts(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch discounts: %w", err)
	}

	for _, d := range resp.Data {
		id, err := uuid.Parse(d.ID)
		if err != nil {
			slog.Warn("catalogsync: skip discount with invalid ID", "id", d.ID)
			continue
		}
		row := store.ModelDiscountRow{
			ID:        id,
			CompanyID: e.localCompanyID,
			ModelType: d.ModelType,
			Discount:  d.Discount,
		}
		if err := e.store.ModelDiscount().InsertFromSync(ctx, row); err != nil {
			slog.Warn("catalogsync: discount insert failed", "modelType", d.ModelType, "error", err)
		}
	}

	return e.store.SyncVersions().Set(ctx, store.GlobalSyncVersion, "discounts", remoteVersion)
}

// syncCurrencies fetches currencies from the platform and inserts new rows (append-only, id-idempotent).
func (e *Executor) syncCurrencies(ctx context.Context) error {
	resp, err := e.client.FetchCurrencies(ctx)
	if err != nil {
		return fmt.Errorf("catalogsync fetch currencies: %w", err)
	}

	for _, c := range resp.Data {
		id, err := uuid.Parse(c.ID)
		if err != nil {
			slog.Warn("catalogsync: skip currency with invalid ID", "id", c.ID)
			continue
		}
		row := store.Currency{
			ID:           id,
			Code:         c.Code,
			QuotaPerUnit: c.QuotaPerUnit,
			Enabled:      c.Enabled,
			UpdatedAt:    time.Unix(c.UpdatedAt, 0),
		}
		if err := e.store.Billing().InsertCurrencyFromSync(ctx, row); err != nil {
			slog.Warn("catalogsync: currency insert failed", "code", c.Code, "error", err)
		}
	}

	// Use resp.Version (actual data version) rather than remote.Currencies.
	return e.store.SyncVersions().Set(ctx, store.GlobalSyncVersion, "currencies", resp.Version)
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

	// Upsert lot transactions (UUID-idempotent, append-only).
	for _, remoteTx := range resp.Transactions {
		txID, err := uuid.Parse(remoteTx.ID)
		if err != nil {
			slog.Warn("catalogsync: skip transaction with invalid ID", "id", remoteTx.ID)
			continue
		}
		lotID, _ := uuid.Parse(remoteTx.LotID)
		var operatorID *uuid.UUID
		if remoteTx.OperatorID != nil {
			if parsed, err := uuid.Parse(*remoteTx.OperatorID); err == nil {
				operatorID = &parsed
			}
		}
		tx := store.LotTransaction{
			ID:              txID,
			CompanyID:       companyID,
			LotID:           lotID,
			Action:          remoteTx.Action,
			QuotaDelta:      remoteTx.QuotaDelta,
			QuotaPerUnit:    remoteTx.QuotaPerUnit,
			MoneyAmount:     remoteTx.MoneyAmount,
			BillingCurrency: remoteTx.BillingCurrency,
			RemainingAfter:  remoteTx.RemainingAfter,
			Source:          remoteTx.Source,
			LotKind:         remoteTx.LotKind,
			OperatorID:      operatorID,
			Note:            remoteTx.Note,
			CreatedAt:       time.Unix(remoteTx.CreatedAt, 0),
		}
		if err := e.store.Billing().UpsertLotTransactionFromSync(ctx, tx); err != nil {
			slog.Warn("catalogsync: lot transaction upsert failed", "txId", txID, "error", err)
		}
	}

	slog.Info("catalogsync: wallet_lots synced", "lots", len(resp.Data), "orders", len(resp.Orders), "transactions", len(resp.Transactions), "walletRemainQuota", resp.WalletRemainQuota)
	return nil
}
