package notification

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/resend/resend-go/v3"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// EmailChannel sends notifications via Resend.
type EmailChannel struct {
	client   *resend.Client
	from     string
	resolver *RecipientResolver
	logger   *slog.Logger
}

type EmailConfig struct {
	APIKey string
	From   string
}

func NewEmailChannel(cfg EmailConfig, resolver *RecipientResolver, logger *slog.Logger) *EmailChannel {
	var client *resend.Client
	if strings.TrimSpace(cfg.APIKey) != "" {
		client = resend.NewClient(cfg.APIKey)
	}
	return &EmailChannel{
		client:   client,
		from:     strings.TrimSpace(cfg.From),
		resolver: resolver,
		logger:   logger,
	}
}

func (c *EmailChannel) Name() string { return domainnotification.ChannelEmail }

func (c *EmailChannel) IsConfigured() bool {
	return c.client != nil && c.from != ""
}

func (c *EmailChannel) Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error {
	info := c.resolver.Resolve(ctx, recipientID)
	to := info.Email
	if to == "" {
		c.logger.Debug("email channel: no email for recipient, skipping",
			"recipient", recipientID)
		return nil
	}
	return c.sendToAddress(to, msg)
}

// SendDirect delivers an email directly to the given address without recipient resolution.
func (c *EmailChannel) SendDirect(ctx context.Context, address string, msg domainnotification.RenderedMessage) error {
	return c.sendToAddress(address, msg)
}

func (c *EmailChannel) sendToAddress(to string, msg domainnotification.RenderedMessage) error {
	html := buildEmailBody(msg)

	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: msg.Title,
		Html:    html,
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send to %s: %w", to, err)
	}

	c.logger.Debug("email sent via resend", "to", to, "subject", msg.Title)
	return nil
}

func buildEmailBody(msg domainnotification.RenderedMessage) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; padding: 20px; color: #333;">
  <div style="max-width: 600px; margin: 0 auto; background: #fff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 24px;">
    <h2 style="margin: 0 0 12px 0; font-size: 18px; color: #111;">%s</h2>
    <p style="margin: 0 0 16px 0; font-size: 14px; line-height: 1.6; color: #555;">%s</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 16px 0;">
    <p style="font-size: 12px; color: #999; margin: 0;">此邮件由 TokenJoy 通知系统自动发送</p>
  </div>
</body>
</html>`, msg.Title, msg.Body)
}

var _ Channel = (*EmailChannel)(nil)
var _ DirectSender = (*EmailChannel)(nil)
