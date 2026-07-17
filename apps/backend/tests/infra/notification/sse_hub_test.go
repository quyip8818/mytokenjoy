package notification_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/infra/notification"
)

var (
	sseUser1 = uuid.MustParse("00000000-0000-7000-0000-000000000001")
	sseUser2 = uuid.MustParse("00000000-0000-7000-0000-000000000002")
)

func TestSSEHubSubscribeAndPublish(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	ch, unsub := hub.Subscribe(sseUser1)
	defer unsub()

	event := notification.SSEEvent{ID: "n1", EventType: "test", Title: "Hello", Body: "World"}
	hub.Publish(sseUser1, event)

	select {
	case got := <-ch:
		if got.ID != "n1" || got.Title != "Hello" {
			t.Fatalf("unexpected event: %+v", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestSSEHubPublishToWrongUser(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	ch, unsub := hub.Subscribe(sseUser1)
	defer unsub()

	hub.Publish(sseUser2, notification.SSEEvent{ID: "n2", Title: "Not for you"})

	select {
	case <-ch:
		t.Fatal("should not receive event for different user")
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

func TestSSEHubUnsubscribe(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	_, unsub := hub.Subscribe(sseUser1)
	if hub.ActiveSubscribers(sseUser1) != 1 {
		t.Fatal("expected 1 subscriber")
	}

	unsub()
	if hub.ActiveSubscribers(sseUser1) != 0 {
		t.Fatal("expected 0 subscribers after unsubscribe")
	}
}

func TestSSEHubMultipleSubscribers(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	ch1, unsub1 := hub.Subscribe(sseUser1)
	ch2, unsub2 := hub.Subscribe(sseUser1)
	defer unsub1()
	defer unsub2()

	hub.Publish(sseUser1, notification.SSEEvent{ID: "n3", Title: "Broadcast"})

	select {
	case got := <-ch1:
		if got.ID != "n3" {
			t.Fatalf("ch1 unexpected: %+v", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch1 timeout")
	}

	select {
	case got := <-ch2:
		if got.ID != "n3" {
			t.Fatalf("ch2 unexpected: %+v", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("ch2 timeout")
	}
}

func TestSSEHubDropsSlowSubscriber(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	// Subscribe but never consume
	_, unsub := hub.Subscribe(sseUser1)
	defer unsub()

	// Fill the buffer (16) and one more — should not block
	for i := 0; i < 20; i++ {
		hub.Publish(sseUser1, notification.SSEEvent{ID: "overflow", Title: "spam"})
	}
	// If we reach here without hanging, the test passes
}
