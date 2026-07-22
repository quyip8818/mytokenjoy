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

// EmailChannel sends notifications via Resend templates.
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
	templateID := resolveTemplateID(msg)
	if templateID == "" {
		return fmt.Errorf("resend: no template configured for event %v", msg.Payload["eventType"])
	}

	vars := make(map[string]any, len(msg.Payload))
	for k, v := range msg.Payload {
		vars[k] = v
	}

	params := &resend.SendEmailRequest{
		From:    c.from,
		To:      []string{to},
		Subject: msg.Title,
		Template: &resend.EmailTemplate{
			Id:        templateID,
			Variables: vars,
		},
	}

	_, err := c.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resend send to %s: %w", to, err)
	}

	c.logger.Debug("email sent via resend", "to", to, "subject", msg.Title, "template", templateID)
	return nil
}

// resolveTemplateID maps eventType to a hardcoded Resend template alias (kebab-case).
func resolveTemplateID(msg domainnotification.RenderedMessage) string {
	eventType, _ := msg.Payload["eventType"].(string)
	switch eventType {
	case domainnotification.EventBudgetAlertReached:
		return "budget-alert"
	case domainnotification.EventOverrunBlocked, domainnotification.EventOverdraftExpanded:
		return "overrun-blocked"
	case domainnotification.EventSyncThresholdExceeded:
		return "sync-threshold-exceeded"
	case "verification_code":
		return "verification-code"
	case "company_invite":
		return "company-invite"
	default:
		return ""
	}
}

var _ Channel = (*EmailChannel)(nil)
var _ DirectSender = (*EmailChannel)(nil)
