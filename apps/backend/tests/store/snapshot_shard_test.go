package store_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestSnapshotToShardsRoundTrip(t *testing.T) {
	cfg := testutil.TestConfig()
	original := seed.Load(cfg)

	shards, err := store.SnapshotToShards(original)
	if err != nil {
		t.Fatal(err)
	}
	if len(shards) != len(store.AllShardIDs()) {
		t.Fatalf("expected %d shards, got %d", len(store.AllShardIDs()), len(shards))
	}

	restored, err := store.ShardsToSnapshot(shards)
	if err != nil {
		t.Fatal(err)
	}
	if !store.SnapshotShardFieldsEqual(original, restored) {
		t.Fatal("shard round-trip changed persisted fields")
	}
}

func TestShardsToSnapshotRequiresAllShards(t *testing.T) {
	cfg := testutil.TestConfig()
	original := seed.Load(cfg)
	shards, err := store.SnapshotToShards(original)
	if err != nil {
		t.Fatal(err)
	}
	delete(shards, store.ShardOrg)
	if _, err := store.ShardsToSnapshot(shards); err == nil {
		t.Fatal("expected error for missing shard")
	}
}
