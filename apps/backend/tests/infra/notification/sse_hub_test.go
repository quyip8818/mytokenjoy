package notification_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/infra/notification"
)

func TestSSEHubSubscribeAndPublish(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	ch, unsub := hub.Subscribe("user-1")
	defer unsub()

	event := notification.SSEEvent{ID: "n1", EventType: "test", Title: "Hello", Body: "World"}
	hub.Publish("user-1", event)

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

	ch, unsub := hub.Subscribe("user-1")
	defer unsub()

	hub.Publish("user-2", notification.SSEEvent{ID: "n2", Title: "Not for you"})

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

	_, unsub := hub.Subscribe("user-1")
	if hub.ActiveSubscribers("user-1") != 1 {
		t.Fatal("expected 1 subscriber")
	}

	unsub()
	if hub.ActiveSubscribers("user-1") != 0 {
		t.Fatal("expected 0 subscribers after unsubscribe")
	}
}

func TestSSEHubMultipleSubscribers(t *testing.T) {
	t.Parallel()
	hub := notification.NewSSEHub()

	ch1, unsub1 := hub.Subscribe("user-1")
	ch2, unsub2 := hub.Subscribe("user-1")
	defer unsub1()
	defer unsub2()

	hub.Publish("user-1", notification.SSEEvent{ID: "n3", Title: "Broadcast"})

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
	_, unsub := hub.Subscribe("user-1")
	defer unsub()

	// Fill the buffer (16) and one more — should not block
	for i := 0; i < 20; i++ {
		hub.Publish("user-1", notification.SSEEvent{ID: "overflow", Title: "spam"})
	}
	// If we reach here without hanging, the test passes
}
