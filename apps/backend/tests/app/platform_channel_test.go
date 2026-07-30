//go:build testhook

package app_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil"
)

// TestEnsurePlatformChannelCreatesOnStartup verifies that a Local-mode app
// with saas_platform_key set creates tokenjoy-upstream channel and persists its ID.
func TestEnsurePlatformChannelCreatesOnStartup(t *testing.T) {
	t.Parallel()

	// Create app with Local config + platform key pre-set.
	mock := testutil.DefaultStubAdminClient()
	application := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		cfg.SupportSaas = false
		cfg.SaasPlatformURL = "https://platform.tokenjoy.test"
		cfg.PlatformSharedNewAPIGroup = "platform_shared"
	}, app.WithoutWorker(), app.WithAdminPort(mock))

	ctx := context.Background()

	// Simulate: setup has stored the platform key (normally done by setup_server).
	_ = application.Store.SystemSettings().Set(ctx, "saas_platform_key", "sk-test-platform-key")
	// Clear any existing channel_id to test fresh creation.
	_ = application.Store.SystemSettings().Set(ctx, "platform_channel_id", "")

	application.Close()

	// Recreate the app — ensurePlatformChannel runs on startup.
	application2 := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		cfg.SupportSaas = false
		cfg.SaasPlatformURL = "https://platform.tokenjoy.test"
		cfg.PlatformSharedNewAPIGroup = "platform_shared"
	}, app.WithoutWorker(), app.WithAdminPort(mock))
	defer application2.Close()

	// Verify platform_channel_id was set.
	channelID, _ := application2.Store.SystemSettings().Get(ctx, "platform_channel_id")
	if channelID == "" {
		t.Fatal("expected platform_channel_id to be set after startup")
	}

	// Verify UpsertChannel was called on the mock.
	if mock.UpsertChannelCalls == 0 {
		t.Fatal("expected UpsertChannel to be called")
	}
}

// TestEnsurePlatformChannelSkipsInSaaSMode verifies that SaaS mode
// does NOT create the platform channel (not needed for SaaS).
func TestEnsurePlatformChannelSkipsInSaaSMode(t *testing.T) {
	t.Parallel()

	mock := testutil.DefaultStubAdminClient()
	application := testutil.NewTestAppWithOptions(t, func(cfg *config.Config) {
		cfg.SupportSaas = true // SaaS mode
		cfg.SaasPlatformURL = "https://platform.tokenjoy.test"
	}, app.WithoutWorker(), app.WithAdminPort(mock))
	defer application.Close()

	ctx := context.Background()

	// Even if platform key is set, SaaS mode should not create a channel.
	_ = application.Store.SystemSettings().Set(ctx, "saas_platform_key", "sk-should-not-trigger")

	channelID, _ := application.Store.SystemSettings().Get(ctx, "platform_channel_id")
	if channelID != "" {
		t.Fatalf("expected no platform_channel_id in SaaS mode, got %q", channelID)
	}
}
