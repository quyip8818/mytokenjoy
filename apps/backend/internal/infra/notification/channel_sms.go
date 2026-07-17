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
	client       *dysmsapi.Client
	signName     string
	templateCode string
	resolver     *RecipientResolver
	logger       *slog.Logger
}

// SMSConfig holds Alibaba Cloud SMS settings.
type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	Endpoint        string // defaults to dysmsapi.aliyuncs.com
}

// NewSMSChannel creates an Aliyun SMS channel.
// If credentials are incomplete the channel reports IsConfigured() == false
// and the registry will skip it during dispatch.
func NewSMSChannel(cfg SMSConfig, resolver *RecipientResolver, logger *slog.Logger) *SMSChannel {
	ch := &SMSChannel{
		signName:     strings.TrimSpace(cfg.SignName),
		templateCode: strings.TrimSpace(cfg.TemplateCode),
		resolver:     resolver,
		logger:       logger,
	}

	keyID := strings.TrimSpace(cfg.AccessKeyID)
	keySecret := strings.TrimSpace(cfg.AccessKeySecret)
	if keyID == "" || keySecret == "" || ch.signName == "" || ch.templateCode == "" {
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

func (c *SMSChannel) Send(ctx context.Context, recipientID string, msg domainnotification.RenderedMessage) error {
	phone := c.resolver.Resolve(ctx, uuid.MustParse(recipientID)).Phone
	if phone == "" {
		c.logger.Debug("sms: no phone for recipient, skipping", "recipient", recipientID)
		return nil
	}

	content := buildSMSContent(msg)
	paramJSON, _ := json.Marshal(map[string]string{"content": content})

	resp, err := c.client.SendSms(&dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(c.signName),
		TemplateCode:  tea.String(c.templateCode),
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

	c.logger.Debug("sms sent", "to", phone, "bizId", tea.StringValue(resp.Body.BizId))
	return nil
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

var _ Channel = (*SMSChannel)(nil)
