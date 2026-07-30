//go:build testhook

package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

// TestIngestPlatformChannelConsumesLots verifies that platform channel requests
// (ChannelID == 0 or matches platform_channel_id) consume lots normally.
func TestIngestPlatformChannelConsumesLots(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	// Get wallet before.
	coBefore, err := fix.Store.Company().GetByID(testutil.Ctx(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletBefore := coBefore.WalletRemainQuota

	// ChannelID = 0 → legacy, treated as platform channel.
	raw := testutil.DefaultConsumeLog(7001, 99)
	raw.ChannelID = 0
	testutil.SeedConsumeLog(t, fix.Store, raw)

	if err := fix.Ingest.IngestByLogID(testutil.Ctx(), 7001, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	// Verify ledger entry was created.
	ingested, err := testutil.HasLedgerLogID(fix.Store, 7001)
	if err != nil || !ingested {
		t.Fatalf("expected log 7001 in ledger, err=%v ingested=%v", err, ingested)
	}

	// Verify wallet was decremented (lot consumption happened).
	coAfter, err := fix.Store.Company().GetByID(testutil.Ctx(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if coAfter.WalletRemainQuota >= walletBefore {
		t.Fatalf("expected wallet to decrease for platform channel: before=%d after=%d",
			walletBefore, coAfter.WalletRemainQuota)
	}
}

// TestIngestNonPlatformChannelSkipsLots verifies that non-platform channel requests
// skip lot consumption but still record budget_consumed and ledger.
func TestIngestNonPlatformChannelSkipsLots(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	// Set platform_channel_id to some value (e.g. 42) so that channel 999 is non-platform.
	if err := fix.Store.SystemSettings().Set(testutil.Ctx(), "platform_channel_id", "42"); err != nil {
		t.Fatal(err)
	}

	// Get wallet before ingest.
	coBefore, err := fix.Store.Company().GetByID(testutil.Ctx(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletBefore := coBefore.WalletRemainQuota

	// Ingest with a non-platform channel ID.
	raw := store.RawConsumeLog{
		ID: 7002, TokenID: 99, ChannelID: 999,
		Quota: 100000, ModelName: "gpt-4o", CreatedAt: 1781866800,
	}
	testutil.SeedConsumeLog(t, fix.Store, raw)

	if err := fix.Ingest.IngestByLogID(testutil.Ctx(), 7002, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	// Verify ledger entry was created (budget tracking works).
	ingested, err := testutil.HasLedgerLogID(fix.Store, 7002)
	if err != nil || !ingested {
		t.Fatalf("expected log 7002 in ledger, err=%v ingested=%v", err, ingested)
	}

	// Verify wallet was NOT decremented (lot consumption skipped).
	coAfter, err := fix.Store.Company().GetByID(testutil.Ctx(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if coAfter.WalletRemainQuota != walletBefore {
		t.Fatalf("expected wallet unchanged for non-platform channel: before=%d after=%d",
			walletBefore, coAfter.WalletRemainQuota)
	}
}

// TestIngestPlatformChannelBumpsWalletLotsVersion verifies that platform channel
// ingest bumps the catalog.wallet_lots_version for sync clients to detect changes.
func TestIngestPlatformChannelBumpsWalletLotsVersion(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	ctx := testutil.Ctx()

	// Get version before.
	vBefore, _ := fix.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")

	// Ingest on platform channel (ChannelID=0 → platform).
	raw := testutil.DefaultConsumeLog(7003, 99)
	raw.ChannelID = 0
	testutil.SeedConsumeLog(t, fix.Store, raw)

	if err := fix.Ingest.IngestByLogID(ctx, 7003, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	// Version should have incremented.
	vAfter, _ := fix.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")
	if vAfter == vBefore {
		t.Fatalf("expected wallet_lots_version to bump after platform ingest: before=%q after=%q", vBefore, vAfter)
	}
}

// TestIngestNonPlatformChannelDoesNotBumpVersion verifies that non-platform channel
// ingest does NOT bump wallet_lots_version (no lot change occurred).
func TestIngestNonPlatformChannelDoesNotBumpVersion(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	ctx := testutil.Ctx()

	// Set platform_channel_id so channel 888 is non-platform.
	_ = fix.Store.SystemSettings().Set(ctx, "platform_channel_id", "42")

	// Get version before.
	vBefore, _ := fix.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")

	// Ingest on non-platform channel.
	raw := store.RawConsumeLog{
		ID: 7004, TokenID: 99, ChannelID: 888,
		Quota: 100000, ModelName: "gpt-4o", CreatedAt: 1781866800,
	}
	testutil.SeedConsumeLog(t, fix.Store, raw)

	if err := fix.Ingest.IngestByLogID(ctx, 7004, types.SourceWebhook); err != nil {
		t.Fatal(err)
	}

	// Version should NOT have changed.
	vAfter, _ := fix.Store.SystemSettings().Get(ctx, "catalog.wallet_lots_version")
	if vAfter != vBefore {
		t.Fatalf("expected wallet_lots_version unchanged for non-platform ingest: before=%q after=%q", vBefore, vAfter)
	}
}
