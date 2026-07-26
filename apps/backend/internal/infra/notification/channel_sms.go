package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// SMSChannel delivers notifications via Alibaba Cloud SMS (阿里云短信服务).
type SMSChannel struct {
	client    *dysmsapi.Client
	signName  string
	templates map[string]string // eventType → templateCode
	resolver  *RecipientResolver
	logger    *slog.Logger
}

// SMSConfig holds Alibaba Cloud SMS settings.
type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string            // default (verification code)
	Templates       map[string]string // additional eventType → templateCode mappings
	Endpoint        string            // defaults to dysmsapi.aliyuncs.com
}

// NewSMSChannel creates an Aliyun SMS channel.
// If credentials are incomplete the channel reports IsConfigured() == false
// and the registry will skip it during dispatch.
func NewSMSChannel(cfg SMSConfig, resolver *RecipientResolver, logger *slog.Logger) *SMSChannel {
	// Build template map: default template handles "verification_code" and fallback.
	templates := map[string]string{
		"": strings.TrimSpace(cfg.TemplateCode), // fallback/default
	}
	for event, code := range cfg.Templates {
		templates[event] = strings.TrimSpace(code)
	}

	ch := &SMSChannel{
		signName:  strings.TrimSpace(cfg.SignName),
		templates: templates,
		resolver:  resolver,
		logger:    logger,
	}

	keyID := strings.TrimSpace(cfg.AccessKeyID)
	keySecret := strings.TrimSpace(cfg.AccessKeySecret)
	if keyID == "" || keySecret == "" || ch.signName == "" || templates[""] == "" {
		return ch
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = "dysmsapi.aliyuncs.com"
	}

	client, err := dysmsapi.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(keyID),
		AccessKeySecret: tea.String(keySecret),
		Endpoint:        tea.String(endpoint),
	})
	if err != nil {
		logger.Error("failed to create aliyun sms client", "error", err)
		return ch
	}
	ch.client = client
	return ch
}

func (c *SMSChannel) Name() string { return domainnotification.ChannelSMS }

func (c *SMSChannel) IsConfigured() bool {
	return c.client != nil
}

func (c *SMSChannel) Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error {
	phone := c.resolver.Resolve(ctx, recipientID).Phone
	if phone == "" {
		c.logger.Debug("sms: no phone for recipient, skipping", "recipient", recipientID)
		return nil
	}
	return c.sendToPhone(phone, msg)
}

// buildSMSContent formats the notification into a short text for the template variable.
func buildSMSContent(msg domainnotification.RenderedMessage) string {
	s := msg.Title
	if msg.Body != "" {
		s += " - " + msg.Body
	}
	if len(s) > 100 {
		s = s[:97] + "..."
	}
	return s
}

// SendDirect delivers an SMS directly to the given phone number without recipient resolution.
func (c *SMSChannel) SendDirect(ctx context.Context, address string, msg domainnotification.RenderedMessage) error {
	if c.client == nil {
		c.logger.Info("sms: [DEV] direct send", "phone", address, "title", msg.Title, "body", msg.Body)
		return nil
	}
	return c.sendToPhone(address, msg)
}

func (c *SMSChannel) sendToPhone(phone string, msg domainnotification.RenderedMessage) error {
	// Strip +86 prefix for Aliyun API.
	phoneNum := phone
	if strings.HasPrefix(phone, "+86") {
		phoneNum = strings.TrimPrefix(phone, "+86")
	}

	// Resolve template code by eventType (same pattern as email channel).
	eventType, _ := msg.Payload["eventType"].(string)
	tmpl := c.templates[eventType]
	if tmpl == "" {
		tmpl = c.templates[""] // fallback to default
	}
	if tmpl == "" {
		return fmt.Errorf("sms: no template for event %q", eventType)
	}

	// Build template params: all string payload fields except eventType.
	params := make(map[string]string)
	for k, v := range msg.Payload {
		if k == "eventType" {
			continue
		}
		if s, ok := v.(string); ok && s != "" {
			params[k] = s
		}
	}
	if len(params) == 0 {
		params["content"] = buildSMSContent(msg)
	}
	paramJSON, _ := json.Marshal(params)

	resp, err := c.client.SendSms(&dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNum),
		SignName:      tea.String(c.signName),
		TemplateCode:  tea.String(tmpl),
		TemplateParam: tea.String(string(paramJSON)),
	})
	if err != nil {
		return fmt.Errorf("aliyun sms: %w", err)
	}
	if resp.Body == nil {
		return fmt.Errorf("aliyun sms: nil response body")
	}
	if code := tea.StringValue(resp.Body.Code); code != "OK" {
		return fmt.Errorf("aliyun sms rejected: %s - %s", code, tea.StringValue(resp.Body.Message))
	}
	c.logger.Debug("sms direct sent", "to", phoneNum, "bizId", tea.StringValue(resp.Body.BizId))
	return nil
}

var _ Channel = (*SMSChannel)(nil)
var _ DirectSender = (*SMSChannel)(nil)
