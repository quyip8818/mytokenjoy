package notification

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// EmailChannel sends notifications via SMTP.
type EmailChannel struct {
	host     string
	port     int
	user     string
	pass     string
	from     string
	resolver *RecipientResolver
	logger   *slog.Logger
}

type EmailConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func NewEmailChannel(cfg EmailConfig, resolver *RecipientResolver, logger *slog.Logger) *EmailChannel {
	return &EmailChannel{
		host:     strings.TrimSpace(cfg.Host),
		port:     cfg.Port,
		user:     cfg.User,
		pass:     cfg.Pass,
		from:     cfg.From,
		resolver: resolver,
		logger:   logger,
	}
}

func (c *EmailChannel) Name() string { return domainnotification.ChannelEmail }

func (c *EmailChannel) IsConfigured() bool {
	return c.host != "" && c.from != ""
}

func (c *EmailChannel) Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error {
	// Resolve memberID → email address
	info := c.resolver.Resolve(ctx, recipientID)
	to := info.Email
	if to == "" {
		c.logger.Debug("email channel: no email for recipient, skipping",
			"recipient", recipientID)
		return nil
	}

	body := buildEmailBody(msg)

	mime := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n%s",
		c.from, to, msg.Title, body)

	addr := fmt.Sprintf("%s:%d", c.host, c.port)

	var auth smtp.Auth
	if c.user != "" {
		auth = smtp.PlainAuth("", c.user, c.pass, c.host)
	}

	if err := smtp.SendMail(addr, auth, c.from, []string{to}, []byte(mime)); err != nil {
		return fmt.Errorf("email send to %s: %w", to, err)
	}

	c.logger.Debug("email sent", "to", to, "subject", msg.Title)
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
