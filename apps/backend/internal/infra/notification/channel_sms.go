package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// SMSChannel sends notifications via Twilio SMS API.
type SMSChannel struct {
	accountSID string
	authToken  string
	fromNumber string
	resolver   *RecipientResolver
	httpClient *http.Client
	logger     *slog.Logger
}

type SMSConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

func NewSMSChannel(cfg SMSConfig, resolver *RecipientResolver, logger *slog.Logger) *SMSChannel {
	return &SMSChannel{
		accountSID: strings.TrimSpace(cfg.AccountSID),
		authToken:  strings.TrimSpace(cfg.AuthToken),
		fromNumber: strings.TrimSpace(cfg.FromNumber),
		resolver:   resolver,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		logger:     logger,
	}
}

func (c *SMSChannel) Name() string { return domainnotification.ChannelSMS }

func (c *SMSChannel) IsConfigured() bool {
	return c.accountSID != "" && c.authToken != "" && c.fromNumber != ""
}

func (c *SMSChannel) Send(ctx context.Context, recipientID string, msg domainnotification.RenderedMessage) error {
	// Resolve memberID → phone number
	info := c.resolver.Resolve(ctx, recipientID)
	to := info.Phone
	if to == "" {
		c.logger.Debug("sms channel: no phone for recipient, skipping",
			"recipient", recipientID)
		return nil
	}

	body := msg.Title
	if msg.Body != "" {
		body = fmt.Sprintf("%s - %s", msg.Title, msg.Body)
	}
	// Truncate to SMS-friendly length
	if len(body) > 160 {
		body = body[:157] + "..."
	}

	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSID)

	form := url.Values{}
	form.Set("To", to)
	form.Set("From", c.fromNumber)
	form.Set("Body", body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twilioURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("sms new request: %w", err)
	}
	req.SetBasicAuth(c.accountSID, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sms send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("twilio error %d: %s", resp.StatusCode, errResp.Message)
	}

	c.logger.Debug("sms sent", "to", to, "body_len", len(body))
	return nil
}

var _ Channel = (*SMSChannel)(nil)
