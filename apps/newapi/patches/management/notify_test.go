package management

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func resetForTest() {
	startOnce = sync.Once{}
	notifyQueue = nil
	httpClient = nil
	webhookURL = ""
	webhookSecret = ""
	droppedTotal.Store(0)
}

func droppedCount() uint64 {
	return droppedTotal.Load()
}

func formatNotifyBody(logID int64) string {
	return fmt.Sprintf(`{"log_id":%d}`, logID)
}

func TestEnqueueNotifyNoURL(t *testing.T) {
	resetForTest()
	t.Setenv(envWebhookURL, "")
	t.Setenv(envWebhookSecret, "")

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	EnqueueNotify(42)
	time.Sleep(50 * time.Millisecond)

	if calls.Load() != 0 {
		t.Fatalf("expected no HTTP calls when URL unset, got %d", calls.Load())
	}
}

func TestEnqueueNotifySuccess(t *testing.T) {
	resetForTest()

	var calls atomic.Int32
	var gotSecret string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		gotSecret = r.Header.Get("X-Webhook-Secret")
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv(envWebhookURL, server.URL)
	t.Setenv(envWebhookSecret, "test-secret")

	EnqueueNotify(123)
	waitUntil(t, func() bool { return calls.Load() == 1 })

	if gotSecret != "test-secret" {
		t.Fatalf("secret = %q, want test-secret", gotSecret)
	}
	if gotBody != formatNotifyBody(123) {
		t.Fatalf("body = %s, want %s", gotBody, formatNotifyBody(123))
	}
}

func TestEnqueueNotifyRetries503ThenSuccess(t *testing.T) {
	resetForTest()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv(envWebhookURL, server.URL)
	t.Setenv(envWebhookSecret, "")

	EnqueueNotify(7)
	waitUntil(t, func() bool { return calls.Load() >= 2 })

	if calls.Load() < 2 {
		t.Fatalf("expected at least 2 calls, got %d", calls.Load())
	}
}

func TestEnqueueNotifyNoRetryOn401(t *testing.T) {
	resetForTest()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	t.Setenv(envWebhookURL, server.URL)
	t.Setenv(envWebhookSecret, "")

	EnqueueNotify(9)
	time.Sleep(400 * time.Millisecond)

	if calls.Load() != 1 {
		t.Fatalf("expected exactly 1 call on 401, got %d", calls.Load())
	}
}

func TestEnqueueNotifyRetries500(t *testing.T) {
	resetForTest()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv(envWebhookURL, server.URL)
	t.Setenv(envWebhookSecret, "")

	EnqueueNotify(11)
	waitUntil(t, func() bool { return calls.Load() >= notifyMaxAttempts })

	if calls.Load() != int32(notifyMaxAttempts) {
		t.Fatalf("expected %d attempts, got %d", notifyMaxAttempts, calls.Load())
	}
}

func TestEnqueueNotifyDropsWhenQueueFull(t *testing.T) {
	resetForTest()

	block := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	defer close(block)

	t.Setenv(envWebhookURL, server.URL)
	os.Unsetenv(envWebhookSecret)

	notifyQueue = make(chan int64, 1)
	webhookURL = server.URL
	httpClient = &http.Client{Timeout: httpTimeout}
	go notifyWorker()

	notifyQueue <- 1
	enqueueNotifyDirect(2)

	if droppedCount() != 1 {
		t.Fatalf("expected 1 dropped notify, got %d", droppedCount())
	}
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
