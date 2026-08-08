package app

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/store"
)

const (
	settingPlatformChannelID = "platform_channel_id"
	platformChannelName      = "tokenjoy-upstream"
)

// ensurePlatformChannel creates or verifies the tokenjoy-upstream NewAPI channel
// that routes requests to SaaS. Idempotent — skips if channel ID already persisted.
// ponytail: best-effort on startup. If adminport is unavailable, channel
// will be created on next restart. Gateway still works (SaaS precheck is the real gate).
func ensurePlatformChannel(ctx context.Context, cfg config.Config, st store.Store, port adminport.Port, logger *slog.Logger) {
	if port == nil {
		return
	}
	settings := st.SystemSettings()

	// Already configured?
	existing, _ := settings.Get(ctx, settingPlatformChannelID)
	if existing != "" {
		return
	}

	// Need the platform key from setup.
	platformKey, _ := settings.Get(ctx, "saas_platform_key")
	if platformKey == "" {
		// Not yet set up (or SaaS mode) — skip.
		return
	}

	saasURL := cfg.SaasPlatformURL
	if saasURL == "" {
		logger.Warn("ensurePlatformChannel: SAAS_PLATFORM_URL not set, skipping")
		return
	}

	// Create the channel pointing to SaaS /v1/*.
	// ponytail: group="" so all company tokens (any group) can access platform models.
	result, err := port.UpsertChannel(ctx, adminport.UpsertChannelInput{
		Type:    1, // OpenAI-compatible
		Name:    platformChannelName,
		Key:     platformKey,
		Status:  1, // enabled
		BaseURL: saasURL + "/v1",
	})
	if err != nil {
		logger.Warn("ensurePlatformChannel: failed to create channel", "error", err)
		return
	}

	// Persist channel ID for Ingest to identify platform channel requests.
	_ = settings.Set(ctx, settingPlatformChannelID, strconv.Itoa(result.ID))
	logger.Info("ensurePlatformChannel: created", "channelId", result.ID, "name", platformChannelName)
}
