package management

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	envWebhookURL     = "MANAGEMENT_WEBHOOK_URL"
	envWebhookSecret  = "MANAGEMENT_WEBHOOK_SECRET"
	notifyQueueSize   = 512
	notifyWorkerCount = 4
	httpTimeout       = 10 * time.Second
	notifyMaxAttempts = 3
)

var notifyRetryBackoff = []time.Duration{
	200 * time.Millisecond,
	time.Second,
	3 * time.Second,
}

type notifyPayload struct {
	LogID int64 `json:"log_id"`
}

var (
	startOnce   sync.Once
	notifyQueue chan int64
	httpClient  *http.Client
	webhookURL  string
	webhookSecret string
	droppedTotal atomic.Uint64
)

func EnqueueNotify(logID int64) {
	if logID <= 0 {
		return
	}
	startOnce.Do(startWorkers)
	if webhookURL == "" {
		return
	}
	enqueueNotifyDirect(logID)
}

func enqueueNotifyDirect(logID int64) {
	select {
	case notifyQueue <- logID:
	default:
		droppedTotal.Add(1)
		log.Printf("management notify: queue full, dropped log_id=%d", logID)
	}
}

func startWorkers() {
	webhookURL = os.Getenv(envWebhookURL)
	webhookSecret = os.Getenv(envWebhookSecret)
	if webhookURL == "" {
		return
	}
	notifyQueue = make(chan int64, notifyQueueSize)
	httpClient = &http.Client{Timeout: httpTimeout}
	for i := 0; i < notifyWorkerCount; i++ {
		go notifyWorker()
	}
}

func notifyWorker() {
	for logID := range notifyQueue {
		deliverNotify(logID)
	}
}

func deliverNotify(logID int64) {
	body, err := json.Marshal(notifyPayload{LogID: logID})
	if err != nil {
		log.Printf("management notify: marshal log_id=%d: %v", logID, err)
		return
	}
	for attempt := 0; attempt < notifyMaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(notifyRetryBackoff[attempt-1])
		}
		status, sendErr := postOnce(body)
		if sendErr != nil {
			if attempt == notifyMaxAttempts-1 {
				log.Printf("management notify: log_id=%d failed after %d attempts: %v", logID, notifyMaxAttempts, sendErr)
			}
			continue
		}
		if status >= 200 && status < 300 {
			return
		}
		if status == http.StatusUnauthorized || status == http.StatusBadRequest {
			log.Printf("management notify: log_id=%d non-retryable status %d", logID, status)
			return
		}
		if status == http.StatusServiceUnavailable || status >= 500 {
			if attempt == notifyMaxAttempts-1 {
				log.Printf("management notify: log_id=%d retryable status %d exhausted", logID, status)
			}
			continue
		}
		log.Printf("management notify: log_id=%d unexpected status %d", logID, status)
		return
	}
}

func postOnce(body []byte) (int, error) {
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if webhookSecret != "" {
		req.Header.Set("X-Webhook-Secret", webhookSecret)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return res.StatusCode, nil
}
