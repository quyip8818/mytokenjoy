package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// WebhookChannel sends notifications to a configured webhook URL.
type WebhookChannel struct {
	url        string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewWebhookChannel(url string, logger *slog.Logger) *WebhookChannel {
	return &WebhookChannel{
		url:        strings.TrimSpace(url),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

func (c *WebhookChannel) Name() string { return domainnotification.ChannelWebhook }

func (c *WebhookChannel) IsConfigured() bool {
	return c.url != ""
}

func (c *WebhookChannel) Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error {
	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("webhook marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("webhook new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Notification-Title", msg.Title)
	req.Header.Set("X-Notification-Recipient", recipientID.String())

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook send: %w", err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)

	if res.StatusCode >= 400 {
		return fmt.Errorf("webhook response status %d", res.StatusCode)
	}
	return nil
}

var _ Channel = (*WebhookChannel)(nil)
