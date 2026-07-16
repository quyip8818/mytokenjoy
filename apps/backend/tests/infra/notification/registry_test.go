package notification_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/infra/notification"
)

type mockChannel struct {
	name       string
	configured bool
	sent       []string
}

func (c *mockChannel) Name() string       { return c.name }
func (c *mockChannel) IsConfigured() bool  { return c.configured }
func (c *mockChannel) Send(_ context.Context, recipientID string, _ domainnotification.RenderedMessage) error {
	c.sent = append(c.sent, recipientID)
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRegistryRegisterAndGet(t *testing.T) {
	t.Parallel()
	r := notification.NewRegistry(testLogger())

	ch := &mockChannel{name: "test_ch", configured: true}
	r.Register(ch)

	got, ok := r.Get("test_ch")
	if !ok {
		t.Fatal("expected channel to be found")
	}
	if got.Name() != "test_ch" {
		t.Fatalf("expected name test_ch, got %s", got.Name())
	}
}

func TestRegistryConfigured(t *testing.T) {
	t.Parallel()
	r := notification.NewRegistry(testLogger())

	r.Register(&mockChannel{name: "a", configured: true})
	r.Register(&mockChannel{name: "b", configured: false})
	r.Register(&mockChannel{name: "c", configured: true})

	configured := r.Configured()
	if len(configured) != 2 {
		t.Fatalf("expected 2 configured channels, got %d", len(configured))
	}

	names := r.ConfiguredNames()
	hasA, hasC := false, false
	for _, n := range names {
		if n == "a" {
			hasA = true
		}
		if n == "c" {
			hasC = true
		}
	}
	if !hasA || !hasC {
		t.Fatalf("expected a and c in configured names, got %v", names)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	t.Parallel()
	r := notification.NewRegistry(testLogger())

	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected channel not found")
	}
}
