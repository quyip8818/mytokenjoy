package store_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/store"
)

func TestValidPGNotifyChannel(t *testing.T) {
	t.Parallel()
	if err := store.ValidPGNotifyChannel(store.IngestPendingChannel); err != nil {
		t.Fatal(err)
	}
	if err := store.ValidPGNotifyChannel("bad-channel"); err == nil {
		t.Fatal("expected error for hyphenated channel")
	}
	if err := store.ValidPGNotifyChannel(""); err == nil {
		t.Fatal("expected error for empty channel")
	}
	if err := store.ValidPGNotifyChannel("1bad"); err == nil {
		t.Fatal("expected error for leading digit")
	}
}
