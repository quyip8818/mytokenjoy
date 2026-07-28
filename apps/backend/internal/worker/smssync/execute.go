// Package smssync implements the SMS → TokenJoy sync worker.
// Uses River PeriodicJob + partition-based incremental sync.
// See apps/docs/plan/sms-model-sync-v2.md for design details.
package smssync

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/tokenjoy/backend/internal/integration/sms"
	"github.com/tokenjoy/backend/internal/store"
)

// Version keys in system_settings.
const (
	keyChannelsVersion = "sms_sync.channels_version"
	keyModelsVersion   = "sms_sync.models_version"
)

// SMSSyncExecutor holds the dependencies for the v2 partition-based sync logic.
type SMSSyncExecutor struct {
	client *sms.Client
	target SyncTarget
	store  store.Store
}

func NewExecutor(client *sms.Client, target SyncTarget, st store.Store) *SMSSyncExecutor {
	return &SMSSyncExecutor{client: client, target: target, store: st}
}

// Execute performs a single sync cycle using partition-based incremental sync.
// It fetches remote versions, compares with local, and only pulls changed partitions.
func (e *SMSSyncExecutor) Execute(ctx context.Context) error {
	// 1. Fetch remote partition versions.
	remote, err := e.client.FetchVersions(ctx)
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}

	settings := e.store.SystemSettings()

	// 2. Compare each partition and sync if changed.
	var syncErr error

	// --- Channels ---
	if err := e.syncPartition(ctx, settings, keyChannelsVersion, remote.Channels, e.syncChannels); err != nil {
		slog.Error("smssync: channels partition failed", "error", err)
		syncErr = err
	}

	// --- Models ---
	if err := e.syncPartition(ctx, settings, keyModelsVersion, remote.Models, e.syncModels); err != nil {
		slog.Error("smssync: models partition failed", "error", err)
		syncErr = err
	}

	// --- Currencies: skipped this phase ---

	if syncErr != nil {
		return fmt.Errorf("smssync: one or more partitions failed: %w", syncErr)
	}
	slog.Info("smssync: sync cycle complete")
	return nil
}

// syncPartition compares local version with remote, calls syncFn if different,
// and updates local version on success using the version from the partition response.
func (e *SMSSyncExecutor) syncPartition(
	ctx context.Context,
	settings store.SystemSettingsRepository,
	key string,
	remoteVersion int,
	syncFn func(ctx context.Context) (int, error), // returns response version
) error {
	localStr, err := settings.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("get local version %s: %w", key, err)
	}
	local, _ := strconv.Atoi(localStr) // missing/empty → 0

	if local == remoteVersion {
		return nil // up to date
	}

	// Pull partition data and apply.
	responseVersion, err := syncFn(ctx)
	if err != nil {
		return err
	}

	// Update local version from response body (not from /versions endpoint).
	if err := settings.Set(ctx, key, strconv.Itoa(responseVersion)); err != nil {
		return fmt.Errorf("set local version %s: %w", key, err)
	}
	return nil
}

// syncChannels fetches and replaces channels. Returns the response version.
func (e *SMSSyncExecutor) syncChannels(ctx context.Context) (int, error) {
	resp, err := e.client.FetchChannels(ctx)
	if err != nil {
		return 0, err
	}
	if err := e.target.ReplaceChannels(ctx, resp.Data); err != nil {
		return 0, fmt.Errorf("replace channels: %w", err)
	}
	return resp.Version, nil
}

// syncModels fetches models and syncs to global company + updates global model ratios.
func (e *SMSSyncExecutor) syncModels(ctx context.Context) (int, error) {
	resp, err := e.client.FetchModels(ctx)
	if err != nil {
		return 0, err
	}
	if err := e.target.SyncModels(ctx, resp.Data); err != nil {
		return 0, fmt.Errorf("sync models: %w", err)
	}
	if err := e.target.ReplaceModelRatios(ctx, resp.Data); err != nil {
		return 0, fmt.Errorf("replace model ratios: %w", err)
	}
	return resp.Version, nil
}
